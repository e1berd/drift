package runner

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/workspace"
	"github.com/jackc/pgx/v5"
)

const (
	stopGrace      = 5 * time.Second
	maxEventBytes  = 32 << 20
	stderrTailSize = 4 << 10
)

func (r *Runner) execute(ctx context.Context, req workspace.RunRequest) {
	log := r.log.With("session", uuidString(req.Session))
	t := newTranscript(r, req.Owner, req.Session)
	rc, err := r.q.GetRunContext(r.ctx, db.GetRunContextParams{ID: req.Session, OwnerID: req.Owner})
	if err != nil {
		log.Warn("load session for run", "err", err)
		return
	}
	if rc.Agent != "claude" {
		t.note(noteUnsupportedAgent, "")
		return
	}
	sub, err := r.q.GetCurrentSubscription(r.ctx, db.GetCurrentSubscriptionParams{OwnerID: req.Owner, Agent: rc.Agent})
	if errors.Is(err, pgx.ErrNoRows) {
		t.note(noteNoSubscription, "")
		return
	}
	if err != nil {
		log.Warn("load subscription for run", "err", err)
		t.note(noteFailed, err.Error())
		return
	}
	dir, err := r.folders.Session(rc.ProjectPath, rc.ID)
	if err != nil {
		t.note(noteFailed, err.Error())
		return
	}
	if !r.acquire(ctx) {
		t.note(noteStopped, "")
		return
	}
	defer r.release()

	cmd := exec.CommandContext(ctx, r.cli.Claude, claudeArgs(req, rc.AgentSessionID)...)
	cmd.Dir = dir
	cmd.Env = r.cli.Env(r.cli.Home(uuidString(sub.ID)))
	cmd.Stdin = strings.NewReader(req.Text)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return signalGroup(cmd.Process.Pid, syscall.SIGINT) }
	cmd.WaitDelay = stopGrace
	stderr := &tailBuffer{limit: stderrTailSize}
	cmd.Stderr = stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.note(noteFailed, err.Error())
		return
	}
	if err := cmd.Start(); err != nil {
		t.note(noteFailed, err.Error())
		return
	}

	stream := claudeStream{t: t, projectDir: cmd.Dir, agentSessionID: rc.AgentSessionID, saveSessionID: func(id string) {
		if err := r.q.SetAgentSessionID(r.ctx, db.SetAgentSessionIDParams{ID: req.Session, AgentSessionID: id}); err != nil {
			log.Warn("save agent session id", "err", err)
		}
	}}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64<<10), maxEventBytes)
	for scanner.Scan() {
		stream.handle(scanner.Bytes())
	}
	waitErr := cmd.Wait()
	_ = signalGroup(cmd.Process.Pid, syscall.SIGKILL)

	switch {
	case ctx.Err() != nil:
		t.closeTools(toolStopped)
		t.note(noteStopped, "")
	case waitErr != nil && !stream.finished:
		t.closeTools(toolError)
		t.note(noteFailed, lastLine(stderr.String(), waitErr.Error()))
	default:
		t.closeTools(toolDone)
	}
}

func claudeArgs(req workspace.RunRequest, agentSessionID string) []string {
	args := []string{"-p", "--output-format", "stream-json", "--verbose"}
	if req.Model != "" && req.Model != "default" {
		args = append(args, "--model", req.Model)
	}
	if req.Mode != "" && req.Mode != "default" {
		args = append(args, "--permission-mode", req.Mode)
	}
	if agentSessionID != "" {
		args = append(args, "--resume", agentSessionID)
	}
	return args
}

func signalGroup(pid int, signal syscall.Signal) error {
	err := syscall.Kill(-pid, signal)
	if errors.Is(err, syscall.ESRCH) {
		return os.ErrProcessDone
	}
	return err
}

func lastLine(output, fallback string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if line := strings.TrimSpace(lines[len(lines)-1]); line != "" {
		return line
	}
	return fallback
}

type tailBuffer struct {
	limit int
	data  []byte
}

func (b *tailBuffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	if extra := len(b.data) - b.limit; extra > 0 {
		b.data = b.data[extra:]
	}
	return len(p), nil
}

func (b *tailBuffer) String() string { return string(b.data) }
