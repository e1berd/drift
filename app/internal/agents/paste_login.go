package agents

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// startPasteLogin runs a sign-in command that prints a sign-in URL and then reads the code
// the browser shows from stdin.
func startPasteLogin(home string, urlPattern *regexp.Regexp, command func(context.Context) *exec.Cmd) (*Login, error) {
	ctx, cancel := context.WithTimeout(context.Background(), loginTimeout)
	cmd := command(ctx)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	reader, writer := io.Pipe()
	cmd.Stdout, cmd.Stderr = writer, writer
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start %s: %w", cmd.Path, err)
	}

	login := newLogin(cancel)
	login.NeedsCode = true
	var mu sync.Mutex
	login.submit = func(code string) error {
		mu.Lock()
		defer mu.Unlock()
		_, err := io.WriteString(stdin, strings.TrimSpace(code)+"\n")
		return err
	}

	found := make(chan string, 1)
	var output bytes.Buffer
	scanned := make(chan struct{})
	go func() {
		defer close(scanned)
		scanner := bufio.NewScanner(reader)
		scanner.Buffer(make([]byte, 64<<10), 1<<20)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line + "\n")
			if url := urlPattern.FindString(ansiEscape.ReplaceAllString(line, "")); url != "" {
				select {
				case found <- url:
				default:
				}
			}
		}
	}()
	go func() {
		err := cmd.Wait()
		writer.Close()
		<-scanned
		if err != nil {
			err = loginFailure(output.String(), err)
		}
		cancel()
		login.finish(err)
	}()

	select {
	case url := <-found:
		login.URL = url
		return login, nil
	case <-login.Finished():
		if err := login.Result(); err != nil {
			return nil, err
		}
		return nil, errors.New("the sign-in ended without a sign-in link")
	case <-time.After(commandTimeout):
		login.Cancel()
		return nil, fmt.Errorf("%s did not print a sign-in link", cmd.Path)
	}
}

func loginFailure(output string, err error) error {
	if message := lastMessage(output); message != "" {
		return errors.New(message)
	}
	return err
}
