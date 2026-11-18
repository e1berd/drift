package runner

import (
	"context"
	"encoding/json"
	"time"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/workspace"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	continueAfterReset = time.Minute
	continuePrompt     = "The usage limit has reset. Continue the task from where you stopped."
)

type parkedSession struct {
	owner   string
	message pgtype.UUID
	body    limitBody
	timer   *time.Timer
	queue   []workspace.RunRequest
}

// ScheduleContinue makes the session continue on its own once the limit of the message resets;
// messages sent meanwhile wait and go to the agent with the continuation.
func (r *Runner) ScheduleContinue(owner string, session, message pgtype.UUID) error {
	body, err := r.loadLimit(session, message)
	if err != nil {
		return err
	}
	if body.Status != limitReached || body.ResetTime == nil {
		return workspace.ErrNotContinuable
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, busy := r.runs[session]; busy {
		return workspace.ErrNotContinuable
	}
	body.Status = limitScheduled
	r.updateMessage(owner, message, body)
	r.park(owner, session, message, body)
	return nil
}

// CancelContinue drops the scheduled continuation; messages that waited for it run now.
func (r *Runner) CancelContinue(owner string, session, message pgtype.UUID) error {
	r.mu.Lock()
	parked, ok := r.parked[session]
	if !ok || parked.message != message {
		r.mu.Unlock()
		return workspace.ErrNotContinuable
	}
	parked.timer.Stop()
	delete(r.parked, session)
	r.mu.Unlock()

	parked.body.Status = limitReached
	r.updateMessage(owner, message, parked.body)
	if len(parked.queue) > 0 {
		r.Enqueue(merge(parked.queue))
	}
	return nil
}

// ContinueNow continues the session at once, without waiting for the limit to reset.
func (r *Runner) ContinueNow(owner string, session, message pgtype.UUID) error {
	body, err := r.loadLimit(session, message)
	if err != nil {
		return err
	}
	switch body.Status {
	case limitScheduled:
		r.continueParked(session, message)
	case limitReached:
		r.continueSession(owner, session, message, body, nil)
	default:
		return workspace.ErrNotContinuable
	}
	return nil
}

func (r *Runner) park(owner string, session, message pgtype.UUID, body limitBody) {
	wait := time.Until(body.ResetTime.Add(continueAfterReset))
	r.parked[session] = &parkedSession{
		owner: owner, message: message, body: body,
		timer: time.AfterFunc(max(wait, 0), func() { r.continueParked(session, message) }),
	}
}

func (r *Runner) continueParked(session, message pgtype.UUID) {
	r.mu.Lock()
	parked, ok := r.parked[session]
	if !ok || parked.message != message {
		r.mu.Unlock()
		return
	}
	parked.timer.Stop()
	delete(r.parked, session)
	r.mu.Unlock()
	r.continueSession(parked.owner, session, message, parked.body, parked.queue)
}

func (r *Runner) continueSession(owner string, session, message pgtype.UUID, body limitBody, queue []workspace.RunRequest) {
	body.Status = limitContinued
	r.updateMessage(owner, message, body)
	req := workspace.RunRequest{Owner: owner, Session: session, Text: continuePrompt, Model: body.Model, Mode: body.Mode}
	if len(queue) > 0 {
		waiting := merge(queue)
		req.Text = continuePrompt + "\n\n" + waiting.Text
		req.Attachments, req.Model, req.Mode = waiting.Attachments, waiting.Model, waiting.Mode
	}
	r.Enqueue(req)
}

func (r *Runner) loadLimit(session, message pgtype.UUID) (limitBody, error) {
	m, err := r.q.GetSessionMessage(context.WithoutCancel(r.ctx), db.GetSessionMessageParams{ID: message, SessionID: session})
	if err != nil {
		return limitBody{}, err
	}
	var body limitBody
	if m.Kind != kindLimit || json.Unmarshal(m.Body, &body) != nil {
		return limitBody{}, workspace.ErrNotContinuable
	}
	return body, nil
}

func (r *Runner) restoreContinuations() error {
	rows, err := r.q.ListScheduledContinuations(r.ctx)
	if err != nil {
		return err
	}
	for _, row := range rows {
		var body limitBody
		if json.Unmarshal(row.Body, &body) != nil || body.ResetTime == nil {
			continue
		}
		r.park(row.OwnerID, row.SessionID, row.ID, body)
	}
	return nil
}
