package delivery

import (
	"regexp"

	"golang.org/x/mod/semver"
)

const developmentVersion = "dev"

var version = developmentVersion

var gitDescribeSuffix = regexp.MustCompile(`-\d+-g[0-9a-f]+(-dirty)?$|-dirty$`)

// Version is the version the binary was built as: a release tag such as v1.2.0, `git describe`
// output for local builds, or "dev" when built without -ldflags.
func Version() string {
	return version
}

func baseRelease(v string) string {
	return gitDescribeSuffix.ReplaceAllString(v, "")
}

func isReleaseBuild(v string) bool {
	return semver.IsValid(v) && baseRelease(v) == v
}

func newerThan(candidate, current string) bool {
	if current == developmentVersion {
		return false
	}
	base := baseRelease(current)
	if !semver.IsValid(base) {
		return true
	}
	return semver.Compare(candidate, base) > 0
}
