package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type UpdateInfo struct {
	HasUpdate bool   `json:"hasUpdate"`
	Version   string `json:"version"`
	URL       string `json:"url"`
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// CheckForUpdate calls the GitHub releases API and returns update info.
// Called from JS on startup; runs with a short timeout so it never hangs the UI.
func (a *App) CheckForUpdate() UpdateInfo {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/cyuvop/vibemd/releases/latest")
	if err != nil {
		return UpdateInfo{}
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return UpdateInfo{}
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	if !isNewerVersion(latest, AppVersion) {
		return UpdateInfo{}
	}

	return UpdateInfo{
		HasUpdate: true,
		Version:   latest,
		URL:       release.HTMLURL,
	}
}

// isNewerVersion returns true if candidate is a higher semver than current.
func isNewerVersion(candidate, current string) bool {
	return compareSemver(candidate, current) > 0
}

func compareSemver(a, b string) int {
	ap := parseSemver(a)
	bp := parseSemver(b)
	for i := range ap {
		if i >= len(bp) {
			return 1
		}
		if ap[i] != bp[i] {
			if ap[i] > bp[i] {
				return 1
			}
			return -1
		}
	}
	if len(bp) > len(ap) {
		return -1
	}
	return 0
}

func parseSemver(v string) []int {
	parts := strings.Split(strings.TrimSpace(v), ".")
	result := make([]int, len(parts))
	for i, p := range parts {
		n, _ := strconv.Atoi(p)
		result[i] = n
	}
	return result
}
