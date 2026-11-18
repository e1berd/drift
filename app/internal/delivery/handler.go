package delivery

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"golang.org/x/mod/semver"
)

const (
	StateIdle        = "idle"
	StateDownloading = "downloading"
	StateRestarting  = "restarting"
	StateFailed      = "failed"

	blockerDevelopment = "development_build"
	blockerReadOnly    = "read_only"
)

// Release is one GitHub release of Drift.
type Release struct {
	Version     string    `json:"version"`
	Name        string    `json:"name"`
	Notes       string    `json:"notes"`
	URL         string    `json:"url"`
	PublishTime time.Time `json:"publishTime"`
	Prerelease  bool      `json:"prerelease"`
	Installable bool      `json:"installable"`
}

// Repository is the GitHub repository releases come from.
type Repository struct {
	URL   string `json:"url"`
	Stars int    `json:"stars"`
}

// Installation is the update in progress, shared by every user of the instance.
type Installation struct {
	State           string `json:"state"`
	TargetVersion   string `json:"targetVersion,omitempty"`
	DownloadedBytes int64  `json:"downloadedBytes"`
	TotalBytes      int64  `json:"totalBytes"`
	Error           string `json:"error,omitempty"`
}

// Delivery is the update singleton of the instance (AIP-156).
type Delivery struct {
	CurrentVersion  string       `json:"currentVersion"`
	ReleaseBuild    bool         `json:"releaseBuild"`
	LatestRelease   *Release     `json:"latestRelease,omitempty"`
	UpdateAvailable bool         `json:"updateAvailable"`
	Repository      Repository   `json:"repository"`
	InstallBlocker  string       `json:"installBlocker,omitempty"`
	Installation    Installation `json:"installation"`
	CheckTime       *time.Time   `json:"checkTime,omitempty"`
}

type installInput struct {
	Version string `json:"version"`
}

// Register mounts the routes on a group that already requires a signed-in identity.
func (s *Service) Register(g *echo.Group) {
	g.GET("/delivery", s.get)
	g.GET("/delivery/releases", s.listReleases)
	g.POST(`/delivery\:check`, s.checkNow)
	g.POST(`/delivery\:install`, s.install)
}

func toRelease(release githubRelease) Release {
	return Release{
		Version: release.TagName, Name: release.Name, Notes: release.Body, URL: release.HTMLURL,
		PublishTime: release.PublishedAt, Prerelease: release.Prerelease, Installable: release.installable(),
	}
}

func (s *Service) installBlocker() string {
	if version == developmentVersion {
		return blockerDevelopment
	}
	if !s.installer.writable() {
		return blockerReadOnly
	}
	return ""
}

func (s *Service) snapshot() Delivery {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := Delivery{
		CurrentVersion: version,
		ReleaseBuild:   isReleaseBuild(version),
		Repository:     Repository{URL: s.repository.HTMLURL, Stars: s.repository.StargazersCount},
		InstallBlocker: s.installBlocker(),
		Installation:   s.installation,
	}
	if !s.checkTime.IsZero() {
		checked := s.checkTime
		out.CheckTime = &checked
	}
	if latest := s.latestLocked(); latest != nil {
		release := toRelease(*latest)
		out.LatestRelease = &release
		out.UpdateAvailable = newerThan(latest.TagName, version)
	}
	return out
}

func (s *Service) get(c *echo.Context) error {
	return c.JSON(http.StatusOK, s.snapshot())
}

func (s *Service) checkNow(c *echo.Context) error {
	if err := s.refreshIfStale(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "GitHub is unavailable: "+err.Error())
	}
	return c.JSON(http.StatusOK, s.snapshot())
}

func (s *Service) listReleases(c *echo.Context) error {
	if err := s.refreshIfStale(c.Request().Context()); err != nil {
		s.log.Warn("list releases", "err", err)
	}
	s.mu.Lock()
	out := make([]Release, 0, len(s.releases))
	for _, release := range s.releases {
		if !release.Draft && semver.IsValid(release.TagName) {
			out = append(out, toRelease(release))
		}
	}
	s.mu.Unlock()
	return c.JSON(http.StatusOK, map[string]any{"releases": out})
}

func (s *Service) install(c *echo.Context) error {
	var in installInput
	if err := c.Bind(&in); err != nil || strings.TrimSpace(in.Version) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "version is required")
	}
	if blocker := s.installBlocker(); blocker != "" {
		return echo.NewHTTPError(http.StatusConflict, "this Drift cannot update itself: "+blocker)
	}
	if err := s.refreshIfStale(c.Request().Context()); err != nil {
		s.log.Warn("refresh before install", "err", err)
	}

	s.mu.Lock()
	release, found := s.releaseLocked(in.Version)
	busy := s.installation.State == StateDownloading || s.installation.State == StateRestarting
	if found && release.installable() && !busy {
		s.installation = Installation{State: StateDownloading, TargetVersion: release.TagName}
	}
	s.mu.Unlock()

	switch {
	case !found:
		return echo.NewHTTPError(http.StatusNotFound, "no release "+in.Version)
	case !release.installable():
		return echo.NewHTTPError(http.StatusBadRequest, "release "+in.Version+" has no build for this platform")
	case busy:
		return echo.NewHTTPError(http.StatusConflict, "an update is already in progress")
	}

	s.broadcast()
	go s.runInstall(release)
	return c.JSON(http.StatusAccepted, s.snapshot())
}
