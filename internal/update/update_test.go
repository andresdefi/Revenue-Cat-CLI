package update

import (
	"testing"

	"github.com/andresdefi/rc/internal/version"
)

func TestNewerVersionAvailable(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{name: "newer patch", current: "v0.2.0", latest: "v0.2.1", want: true},
		{name: "newer minor avoids lexical compare", current: "v0.9.0", latest: "v0.10.0", want: true},
		{name: "older minor", current: "v0.10.0", latest: "v0.9.0", want: false},
		{name: "same version", current: "v0.2.0", latest: "0.2.0", want: false},
		{name: "release newer than prerelease", current: "v1.0.0-rc.1", latest: "v1.0.0", want: true},
		{name: "prerelease older than release", current: "v1.0.0", latest: "v1.0.1-rc.1", want: true},
		{name: "invalid current", current: "dev", latest: "v1.0.0", want: false},
		{name: "invalid latest", current: "v1.0.0", latest: "latest", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newerVersionAvailable(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("newerVersionAvailable(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestCompareVersionsMessage(t *testing.T) {
	versionForTest(t, "v0.9.0")

	got := compareVersions("v0.10.0")
	if got == "" {
		t.Fatal("compareVersions should report v0.10.0 as newer than v0.9.0")
	}
}

func versionForTest(t *testing.T, v string) string {
	t.Helper()
	old := version.Version
	version.Version = v
	t.Cleanup(func() {
		version.Version = old
	})
	return old
}
