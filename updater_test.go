package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeReleaseServer spins up a local HTTP server that returns a fake
// GitHub releases response. No network required.
func fakeReleaseServer(t *testing.T, tagName, htmlURL string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"tag_name":%q,"html_url":%q}`, tagName, htmlURL)
	}))
}

func TestCheckForUpdate_UpdateAvailable(t *testing.T) {
	srv := fakeReleaseServer(t, "v2.0.0", "https://github.com/cyuvop/vibemd/releases/tag/v2.0.0")
	defer srv.Close()

	info := checkForUpdate(srv.URL, "1.1.0")

	if !info.HasUpdate {
		t.Fatal("expected HasUpdate=true")
	}
	if info.Version != "2.0.0" {
		t.Errorf("Version = %q, want 2.0.0", info.Version)
	}
	if info.URL == "" {
		t.Error("expected non-empty URL")
	}
}

func TestCheckForUpdate_AlreadyCurrent(t *testing.T) {
	srv := fakeReleaseServer(t, "v1.1.0", "https://github.com/cyuvop/vibemd/releases/tag/v1.1.0")
	defer srv.Close()

	info := checkForUpdate(srv.URL, "1.1.0")

	if info.HasUpdate {
		t.Error("expected HasUpdate=false when already on latest")
	}
}

func TestCheckForUpdate_OlderRelease(t *testing.T) {
	srv := fakeReleaseServer(t, "v1.0.0", "https://example.com")
	defer srv.Close()

	info := checkForUpdate(srv.URL, "1.1.0")

	if info.HasUpdate {
		t.Error("expected HasUpdate=false when release is older than current")
	}
}

func TestCheckForUpdate_NetworkError(t *testing.T) {
	// Point at a port nothing is listening on
	info := checkForUpdate("http://127.0.0.1:1", "1.0.0")
	if info.HasUpdate {
		t.Error("expected HasUpdate=false on network error")
	}
}

func TestCheckForUpdate_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `not json at all`)
	}))
	defer srv.Close()

	info := checkForUpdate(srv.URL, "1.0.0")
	if info.HasUpdate {
		t.Error("expected HasUpdate=false on malformed JSON")
	}
}

func TestIsNewerVersion(t *testing.T) {
	cases := []struct {
		candidate, current string
		want               bool
	}{
		{"1.2.0", "1.1.0", true},
		{"1.1.1", "1.1.0", true},
		{"2.0.0", "1.9.9", true},
		{"1.1.0", "1.1.0", false}, // same version
		{"1.0.0", "1.1.0", false}, // older
		{"1.1.0", "1.2.0", false}, // older minor
		{"0.9.9", "1.0.0", false}, // older major
	}
	for _, c := range cases {
		got := isNewerVersion(c.candidate, c.current)
		if got != c.want {
			t.Errorf("isNewerVersion(%q, %q) = %v, want %v",
				c.candidate, c.current, got, c.want)
		}
	}
}

func TestParseSemver(t *testing.T) {
	cases := []struct {
		in   string
		want []int
	}{
		{"1.2.3", []int{1, 2, 3}},
		{"10.0.0", []int{10, 0, 0}},
		{"1.1.0", []int{1, 1, 0}},
	}
	for _, c := range cases {
		got := parseSemver(c.in)
		if len(got) != len(c.want) {
			t.Errorf("parseSemver(%q) len = %d, want %d", c.in, len(got), len(c.want))
			continue
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("parseSemver(%q)[%d] = %d, want %d", c.in, i, got[i], c.want[i])
			}
		}
	}
}
