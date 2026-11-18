package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

const (
	checksumsAsset = "checksums.txt"
	releasesToList = 30
)

type githubAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
}

type githubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	HTMLURL     string        `json:"html_url"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	PublishedAt time.Time     `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

type githubRepository struct {
	HTMLURL         string `json:"html_url"`
	StargazersCount int    `json:"stargazers_count"`
}

type github struct {
	api  *url.URL
	repo string
	http *http.Client
}

func platformAsset() string {
	return fmt.Sprintf("drift_%s_%s.gz", runtime.GOOS, runtime.GOARCH)
}

func (r githubRelease) asset(name string) (githubAsset, bool) {
	for _, asset := range r.Assets {
		if asset.Name == name {
			return asset, true
		}
	}
	return githubAsset{}, false
}

func (r githubRelease) installable() bool {
	_, binary := r.asset(platformAsset())
	_, checksums := r.asset(checksumsAsset)
	return binary && checksums
}

func (g github) repository(ctx context.Context) (githubRepository, error) {
	var out githubRepository
	err := g.getJSON(ctx, g.api.JoinPath("repos", g.repo), &out)
	return out, err
}

func (g github) releases(ctx context.Context) ([]githubRelease, error) {
	endpoint := g.api.JoinPath("repos", g.repo, "releases")
	query := endpoint.Query()
	query.Set("per_page", fmt.Sprint(releasesToList))
	endpoint.RawQuery = query.Encode()
	var out []githubRelease
	err := g.getJSON(ctx, endpoint, &out)
	return out, err
}

func (g github) getJSON(ctx context.Context, endpoint *url.URL, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := g.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github %s: %s", endpoint.Path, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
