package agents

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string { return e.Message }

type rpcMessage struct {
	ID     *int64          `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *rpcError       `json:"error,omitempty"`
}

// rpcServer speaks newline-delimited JSON-RPC with a child process over stdio.
type rpcServer struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	cancel context.CancelFunc
	notes  chan rpcMessage
	closed chan struct{}

	mu      sync.Mutex
	next    int64
	waiting map[int64]chan rpcMessage
}

func startRPC(ctx context.Context, bin string, env []string, dir string, args ...string) (*rpcServer, error) {
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env, cmd.Dir, cmd.WaitDelay = env, dir, 2*time.Second
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start %s: %w", bin, err)
	}
	server := &rpcServer{
		cmd: cmd, stdin: stdin, cancel: cancel,
		notes: make(chan rpcMessage, 32), closed: make(chan struct{}),
		waiting: make(map[int64]chan rpcMessage),
	}
	go server.read(stdout)
	return server, nil
}

func (s *rpcServer) read(stdout io.Reader) {
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64<<10), 16<<20)
	for scanner.Scan() {
		var msg rpcMessage
		if json.Unmarshal(scanner.Bytes(), &msg) != nil {
			continue
		}
		s.dispatch(msg)
	}
	_ = s.cmd.Wait()
	s.mu.Lock()
	close(s.closed)
	for id, reply := range s.waiting {
		close(reply)
		delete(s.waiting, id)
	}
	s.mu.Unlock()
}

func (s *rpcServer) dispatch(msg rpcMessage) {
	switch {
	case msg.ID != nil && msg.Method == "":
		s.mu.Lock()
		reply, ok := s.waiting[*msg.ID]
		delete(s.waiting, *msg.ID)
		s.mu.Unlock()
		if ok {
			reply <- msg
		}
	case msg.ID != nil:
		_ = s.write(rpcMessage{ID: msg.ID, Error: &rpcError{Code: -32601, Message: "not supported by drift"}})
	default:
		select {
		case s.notes <- msg:
		default:
		}
	}
}

func (s *rpcServer) write(msg rpcMessage) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err = s.stdin.Write(append(raw, '\n'))
	return err
}

func (s *rpcServer) call(ctx context.Context, method string, params, out any) error {
	reply := make(chan rpcMessage, 1)
	s.mu.Lock()
	s.next++
	id := s.next
	s.waiting[id] = reply
	s.mu.Unlock()

	msg := rpcMessage{ID: &id, Method: method}
	if params != nil {
		raw, err := json.Marshal(params)
		if err != nil {
			return err
		}
		msg.Params = raw
	}
	if err := s.write(msg); err != nil {
		return err
	}
	select {
	case resp, ok := <-reply:
		if !ok {
			return errors.New(method + ": process exited")
		}
		if resp.Error != nil {
			return resp.Error
		}
		if out == nil {
			return nil
		}
		return json.Unmarshal(resp.Result, out)
	case <-ctx.Done():
		s.mu.Lock()
		delete(s.waiting, id)
		s.mu.Unlock()
		return fmt.Errorf("%s: %w", method, ctx.Err())
	}
}

func (s *rpcServer) notify(method string) error {
	return s.write(rpcMessage{Method: method})
}

func (s *rpcServer) close() {
	s.stdin.Close()
	s.cancel()
	<-s.closed
}
