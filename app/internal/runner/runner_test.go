package runner

import (
	"testing"

	"github.com/e1berd/drift/app/internal/workspace"
	"github.com/jackc/pgx/v5/pgtype"
)

func busyRunner(session pgtype.UUID) (*Runner, *run, *bool) {
	cancelled := false
	current := &run{queue: []workspace.RunRequest{{Text: "queued"}}, cancel: func() { cancelled = true }}
	return &Runner{runs: map[pgtype.UUID]*run{session: current}}, current, &cancelled
}

func TestInterruptKeepsTheQueue(t *testing.T) {
	session := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	r, current, cancelled := busyRunner(session)
	if !r.Interrupt("owner", session) || !*cancelled {
		t.Fatal("running answer was not cancelled")
	}
	if len(current.queue) != 1 {
		t.Errorf("queue %v", current.queue)
	}
	if r.Interrupt("owner", pgtype.UUID{Bytes: [16]byte{2}, Valid: true}) {
		t.Error("interrupted a session that is not running")
	}
}

func TestStopDropsTheQueue(t *testing.T) {
	session := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	r, current, cancelled := busyRunner(session)
	if !r.Stop("owner", session) || !*cancelled || current.queue != nil {
		t.Fatalf("cancelled %v, queue %v", *cancelled, current.queue)
	}
}
