package terminal

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/labstack/echo/v5"
)

func TestScrollbackKeepsTheTailFromALineStart(t *testing.T) {
	s := scrollback{limit: 10}
	s.write([]byte("first\nsecond\nthird"))
	if got := string(s.data); got != "third" {
		t.Fatalf("history = %q, want %q", got, "third")
	}
	s.write([]byte("0123456789abc"))
	if len(s.data) > 10 {
		t.Fatalf("history grew past its limit: %d bytes", len(s.data))
	}
}

func startServer(t *testing.T, folder string) *httptest.Server {
	t.Helper()
	resolve := func(_ context.Context, _, session string) (string, string, error) {
		if session != "s1" {
			return "", "", echo.NewHTTPError(http.StatusNotFound, "not found")
		}
		return session, folder, nil
	}
	manager := New("/bin/sh", resolve, slog.New(slog.NewTextHandler(io.Discard, nil)))
	e := echo.New()
	manager.Register(e.Group("/api"))
	server := httptest.NewServer(e)
	t.Cleanup(server.Close)
	return server
}

func dial(t *testing.T, ctx context.Context, server *httptest.Server) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/api/sessions/s1/terminal?cols=100&rows=30", nil)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func readUntil(t *testing.T, ctx context.Context, conn *websocket.Conn, want string) {
	t.Helper()
	var seen bytes.Buffer
	for !strings.Contains(seen.String(), want) {
		kind, data, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("waiting for %q, got %q: %v", want, seen.String(), err)
		}
		if kind == websocket.MessageBinary {
			seen.Write(data)
		}
	}
}

func TestShellRunsInTheSessionFolderAndSurvivesReconnects(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	folder := t.TempDir()
	server := startServer(t, folder)

	first := dial(t, ctx, server)
	if err := first.Write(ctx, websocket.MessageBinary, []byte("echo answer-$((40+2)); pwd\n")); err != nil {
		t.Fatal(err)
	}
	readUntil(t, ctx, first, folder)
	first.CloseNow()

	second := dial(t, ctx, server)
	defer second.CloseNow()
	readUntil(t, ctx, second, "answer-42")

	resize, _ := json.Marshal(control{Type: "resize", Cols: 120, Rows: 40})
	if err := second.Write(ctx, websocket.MessageText, resize); err != nil {
		t.Fatal(err)
	}
	if err := second.Write(ctx, websocket.MessageBinary, []byte("stty size; exit 3\n")); err != nil {
		t.Fatal(err)
	}
	readUntil(t, ctx, second, "40 120")
	for {
		kind, data, err := second.Read(ctx)
		if err != nil {
			t.Fatalf("no exit event: %v", err)
		}
		if kind != websocket.MessageText {
			continue
		}
		var got event
		if err := json.Unmarshal(data, &got); err != nil || got.Type != "exit" || got.ExitCode != 3 {
			t.Fatalf("event = %s, want exit with code 3", data)
		}
		return
	}
}

func TestUnknownSessionIsRejectedBeforeTheUpgrade(t *testing.T) {
	server := startServer(t, t.TempDir())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, response, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/api/sessions/other/terminal", nil)
	if err == nil || response == nil || response.StatusCode != http.StatusNotFound {
		t.Fatalf("dial = %v, %v; want 404", response, err)
	}
}
