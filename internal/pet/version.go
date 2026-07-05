package pet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Version is the current release version. Overridden at build time via ldflags.
var Version = "0.0.7"

// CheckLatestRelease queries GitHub for the latest release tag and returns it.
// Returns empty string if already on latest or if the check fails.
func CheckLatestRelease() (string, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/jansuthacheeva/ccpetline/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api status %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(Version, "v")

	if compareSemver(latest, current) > 0 {
		return release.TagName, nil
	}
	return "", nil
}

// compareSemver compares two semver strings (without leading "v").
// Returns 1 if a > b, -1 if a < b, 0 if equal. Pre-release suffixes
// ("0.0.8-rc1", "0.0.8+build") are stripped before comparing: Atoi would
// otherwise silently parse them as 0 and misorder against plain releases.
func compareSemver(a, b string) int {
	pa := strings.Split(stripPreRelease(a), ".")
	pb := strings.Split(stripPreRelease(b), ".")
	for i := 0; i < 3; i++ {
		va, vb := 0, 0
		if i < len(pa) {
			va, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			vb, _ = strconv.Atoi(pb[i])
		}
		if va > vb {
			return 1
		}
		if va < vb {
			return -1
		}
	}
	return 0
}

// stripPreRelease drops the pre-release/build suffix from a semver string.
func stripPreRelease(v string) string {
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		return v[:i]
	}
	return v
}
