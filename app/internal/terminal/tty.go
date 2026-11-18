package terminal

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
)

const (
	historyLimit = 256 << 10
	clientQueue  = 256
	killGrace    = 3 * time.Second
	maxDimension = 1000
)

type client struct {
	out chan []byte
}

type tty struct {
	cmd      *exec.Cmd
	pty      *os.File
	done     chan struct{}
	exitCode int

	mu      sync.Mutex
	history scrollback
	clients map[*client]struct{}
	idle    *time.Timer
}

func startTTY(shell, dir string, env []string, size *pty.Winsize, idleAfter time.Duration, onExit func(*tty)) (*tty, error) {
	cmd := exec.Command(shell)
	cmd.Dir = dir
	cmd.Env = env
	file, err := pty.StartWithSize(cmd, size)
	if err != nil {
		return nil, err
	}
	t := &tty{
		cmd:     cmd,
		pty:     file,
		done:    make(chan struct{}),
		history: scrollback{limit: historyLimit},
		clients: make(map[*client]struct{}),
	}
	t.idle = time.AfterFunc(idleAfter, t.kill)
	go t.pump(onExit)
	return t, nil
}

func (t *tty) pump(onExit func(*tty)) {
	buf := make([]byte, 32<<10)
	for {
		n, err := t.pty.Read(buf)
		if n > 0 {
			t.broadcast(bytes.Clone(buf[:n]))
		}
		if err != nil {
			break
		}
	}
	t.exitCode = exitCode(t.cmd.Wait())
	t.idle.Stop()
	_ = t.pty.Close()
	close(t.done)
	onExit(t)
}

func exitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	if err != nil {
		return -1
	}
	return 0
}

func (t *tty) broadcast(chunk []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.history.write(chunk)
	for c := range t.clients {
		select {
		case c.out <- chunk:
		default:
			delete(t.clients, c)
			close(c.out)
		}
	}
}

func (t *tty) attach() (*client, []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.idle.Stop()
	c := &client{out: make(chan []byte, clientQueue)}
	t.clients[c] = struct{}{}
	return c, bytes.Clone(t.history.data)
}

func (t *tty) detach(c *client, idleAfter time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.clients[c]; ok {
		delete(t.clients, c)
		close(c.out)
	}
	if len(t.clients) == 0 && !t.exited() {
		t.idle.Reset(idleAfter)
	}
}

func (t *tty) write(input []byte) error {
	_, err := t.pty.Write(input)
	return err
}

func (t *tty) resize(cols, rows int) error {
	if cols < 1 || rows < 1 || cols > maxDimension || rows > maxDimension {
		return errors.New("terminal size out of range")
	}
	return pty.Setsize(t.pty, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
}

func (t *tty) exited() bool {
	select {
	case <-t.done:
		return true
	default:
		return false
	}
}

func (t *tty) kill() {
	if t.exited() {
		return
	}
	pid := t.cmd.Process.Pid
	_ = syscall.Kill(-pid, syscall.SIGHUP)
	select {
	case <-t.done:
	case <-time.After(killGrace):
		_ = syscall.Kill(-pid, syscall.SIGKILL)
	}
}

type scrollback struct {
	data  []byte
	limit int
}

func (s *scrollback) write(chunk []byte) {
	s.data = append(s.data, chunk...)
	over := len(s.data) - s.limit
	if over <= 0 {
		return
	}
	cut := over
	if newline := bytes.IndexByte(s.data[over:], '\n'); newline >= 0 && newline < 4<<10 {
		cut = over + newline + 1
	}
	s.data = s.data[:copy(s.data, s.data[cut:])]
}
