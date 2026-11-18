package agents

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	antigravitySignInURL = regexp.MustCompile(`https://accounts\.google\.com/\S+`)
	emailAddress         = regexp.MustCompile(`[\w.+-]+@[\w-]+(\.[\w-]+)+`)
	percentValue         = regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*%$`)
	notSignedIn          = regexp.MustCompile(`(?i)authentication (required|failed)|not signed in|sign in once`)
)

var antigravityTimeLayouts = []string{time.RFC3339, "2006-01-02 15:04:05 MST", "2006-01-02 15:04 MST", "2006-01-02 15:04:05Z07:00"}

const antigravitySettings = `{"agentMode": ""}` + "\n"

type antigravity struct{ cli CLI }

func (d antigravity) command(ctx context.Context, home string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, d.cli.Antigravity, args...)
	cmd.Env = d.cli.Env(home)
	cmd.Dir = home
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.WaitDelay = 2 * time.Second
	return cmd
}

// StartLogin runs `agy acp --login`, the terminal Google sign-in for machines without a
// browser: it prints a Google URL and reads the authorization code from stdin. The process
// has no controlling terminal, so the CLI cannot ask on /dev/tty instead.
func (d antigravity) StartLogin(home string) (*Login, error) {
	return startPasteLogin(home, antigravitySignInURL, func(ctx context.Context) *exec.Cmd {
		return d.command(ctx, home, "acp", "--login")
	})
}

// Account reads the signed-in Google account from the print-mode answer to /usage.
func (d antigravity) Account(ctx context.Context, home string) (Account, error) {
	report, err := d.usageReport(ctx, home)
	if err != nil {
		return Account{}, err
	}
	return report.account, nil
}

// Usage reads the per-model quota from the print-mode answer to /usage.
func (d antigravity) Usage(ctx context.Context, home string) ([]Window, error) {
	report, err := d.usageReport(ctx, home)
	if err != nil {
		if errors.Is(err, errAntigravitySignedOut) {
			return nil, UsageUnavailable{Reason: "signed_out"}
		}
		return nil, err
	}
	return report.windows, nil
}

var errAntigravitySignedOut = errors.New("antigravity is not signed in")

func (d antigravity) usageReport(ctx context.Context, home string) (antigravityReport, error) {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	cmd := d.command(ctx, home, "-p", "/usage")
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.CombinedOutput()
	text := ansiEscape.ReplaceAllString(string(out), "")
	if notSignedIn.MatchString(text) {
		return antigravityReport{}, errAntigravitySignedOut
	}
	if err != nil {
		return antigravityReport{}, loginFailure(text, err)
	}
	return parseAntigravityUsage(text), nil
}

// prepareAntigravity keeps headless runs in the Default agent mode, which only asks when a
// tool needs approval; without a saved mode agy runs in Auto. Existing settings stay.
func prepareAntigravity(home string) error {
	dir := filepath.Join(home, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(dir, "settings.json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := file.WriteString(antigravitySettings); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

type antigravityReport struct {
	account Account
	windows []Window
}

func parseAntigravityUsage(text string) antigravityReport {
	report := antigravityReport{windows: []Window{}}
	report.account.Email = emailAddress.FindString(text)
	var header []string
	for line := range strings.SplitSeq(text, "\n") {
		fields := splitRecord(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) >= 2 && isPlanKey(fields[0]) {
			report.account.Plan = fields[1]
			continue
		}
		window, ok := quotaWindow(fields, header)
		if ok {
			report.windows = append(report.windows, window)
			continue
		}
		if len(fields) >= 2 && !emailAddress.MatchString(line) {
			header = lowerAll(fields)
		}
	}
	return report
}

func splitRecord(line string) []string {
	fields := []string{}
	for field := range strings.SplitSeq(line, "\t") {
		if field = strings.TrimSpace(field); field != "" {
			fields = append(fields, field)
		}
	}
	return fields
}

func isPlanKey(key string) bool {
	key = strings.ToLower(strings.TrimSuffix(key, ":"))
	return key == "plan" || key == "tier" || key == "your plan"
}

func lowerAll(fields []string) []string {
	out := make([]string, len(fields))
	for i, field := range fields {
		out[i] = strings.ToLower(field)
	}
	return out
}

func quotaWindow(fields, header []string) (Window, bool) {
	var (
		window  Window
		label   string
		percent = -1.0
		column  = -1
	)
	for i, field := range fields {
		if value, ok := parsePercent(field); ok && percent < 0 {
			percent, column = value, i
			continue
		}
		if reset, ok := parseResetTime(field); ok {
			window.ResetTime = &reset
			continue
		}
		if label == "" {
			label = field
		}
	}
	if percent < 0 || label == "" {
		return Window{}, false
	}
	window.Label = label
	window.ID = slug(label)
	window.UsedPercent = 100 - percent
	if column < len(header) && strings.Contains(header[column], "used") {
		window.UsedPercent = percent
	}
	return window, true
}

func parsePercent(field string) (float64, bool) {
	if match := percentValue.FindStringSubmatch(field); match != nil {
		value, err := strconv.ParseFloat(match[1], 64)
		return value, err == nil && value <= 100
	}
	if !strings.Contains(field, ".") {
		return 0, false
	}
	value, err := strconv.ParseFloat(field, 64)
	if err != nil || value < 0 || value > 1 {
		return 0, false
	}
	return value * 100, true
}

func parseResetTime(field string) (time.Time, bool) {
	for _, layout := range antigravityTimeLayouts {
		if parsed, err := time.Parse(layout, field); err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

func slug(label string) string {
	var out strings.Builder
	for _, r := range strings.ToLower(label) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out.WriteRune(r)
		case out.Len() > 0 && !strings.HasSuffix(out.String(), "_"):
			out.WriteByte('_')
		}
	}
	return strings.TrimSuffix(out.String(), "_")
}
