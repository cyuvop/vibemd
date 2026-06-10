package main

import "testing"

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
