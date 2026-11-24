package workspace

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	schema "github.com/e1berd/drift/app/db"
	"github.com/e1berd/drift/app/internal/database"
	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/migrate"
	"github.com/e1berd/drift/app/internal/typeid"
)

const seededMessages = 300

func windowTestPool(t *testing.T) *pgxpool.Pool {
	raw := os.Getenv("DRIFT_TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("DRIFT_TEST_DATABASE_URL is not set")
	}
	databaseURL, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	pool, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := migrate.Up(ctx, pool, schema.Migrations, "migrations", slog.New(slog.DiscardHandler)); err != nil {
		t.Fatal(err)
	}
	return pool
}

func seedTranscript(t *testing.T, pool *pgxpool.Pool) (pgtype.UUID, []db.Message) {
	ctx := context.Background()
	var project, session pgtype.UUID
	if err := pool.QueryRow(ctx, `insert into projects (owner_id, name, path) values ('window-test', 'p', '/p') returning id`).Scan(&project); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `insert into sessions (project_id, owner_id, title, agent) values ($1, 'window-test', 't', 'claude') returning id`, project).Scan(&session); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `delete from projects where id = $1`, project); err != nil {
			t.Error(err)
		}
	})
	base := time.Now()
	messages := make([]db.Message, 0, seededMessages)
	for index := range seededMessages {
		role := "assistant"
		if index%10 == 0 {
			role = "user"
		}
		var m db.Message
		err := pool.QueryRow(ctx,
			`insert into messages (session_id, role, kind, body, created_at) values ($1, $2, 'text', $3, $4) returning id, session_id, role, kind, body, created_at`,
			session, role, fmt.Sprintf(`{"text":"m%d"}`, index), base.Add(time.Duration(index)*time.Millisecond),
		).Scan(&m.ID, &m.SessionID, &m.Role, &m.Kind, &m.Body, &m.CreatedAt)
		if err != nil {
			t.Fatal(err)
		}
		messages = append(messages, m)
	}
	return session, messages
}

func positionOf(m db.Message) *db.GetMessagePositionRow {
	return &db.GetMessagePositionRow{CreatedAt: m.CreatedAt, ID: m.ID}
}

func assertWindow(t *testing.T, got listMessagesResponse, all []db.Message, from, to int, older, newer bool) {
	t.Helper()
	if len(got.Messages) != to-from {
		t.Fatalf("got %d messages, want %d", len(got.Messages), to-from)
	}
	if got.Messages[0].ID != typeid.MessageID(all[from].ID) || got.Messages[len(got.Messages)-1].ID != typeid.MessageID(all[to-1].ID) {
		t.Fatalf("window bounds %s..%s, want m%d..m%d", got.Messages[0].Body, got.Messages[len(got.Messages)-1].Body, from, to-1)
	}
	if got.HasOlder != older || got.HasNewer != newer {
		t.Fatalf("hasOlder=%v hasNewer=%v, want %v %v", got.HasOlder, got.HasNewer, older, newer)
	}
}

func TestMessageWindow(t *testing.T) {
	pool := windowTestPool(t)
	session, all := seedTranscript(t, pool)
	h := &Handler{q: db.New(pool)}
	ctx := context.Background()
	window := func(start, end *db.GetMessagePositionRow) listMessagesResponse {
		out, err := messageWindow{q: h.q, session: session, size: 64, start: start, end: end}.read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}

	assertWindow(t, window(nil, nil), all, 236, 300, true, false)
	assertWindow(t, window(nil, positionOf(all[100])), all, 37, 101, true, true)
	assertWindow(t, window(positionOf(all[100]), nil), all, 100, 164, true, true)
	assertWindow(t, window(positionOf(all[250]), nil), all, 250, 300, true, false)
	assertWindow(t, window(positionOf(all[0]), positionOf(all[20])), all, 0, 21, false, true)
	assertWindow(t, window(positionOf(all[10]), positionOf(all[299])), all, 10, 74, true, true)

	prompt, err := h.lastPrompt(ctx, session)
	if err != nil || prompt == nil || prompt.ID != typeid.MessageID(all[290].ID) {
		t.Fatalf("lastPrompt = %v, %v; want m290", prompt, err)
	}
}
