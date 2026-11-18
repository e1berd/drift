package delivery

import "testing"

func TestNewerThan(t *testing.T) {
	cases := []struct {
		candidate, current string
		want               bool
	}{
		{"v1.2.0", "v1.1.9", true},
		{"v1.2.0", "v1.2.0", false},
		{"v1.1.0", "v1.2.0", false},
		{"v1.2.0", "v1.2.0-3-gabc1234", false},
		{"v1.2.1", "v1.2.0-3-gabc1234-dirty", true},
		{"v1.2.0", "v1.2.0-rc.1", true},
		{"v1.0.0", "ea02852", true},
		{"v1.0.0", developmentVersion, false},
	}
	for _, tc := range cases {
		if got := newerThan(tc.candidate, tc.current); got != tc.want {
			t.Errorf("newerThan(%q, %q) = %v, want %v", tc.candidate, tc.current, got, tc.want)
		}
	}
}

func TestIsReleaseBuild(t *testing.T) {
	for v, want := range map[string]bool{
		"v1.2.0":           true,
		"v1.2.0-rc.1":      true,
		"v1.2.0-3-gabc123": false,
		"v1.2.0-dirty":     false,
		"ea02852":          false,
		developmentVersion: false,
	} {
		if got := isReleaseBuild(v); got != want {
			t.Errorf("isReleaseBuild(%q) = %v, want %v", v, got, want)
		}
	}
}
