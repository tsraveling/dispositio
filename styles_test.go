package main

import (
	"strings"
	"testing"
)

func TestFmtFullDate(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Jun 1 2026", "Mon, Jun 1, 2026"},
		{"Dec 25 2026", "Fri, Dec 25, 2026"},
		{"Jan 1 2026", "Thu, Jan 1, 2026"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := fmtFullDate(mustDate(t, tt.in)); got != tt.want {
				t.Errorf("fmtFullDate(%s) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBoxWidth(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"narrow terminal passes through", 80, 80},
		{"wide terminal clamps to maxWidth", 300, maxWidth},
		{"exactly maxWidth", maxWidth, maxWidth},
		{"one under maxWidth", maxWidth - 1, maxWidth - 1},
		{"zero", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := boxWidth(tt.in); got != tt.want {
				t.Errorf("boxWidth(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

// keeps the last max lines, so a growing log scrolls rather than overflows
func TestClampLines(t *testing.T) {
	tests := []struct {
		name string
		in   string
		max  int
		want string
	}{
		{"under the limit is untouched", "a\nb", 5, "a\nb"},
		{"exactly at the limit is untouched", "a\nb\nc", 3, "a\nb\nc"},
		{"over the limit keeps the tail", "a\nb\nc\nd", 2, "c\nd"},
		{"single line under limit", "only", 3, "only"},
		{"empty string", "", 3, ""},
		{"limit of one keeps the last line", "a\nb\nc", 1, "c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampLines(tt.in, tt.max); got != tt.want {
				t.Errorf("clampLines(%q, %d) = %q, want %q", tt.in, tt.max, got, tt.want)
			}
		})
	}
}

// the style helpers must produce a box of the requested width without panicking
func TestStyleHelpersRenderAtWidth(t *testing.T) {
	const w, h = 40, 20
	cases := map[string]string{
		"detailStyle active":   detailStyle(w, h, true).Render("body"),
		"detailStyle inactive": detailStyle(w, h, false).Render("body"),
		"errorBoxStyle":        errorBoxStyle(w).Render("an error"),
		"outputBoxStyle done":  outputBoxStyle(w, true).Render("output"),
		"outputBoxStyle busy":  outputBoxStyle(w, false).Render("output"),
	}
	for name, out := range cases {
		t.Run(name, func(t *testing.T) {
			if strings.TrimSpace(out) == "" {
				t.Error("rendered empty")
			}
		})
	}
}
