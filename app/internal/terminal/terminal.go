// Package terminal gives every session an interactive shell in its folder, streamed to xterm.js
// over a WebSocket. A shell outlives its socket: a reload or a second browser tab reattaches to
// the same shell and first receives its recent output. Shells nobody watches end after an idle timeout.
//
// Wire protocol: binary frames carry raw terminal bytes in both directions; text frames carry
// JSON control messages (TerminalControl from the browser, TerminalEvent from the server).
package terminal

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/creack/pty"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/labstack/echo/v5"
)

const (
	defaultIdleTimeout = 30 * time.Minute
	defaultLimit       = 8
	readLimit          = 1 << 20
	writeTimeout       = 10 * time.Second
)

var shellCandidates = []string{"fish", "bash", "zsh", "sh"}

var passthroughEnv = []string{
	"PATH", "HOME", "USER", "LANG", "LC_ALL", "TZ", "SSL_CERT_FILE", "SSL_CERT_DIR",
	"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "http_proxy", "https_proxy", "no_proxy",
	"XDG_RUNTIME_DIR", "DOCKER_HOST", "TESTCONTAINERS_RYUK_DISABLED",
}

var errTooMany = errors.New("too many terminals")

// FolderResolver checks that the session belongs to owner and returns its canonical id and the folder it works in.
// Errors should be echo HTTP errors; they are returned to the browser before the upgrade.
type FolderResolver func(ctx context.Context, owner, session string) (id, folder string, err error)

type key struct{ owner, session string }

// Manager owns the running shells, at most one per session.
type Manager struct {
	shell       string
	env         []string
	resolve     FolderResolver
	log         *slog.Logger
	idleTimeout time.Duration
	limit       int

	mu   sync.Mutex
	ttys map[key]*tty
}

// New returns a manager that starts shell (see DetectShell) in session folders found by resolve.
func New(shell string, resolve FolderResolver, log *slog.Logger) *Manager {
	return &Manager{
		shell:       shell,
		env:         Environ(shell),
		resolve:     resolve,
		log:         log,
		idleTimeout: defaultIdleTimeout,
		limit:       defaultLimit,
		ttys:        make(map[key]*tty),
	}
}

// DetectShell returns preferred when set, else the first of fish, bash, zsh and sh found in PATH.
func DetectShell(preferred string) string {
	if preferred != "" {
		return preferred
	}
	for _, name := range shellCandidates {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return "/bin/sh"
}

// Environ is the environment of a shell: the server's safe variables, never its secrets such as DATABASE_URL.
func Environ(shell string) []string {
	env := []string{"TERM=xterm-256color", "COLORTERM=truecolor", "SHELL=" + shell}
	for _, name := range passthroughEnv {
		if value, ok := os.LookupEnv(name); ok {
			env = append(env, name+"="+value)
		}
	}
	if _, ok := os.LookupEnv("LANG"); !ok {
		env = append(env, "LANG=C.UTF-8")
	}
	return env
}

// Register mounts GET /sessions/{session}/terminal on a group that already requires a signed-in identity.
func (m *Manager) Register(g *echo.Group) {
	g.GET("/sessions/:session/terminal", m.serve)
}

// Close ends the shell of a session, e.g. when the session is deleted.
func (m *Manager) Close(owner, session string) {
	m.mu.Lock()
	t := m.ttys[key{owner, session}]
	m.mu.Unlock()
	if t != nil {
		go t.kill()
	}
}

func (m *Manager) open(k key, folder string, size *pty.Winsize) (*tty, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.ttys[k]; ok {
		return t, nil
	}
	if len(m.ttys) >= m.limit {
		return nil, errTooMany
	}
	t, err := startTTY(m.shell, folder, m.env, size, m.idleTimeout, func(t *tty) { m.forget(k, t) })
	if err != nil {
		return nil, err
	}
	m.ttys[k] = t
	return t, nil
}

func (m *Manager) forget(k key, t *tty) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ttys[k] == t {
		delete(m.ttys, k)
	}
}

func (m *Manager) serve(c *echo.Context) error {
	owner := identity.From(c)
	id, folder, err := m.resolve(c.Request().Context(), owner, c.Param("session"))
	if err != nil {
		return err
	}
	t, err := m.open(key{owner, id}, folder, initialSize(c))
	if errors.Is(err, errTooMany) {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "too many open terminals")
	}
	if err != nil {
		m.log.Warn("start terminal", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "cannot start the shell")
	}
	conn, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		return nil
	}
	defer conn.CloseNow()
	conn.SetReadLimit(readLimit)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	viewer, history := t.attach()
	defer t.detach(viewer, m.idleTimeout)
	go func() {
		defer cancel()
		sendOutput(ctx, conn, t, viewer, history)
	}()
	m.receiveInput(ctx, conn, t)
	return nil
}

func initialSize(c *echo.Context) *pty.Winsize {
	cols, _ := strconv.Atoi(c.QueryParam("cols"))
	rows, _ := strconv.Atoi(c.QueryParam("rows"))
	if cols < 1 || cols > maxDimension {
		cols = 80
	}
	if rows < 1 || rows > maxDimension {
		rows = 24
	}
	return &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)}
}

type control struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

type event struct {
	Type     string `json:"type"`
	ExitCode int    `json:"exitCode"`
}

func (m *Manager) receiveInput(ctx context.Context, conn *websocket.Conn, t *tty) {
	for {
		kind, data, err := conn.Read(ctx)
		if err != nil {
			return
		}
		if kind == websocket.MessageBinary {
			if err := t.write(data); err != nil {
				return
			}
			continue
		}
		var msg control
		if err := json.Unmarshal(data, &msg); err != nil || msg.Type != "resize" {
			continue
		}
		if err := t.resize(msg.Cols, msg.Rows); err != nil {
			m.log.Debug("resize terminal", "err", err)
		}
	}
}

func sendOutput(ctx context.Context, conn *websocket.Conn, t *tty, viewer *client, history []byte) {
	if len(history) > 0 && write(ctx, conn, websocket.MessageBinary, history) != nil {
		return
	}
	for {
		select {
		case chunk, ok := <-viewer.out:
			if !ok {
				_ = conn.Close(websocket.StatusTryAgainLater, "terminal output overflow")
				return
			}
			if write(ctx, conn, websocket.MessageBinary, chunk) != nil {
				return
			}
		case <-t.done:
			flush(ctx, conn, viewer)
			exited, _ := json.Marshal(event{Type: "exit", ExitCode: t.exitCode})
			if write(ctx, conn, websocket.MessageText, exited) == nil {
				_ = conn.Close(websocket.StatusNormalClosure, "shell exited")
			}
			return
		case <-ctx.Done():
			return
		}
	}
}

func flush(ctx context.Context, conn *websocket.Conn, viewer *client) {
	for {
		select {
		case chunk, ok := <-viewer.out:
			if !ok || write(ctx, conn, websocket.MessageBinary, chunk) != nil {
				return
			}
		default:
			return
		}
	}
}

func write(ctx context.Context, conn *websocket.Conn, kind websocket.MessageType, data []byte) error {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	return conn.Write(ctx, kind, data)
}
