// Package delivery keeps Drift up to date from GitHub Releases: it polls the repository for
// new releases, reports the running version, and installs a chosen release by swapping the
// binary in place and relaunching the process.
package delivery

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/e1berd/drift/app/internal/events"
	"golang.org/x/mod/semver"
)

const (
	eventType       = "delivery.updated"
	checkInterval   = time.Hour
	firstCheckDelay = 10 * time.Second
	minCheckGap     = time.Minute
	installTimeout  = 15 * time.Minute
	restartGrace    = time.Second
)

// Options configures where releases come from and how the process is stopped for a relaunch.
type Options struct {
	Repository string
	APIURL     *url.URL
	Shutdown   context.CancelFunc
}

// Service tracks the releases of the repository and the state of an update in progress.
type Service struct {
	github    github
	installer installer
	hub       *events.Hub
	log       *slog.Logger
	shutdown  context.CancelFunc

	mu           sync.Mutex
	repository   githubRepository
	releases     []githubRelease
	checkTime    time.Time
	installation Installation
	relaunchPath string
}

// New returns the service; the path of the running binary is resolved now, before any update
// can replace it.
func New(opts Options, hub *events.Hub, log *slog.Logger) (*Service, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: installTimeout}
	return &Service{
		github:       github{api: opts.APIURL, repo: opts.Repository, http: &http.Client{Timeout: 30 * time.Second}},
		installer:    installer{http: client, executable: executable},
		hub:          hub,
		log:          log,
		shutdown:     opts.Shutdown,
		repository:   githubRepository{HTMLURL: "https://github.com/" + opts.Repository},
		installation: Installation{State: StateIdle},
	}, nil
}

// Watch checks for releases shortly after start and then every hour until ctx ends.
func (s *Service) Watch(ctx context.Context) {
	timer := time.NewTimer(firstCheckDelay)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			s.check(ctx)
			timer.Reset(checkInterval)
		}
	}
}

// RelaunchPath is the binary to exec after shutdown, or "" when no update was installed.
func (s *Service) RelaunchPath() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.relaunchPath
}

func (s *Service) check(ctx context.Context) {
	if err := s.refresh(ctx); err != nil {
		s.log.Warn("check releases", "err", err)
	}
}

func (s *Service) refresh(ctx context.Context) error {
	releases, err := s.github.releases(ctx)
	if err != nil {
		return err
	}
	repository, err := s.github.repository(ctx)
	if err != nil {
		return err
	}
	s.mu.Lock()
	previous := s.latestLocked()
	s.releases = releases
	s.repository = repository
	s.checkTime = time.Now()
	latest := s.latestLocked()
	s.mu.Unlock()
	if latest != nil && (previous == nil || previous.TagName != latest.TagName) {
		s.broadcast()
	}
	return nil
}

func (s *Service) refreshIfStale(ctx context.Context) error {
	s.mu.Lock()
	fresh := time.Since(s.checkTime) < minCheckGap
	s.mu.Unlock()
	if fresh {
		return nil
	}
	return s.refresh(ctx)
}

func (s *Service) latestLocked() *githubRelease {
	var latest *githubRelease
	for index := range s.releases {
		release := &s.releases[index]
		if release.Draft || release.Prerelease || !semver.IsValid(release.TagName) || !release.installable() {
			continue
		}
		if latest == nil || semver.Compare(release.TagName, latest.TagName) > 0 {
			latest = release
		}
	}
	return latest
}

func (s *Service) releaseLocked(tag string) (githubRelease, bool) {
	for _, release := range s.releases {
		if release.TagName == tag && !release.Draft {
			return release, true
		}
	}
	return githubRelease{}, false
}

func (s *Service) broadcast() {
	s.hub.PublishAll(eventType, s.snapshot())
}

func (s *Service) setInstallation(installation Installation) {
	s.mu.Lock()
	s.installation = installation
	s.mu.Unlock()
	s.broadcast()
}

func (s *Service) runInstall(release githubRelease) {
	ctx, cancel := context.WithTimeout(context.Background(), installTimeout)
	defer cancel()

	target := release.TagName
	report := func(downloaded, total int64) {
		s.setInstallation(Installation{State: StateDownloading, TargetVersion: target, DownloadedBytes: downloaded, TotalBytes: total})
	}
	if err := s.installer.install(ctx, release, report); err != nil {
		s.log.Error("install release", "version", target, "err", err)
		s.setInstallation(Installation{State: StateFailed, TargetVersion: target, Error: err.Error()})
		return
	}

	s.log.Info("release installed, relaunching", "from", version, "to", target)
	s.mu.Lock()
	s.relaunchPath = s.installer.executable
	s.mu.Unlock()
	s.setInstallation(Installation{State: StateRestarting, TargetVersion: target})
	time.Sleep(restartGrace)
	s.shutdown()
}
