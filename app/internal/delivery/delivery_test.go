package delivery

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/e1berd/drift/app/internal/events"
)

func installableAssets() []githubAsset {
	return []githubAsset{{Name: platformAsset()}, {Name: checksumsAsset}}
}

func fakeGitHub(t *testing.T, releases []githubRelease) *url.URL {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/drift", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(githubRepository{HTMLURL: "https://github.com/owner/drift", StargazersCount: 42})
	})
	mux.HandleFunc("/repos/owner/drift/releases", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(releases)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	api, _ := url.Parse(server.URL)
	return api
}

func TestSnapshotPicksLatestStableInstallableRelease(t *testing.T) {
	previous := version
	version = "v1.0.0"
	t.Cleanup(func() { version = previous })

	api := fakeGitHub(t, []githubRelease{
		{TagName: "v2.0.0-rc.1", Prerelease: true, Assets: installableAssets()},
		{TagName: "v1.9.0"},
		{TagName: "v1.2.0", Body: "notes", Assets: installableAssets()},
		{TagName: "v1.1.0", Assets: installableAssets()},
		{TagName: "v3.0.0", Draft: true, Assets: installableAssets()},
	})
	service, err := New(Options{Repository: "owner/drift", APIURL: api, Shutdown: func() {}}, events.NewHub(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	if err := service.refresh(context.Background()); err != nil {
		t.Fatal(err)
	}

	got := service.snapshot()
	if got.LatestRelease == nil || got.LatestRelease.Version != "v1.2.0" || got.LatestRelease.Notes != "notes" {
		t.Fatalf("latest release = %+v", got.LatestRelease)
	}
	if !got.UpdateAvailable || got.Repository.Stars != 42 || !got.ReleaseBuild {
		t.Fatalf("snapshot = %+v", got)
	}
}
