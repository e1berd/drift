// Package gitrepo serves the Git panel of a session: status, diffs, staging, commits, branches,
// pull and push of the folder the session's agent works in, by running the git CLI.
//
// Names follow the Google API design guide: the status is a singleton sub-resource of the session
// (AIP-156), branches and commits are list methods with page_size/page_token (AIP-158), and every
// change is a custom method such as git:stage or git:commit (AIP-136) that returns the fresh status.
package gitrepo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/labstack/echo/v5"
)

const (
	localTimeout  = 15 * time.Second
	remoteTimeout = 2 * time.Minute
	maxStderr     = 4 << 10
)

var passthroughEnv = []string{
	"PATH", "HOME", "USER", "TZ", "SSH_AUTH_SOCK", "SSL_CERT_FILE", "SSL_CERT_DIR", "XDG_CONFIG_HOME",
	"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY", "http_proxy", "https_proxy", "all_proxy", "no_proxy",
}

// Environ is the environment of git commands: the server's safe variables, never its secrets such as
// DATABASE_URL, with prompts disabled and repository discovery stopped at the parent of folder, so a
// folder that lives inside another checkout is not mistaken for that checkout.
func Environ(folder string) []string {
	env := []string{
		"LC_ALL=C",
		"GIT_TERMINAL_PROMPT=0",
		"GIT_OPTIONAL_LOCKS=0",
		"GIT_LITERAL_PATHSPECS=1",
		"GIT_CEILING_DIRECTORIES=" + filepath.Dir(folder),
	}
	for _, name := range passthroughEnv {
		if value, ok := os.LookupEnv(name); ok {
			env = append(env, name+"="+value)
		}
	}
	return env
}

// Error is a git command that ran and failed; Message is what git printed to stderr.
type Error struct {
	Message  string
	ExitCode int
}

func (e *Error) Error() string { return e.Message }

type repo struct {
	dir string
	env []string
}

func open(folder string) (repo, error) {
	dir, err := filepath.EvalSymlinks(folder)
	if err != nil {
		return repo{}, echo.NewHTTPError(http.StatusNotFound, "session folder is missing")
	}
	return repo{dir: dir, env: Environ(dir)}, nil
}

type command struct {
	args    []string
	timeout time.Duration
	stdout  io.Writer
	okCodes []int
}

func (r repo) git(ctx context.Context, args ...string) ([]byte, error) {
	var out bytes.Buffer
	err := r.run(ctx, command{args: args, timeout: localTimeout, stdout: &out})
	return out.Bytes(), err
}

func (r repo) remote(ctx context.Context, args ...string) error {
	return r.run(ctx, command{args: args, timeout: remoteTimeout, stdout: &bytes.Buffer{}})
}

func (r repo) run(ctx context.Context, c command) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	args := append([]string{"-c", "safe.directory=" + r.dir, "-c", "core.quotePath=false"}, c.args...)
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.dir
	cmd.Env = r.env
	stderr := &capped{limit: maxStderr}
	cmd.Stderr = stderr
	cmd.Stdout = c.stdout
	err := cmd.Run()
	if ctx.Err() != nil {
		return echo.NewHTTPError(http.StatusGatewayTimeout, "git did not finish in time")
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		for _, ok := range c.okCodes {
			if exit.ExitCode() == ok {
				return nil
			}
		}
		return &Error{Message: gitMessage(stderr.bytes(), exit.ExitCode()), ExitCode: exit.ExitCode()}
	}
	if err != nil {
		return fmt.Errorf("run git: %w", err)
	}
	return nil
}

func gitMessage(stderr []byte, code int) string {
	lines := make([]string, 0, 4)
	for line := range strings.SplitSeq(string(stderr), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "hint:") {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return fmt.Sprintf("git exited with code %d", code)
	}
	return strings.ToValidUTF8(strings.Join(lines, "\n"), string(utf8.RuneError))
}

type capped struct {
	buf       *bytes.Buffer
	limit     int
	truncated bool
}

func (c *capped) Write(p []byte) (int, error) {
	if c.buf == nil {
		c.buf = &bytes.Buffer{}
	}
	room := c.limit - c.buf.Len()
	if room <= 0 {
		c.truncated = c.truncated || len(p) > 0
		return len(p), nil
	}
	if len(p) > room {
		c.buf.Write(p[:room])
		c.truncated = true
		return len(p), nil
	}
	return c.buf.Write(p)
}

func (c *capped) bytes() []byte {
	if c.buf == nil {
		return nil
	}
	return c.buf.Bytes()
}

func (r repo) isRepository(ctx context.Context) (bool, error) {
	out, err := r.git(ctx, "rev-parse", "--show-toplevel")
	var gitErr *Error
	if errors.As(err, &gitErr) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	top, err := filepath.EvalSymlinks(strings.TrimSpace(string(out)))
	return err == nil && top == r.dir, nil
}

func (r repo) hasHead(ctx context.Context) bool {
	_, err := r.git(ctx, "rev-parse", "--verify", "--quiet", "HEAD")
	return err == nil
}

// CleanPaths checks paths sent by the browser: relative to the repository root, without "..",
// absolute paths or empty names, and returns them in slash form.
func CleanPaths(raw []string) ([]string, error) {
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		clean, err := cleanPath(p)
		if err != nil {
			return nil, err
		}
		out = append(out, clean)
	}
	return out, nil
}

func cleanPath(p string) (string, error) {
	if p == "" || strings.ContainsRune(p, 0) || strings.Contains(p, "\\") || !filepath.IsLocal(p) {
		return "", echo.NewHTTPError(http.StatusBadRequest, "invalid path: "+p)
	}
	clean := filepath.ToSlash(filepath.Clean(p))
	if clean == ".git" || strings.HasPrefix(clean, ".git/") {
		return "", echo.NewHTTPError(http.StatusBadRequest, "invalid path: "+p)
	}
	return clean, nil
}
