package delivery

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	maxBinarySize   = 512 << 20
	smokeTestLimit  = 15 * time.Second
	versionArgument = "version"
	writeAccess     = 2
)

type progressFunc func(downloaded, total int64)

type installer struct {
	http       *http.Client
	executable string
}

type countingReader struct {
	reader     io.Reader
	downloaded int64
	total      int64
	report     progressFunc
	reportedAt time.Time
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.downloaded += int64(n)
	if time.Since(r.reportedAt) > 250*time.Millisecond || errors.Is(err, io.EOF) {
		r.reportedAt = time.Now()
		r.report(r.downloaded, r.total)
	}
	return n, err
}

func (i installer) writable() bool {
	return syscall.Access(filepath.Dir(i.executable), writeAccess) == nil
}

func (i installer) install(ctx context.Context, release githubRelease, report progressFunc) error {
	binary, ok := release.asset(platformAsset())
	if !ok {
		return fmt.Errorf("release %s has no %s", release.TagName, platformAsset())
	}
	checksums, ok := release.asset(checksumsAsset)
	if !ok {
		return fmt.Errorf("release %s has no %s", release.TagName, checksumsAsset)
	}
	expected, err := i.expectedChecksum(ctx, checksums, binary.Name)
	if err != nil {
		return err
	}
	staged, err := i.download(ctx, binary, expected, report)
	if err != nil {
		return err
	}
	defer os.Remove(staged)
	if err := smokeTest(ctx, staged, release.TagName); err != nil {
		return err
	}
	return os.Rename(staged, i.executable)
}

func (i installer) expectedChecksum(ctx context.Context, checksums githubAsset, name string) (string, error) {
	body, err := i.get(ctx, checksums.DownloadURL)
	if err != nil {
		return "", err
	}
	defer body.Close()
	scanner := bufio.NewScanner(io.LimitReader(body, 1<<20))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == name {
			return strings.ToLower(fields[0]), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("%s has no checksum for %s", checksums.Name, name)
}

func (i installer) download(ctx context.Context, asset githubAsset, expected string, report progressFunc) (string, error) {
	body, err := i.get(ctx, asset.DownloadURL)
	if err != nil {
		return "", err
	}
	defer body.Close()

	compressed, err := os.CreateTemp(filepath.Dir(i.executable), ".drift-download-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(compressed.Name())
	defer compressed.Close()

	hash := sha256.New()
	counter := &countingReader{reader: io.LimitReader(body, maxBinarySize), total: asset.Size, report: report}
	if _, err := io.Copy(io.MultiWriter(compressed, hash), counter); err != nil {
		return "", fmt.Errorf("download %s: %w", asset.Name, err)
	}
	if actual := hex.EncodeToString(hash.Sum(nil)); actual != expected {
		return "", fmt.Errorf("checksum mismatch for %s: got %s, want %s", asset.Name, actual, expected)
	}
	if _, err := compressed.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	return i.unpack(compressed)
}

func (i installer) unpack(compressed io.Reader) (string, error) {
	archive, err := gzip.NewReader(compressed)
	if err != nil {
		return "", err
	}
	defer archive.Close()

	staged, err := os.CreateTemp(filepath.Dir(i.executable), ".drift-new-*")
	if err != nil {
		return "", err
	}
	if err := writeExecutable(staged, archive); err != nil {
		os.Remove(staged.Name())
		return "", err
	}
	return staged.Name(), nil
}

func writeExecutable(file *os.File, content io.Reader) error {
	defer file.Close()
	if _, err := io.Copy(file, io.LimitReader(content, maxBinarySize)); err != nil {
		return err
	}
	if err := file.Chmod(0o755); err != nil {
		return err
	}
	return file.Sync()
}

func smokeTest(ctx context.Context, binary, want string) error {
	ctx, cancel := context.WithTimeout(ctx, smokeTestLimit)
	defer cancel()
	out, err := exec.CommandContext(ctx, binary, versionArgument).Output()
	if err != nil {
		return fmt.Errorf("new binary does not start: %w", err)
	}
	if got := string(bytes.TrimSpace(out)); got != want {
		return fmt.Errorf("new binary reports version %q, want %q", got, want)
	}
	return nil
}

func (i installer) get(ctx context.Context, rawURL string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := i.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download %s: %s", rawURL, resp.Status)
	}
	return resp.Body, nil
}

// Relaunch replaces the current process with the binary at path, keeping its pid, arguments
// and environment, so systemd and container supervisors see the same service keep running.
func Relaunch(path string) error {
	return syscall.Exec(path, os.Args, os.Environ())
}
