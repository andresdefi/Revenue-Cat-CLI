package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/andresdefi/rc/internal/output"
	"github.com/andresdefi/rc/internal/version"
)

const (
	releaseURL   = "https://api.github.com/repos/andresdefi/Revenue-Cat-CLI/releases/latest"
	checkFile    = "update-check"
	checkTTL     = 24 * time.Hour
	fetchTimeout = 3 * time.Second
)

type cachedCheck struct {
	LatestVersion string `json:"latest_version"`
	CheckedAt     int64  `json:"checked_at"`
}

// CheckAsync runs a non-blocking update check. If a newer version is available,
// it prints a notice to stderr after the command completes. Safe to call from
// Execute() - never blocks longer than fetchTimeout.
func CheckAsync(done chan<- string) {
	go func() {
		defer close(done)
		msg := check()
		if msg != "" {
			done <- msg
		}
	}()
}

// PrintNotice prints the update notice if one was found.
func PrintNotice(done <-chan string) {
	select {
	case msg := <-done:
		if msg != "" && !output.Quiet {
			fmt.Fprintln(os.Stderr, msg)
		}
	case <-time.After(fetchTimeout + time.Second):
	}
}

func check() string {
	if version.Version == "dev" {
		return ""
	}

	cached, err := loadCache()
	if err == nil && time.Since(time.Unix(cached.CheckedAt, 0)) < checkTTL {
		return compareVersions(cached.LatestVersion)
	}

	latest, err := fetchLatest()
	if err != nil {
		return ""
	}

	saveCache(&cachedCheck{
		LatestVersion: latest,
		CheckedAt:     time.Now().Unix(),
	})

	return compareVersions(latest)
}

func compareVersions(latest string) string {
	current := version.Version
	if newerVersionAvailable(current, latest) {
		return fmt.Sprintf("\nA new version of rc is available: %s -> %s\nUpdate with: brew upgrade rc || go install github.com/andresdefi/rc@latest", strings.TrimPrefix(current, "v"), strings.TrimPrefix(latest, "v"))
	}
	return ""
}

func newerVersionAvailable(current, latest string) bool {
	c, ok := parseVersion(current)
	if !ok {
		return false
	}
	l, ok := parseVersion(latest)
	if !ok {
		return false
	}
	return compareParsedVersion(l, c) > 0
}

type parsedVersion struct {
	major      int
	minor      int
	patch      int
	prerelease string
}

func parseVersion(v string) (parsedVersion, bool) {
	v = strings.TrimSpace(strings.TrimPrefix(v, "v"))
	if v == "" {
		return parsedVersion{}, false
	}
	if i := strings.Index(v, "+"); i >= 0 {
		v = v[:i]
	}

	prerelease := ""
	if i := strings.Index(v, "-"); i >= 0 {
		prerelease = v[i+1:]
		v = v[:i]
	}

	parts := strings.Split(v, ".")
	if len(parts) > 3 {
		return parsedVersion{}, false
	}
	nums := [3]int{}
	for i, part := range parts {
		if part == "" {
			return parsedVersion{}, false
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return parsedVersion{}, false
		}
		nums[i] = n
	}
	return parsedVersion{major: nums[0], minor: nums[1], patch: nums[2], prerelease: prerelease}, true
}

func compareParsedVersion(a, b parsedVersion) int {
	for _, pair := range [][2]int{{a.major, b.major}, {a.minor, b.minor}, {a.patch, b.patch}} {
		switch {
		case pair[0] > pair[1]:
			return 1
		case pair[0] < pair[1]:
			return -1
		}
	}
	switch {
	case a.prerelease == b.prerelease:
		return 0
	case a.prerelease == "":
		return 1
	case b.prerelease == "":
		return -1
	case a.prerelease > b.prerelease:
		return 1
	case a.prerelease < b.prerelease:
		return -1
	default:
		return 0
	}
}

func fetchLatest() (string, error) {
	client := &http.Client{Timeout: fetchTimeout}
	resp, err := client.Get(releaseURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github returned %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

func cacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".rc"), nil
}

func loadCache() (*cachedCheck, error) {
	dir, err := cacheDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(dir, checkFile))
	if err != nil {
		return nil, err
	}
	var c cachedCheck
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func saveCache(c *cachedCheck) {
	dir, err := cacheDir()
	if err != nil {
		return
	}
	_ = os.MkdirAll(dir, 0o700)
	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(dir, checkFile), data, 0o600)
}
