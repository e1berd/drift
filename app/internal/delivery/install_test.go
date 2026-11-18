package delivery

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func gzipped(t *testing.T, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func fakeRelease(t *testing.T, tag, reportedVersion string, corruptChecksum bool) (githubRelease, *httptest.Server) {
	t.Helper()
	archive := gzipped(t, "#!/bin/sh\necho "+reportedVersion+"\n")
	sum := sha256.Sum256(archive)
	checksum := hex.EncodeToString(sum[:])
	if corruptChecksum {
		checksum = strings.Repeat("0", 64)
	}
	checksums := checksum + "  " + platformAsset() + "\n"

	mux := http.NewServeMux()
	mux.HandleFunc("/binary", func(w http.ResponseWriter, _ *http.Request) { w.Write(archive) })
	mux.HandleFunc("/checksums", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte(checksums)) })
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return githubRelease{
		TagName: tag,
		Assets: []githubAsset{
			{Name: platformAsset(), DownloadURL: server.URL + "/binary", Size: int64(len(archive))},
			{Name: checksumsAsset, DownloadURL: server.URL + "/checksums"},
		},
	}, server
}

func installedBinary(t *testing.T) installer {
	t.Helper()
	path := filepath.Join(t.TempDir(), "drift")
	if err := os.WriteFile(path, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	return installer{http: http.DefaultClient, executable: path}
}

func TestInstallReplacesBinary(t *testing.T) {
	release, _ := fakeRelease(t, "v1.3.0", "v1.3.0", false)
	target := installedBinary(t)
	var lastProgress int64
	err := target.install(context.Background(), release, func(downloaded, _ int64) { lastProgress = downloaded })
	if err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(target.executable)
	if !strings.Contains(string(content), "echo v1.3.0") {
		t.Fatalf("binary not replaced: %q", content)
	}
	if lastProgress != release.Assets[0].Size {
		t.Fatalf("progress %d, want %d", lastProgress, release.Assets[0].Size)
	}
	entries, _ := os.ReadDir(filepath.Dir(target.executable))
	if len(entries) != 1 {
		t.Fatalf("temporary files left behind: %v", entries)
	}
}

func TestInstallRejectsBadChecksum(t *testing.T) {
	release, _ := fakeRelease(t, "v1.3.0", "v1.3.0", true)
	target := installedBinary(t)
	err := target.install(context.Background(), release, func(int64, int64) {})
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("want checksum error, got %v", err)
	}
	if content, _ := os.ReadFile(target.executable); string(content) != "old" {
		t.Fatal("binary replaced despite bad checksum")
	}
}

func TestInstallRejectsWrongVersion(t *testing.T) {
	release, _ := fakeRelease(t, "v1.3.0", "v1.2.0", false)
	target := installedBinary(t)
	err := target.install(context.Background(), release, func(int64, int64) {})
	if err == nil || !strings.Contains(err.Error(), "reports version") {
		t.Fatalf("want version error, got %v", err)
	}
	if content, _ := os.ReadFile(target.executable); string(content) != "old" {
		t.Fatal("binary replaced despite wrong version")
	}
}
