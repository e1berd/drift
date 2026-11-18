// Package runner runs the agent CLI of a session in its project folder and turns the CLI's
// event stream into transcript messages. One session runs one agent process at a time;
// messages sent meanwhile wait in the session's queue and go to the agent as one prompt.
package runner

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/e1berd/drift/app/internal/agents"
	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/events"
	"github.com/e1berd/drift/app/internal/workspace"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxParallelRuns = 2

// Runner owns the running agent processes.
type Runner struct {
	ctx     context.Context
	q       *db.Queries
	hub     *events.Hub
	cli     agents.CLI
	folders workspace.Folders
	log     *slog.Logger
	askMCP  mcpCommand

	slots  chan struct{}
	mu     sync.Mutex
	runs   map[pgtype.UUID]*run
	parked map[pgtype.UUID]*parkedSession
	usage  map[string]cachedUsage
	warned map[string]float64
}

type run struct {
	queue  []workspace.RunRequest
	cancel context.CancelFunc
}

// New returns a runner whose processes live as long as ctx, and marks sessions left
// running by a previous shutdown as idle.
func New(ctx context.Context, pool *pgxpool.Pool, hub *events.Hub, cli agents.CLI, folders workspace.Folders, log *slog.Logger) (*Runner, error) {
	q := db.New(pool)
	if err := q.ResetRunningSessions(ctx); err != nil {
		return nil, err
	}
	askMCP, err := askMCPCommand()
	if err != nil {
		return nil, err
	}
	r := &Runner{
		ctx: ctx, q: q, hub: hub, cli: cli, folders: folders, log: log, askMCP: askMCP,
		slots:  make(chan struct{}, maxParallelRuns),
		runs:   make(map[pgtype.UUID]*run),
		parked: make(map[pgtype.UUID]*parkedSession),
		usage:  make(map[string]cachedUsage),
		warned: make(map[string]float64),
	}
	if err := r.restoreContinuations(); err != nil {
		return nil, err
	}
	return r, nil
}

// Enqueue starts the session's agent, or queues the request while it is busy or waits for
// a usage limit to reset.
func (r *Runner) Enqueue(req workspace.RunRequest) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if parked, ok := r.parked[req.Session]; ok {
		parked.queue = append(parked.queue, req)
		return
	}
	if current, ok := r.runs[req.Session]; ok {
		current.queue = append(current.queue, req)
		return
	}
	current := &run{queue: []workspace.RunRequest{req}}
	r.runs[req.Session] = current
	r.setStatus(req.Owner, req.Session, "running")
	go r.drain(req.Owner, req.Session, current)
}

// Interrupt cancels the agent's current answer; messages queued for the session still go to
// the agent next.
func (r *Runner) Interrupt(owner string, session pgtype.UUID) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.runs[session]
	if !ok || current.cancel == nil {
		return false
	}
	current.cancel()
	return true
}

// Stop cancels the running agent of the session and drops its queue.
func (r *Runner) Stop(owner string, session pgtype.UUID) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.runs[session]
	if !ok || current.queue == nil && current.cancel == nil {
		return false
	}
	current.queue = nil
	if current.cancel != nil {
		current.cancel()
	}
	return true
}

func (r *Runner) drain(owner string, session pgtype.UUID, current *run) {
	for {
		r.mu.Lock()
		if len(current.queue) == 0 {
			current.cancel = nil
			delete(r.runs, session)
			r.setStatus(owner, session, "idle")
			r.mu.Unlock()
			return
		}
		batch := merge(current.queue)
		current.queue = nil
		ctx, cancel := context.WithCancel(r.ctx)
		current.cancel = cancel
		r.mu.Unlock()

		r.execute(ctx, batch)
		cancel()
	}
}

func merge(queue []workspace.RunRequest) workspace.RunRequest {
	last := queue[len(queue)-1]
	texts := make([]string, 0, len(queue))
	var attachments []pgtype.UUID
	for _, req := range queue {
		if req.Text != "" {
			texts = append(texts, req.Text)
		}
		attachments = append(attachments, req.Attachments...)
	}
	last.Text = strings.Join(texts, "\n\n")
	last.Attachments = attachments
	return last
}

func (r *Runner) setStatus(owner string, session pgtype.UUID, status string) {
	s, err := r.q.SetSessionStatus(r.ctx, db.SetSessionStatusParams{ID: session, Status: status})
	if err != nil {
		r.log.Warn("set session status", "err", err, "status", status)
		return
	}
	r.hub.Publish(owner, "session.updated", workspace.SessionOf(s), "")
}

func (r *Runner) acquire(ctx context.Context) bool {
	select {
	case r.slots <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

func (r *Runner) release() { <-r.slots }
