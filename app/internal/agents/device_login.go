package agents

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"regexp"
	"time"
)

// startDeviceLogin runs a sign-in command that prints a sign-in URL with the device code in it
// and finishes the sign-in by itself once the user confirms in the browser.
func startDeviceLogin(urlPattern *regexp.Regexp, command func(context.Context) *exec.Cmd) (*Login, error) {
	ctx, cancel := context.WithTimeout(context.Background(), loginTimeout)
	cmd := command(ctx)
	reader, writer := io.Pipe()
	cmd.Stdout, cmd.Stderr = writer, writer
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start %s: %w", cmd.Path, err)
	}

	login := newLogin(cancel)

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
			if raw := urlPattern.FindString(ansiEscape.ReplaceAllString(line, "")); raw != "" {
				select {
				case found <- raw:
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
	case raw := <-found:
		login.URL = raw
		login.UserCode = deviceCode(raw)
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

func deviceCode(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return parsed.Query().Get("user_code")
}
