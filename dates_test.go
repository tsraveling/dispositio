package main

import "testing"

func TestWeekdaysBetween(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
		want int
	}{
		{"same weekday counts one", "Jun 1 2026", "Jun 1 2026", 1},
		{"same weekend day counts none", "Jun 6 2026", "Jun 6 2026", 0},
		{"reversed range counts none", "Jun 5 2026", "Jun 1 2026", 0},
		{"monday to friday", "Jun 1 2026", "Jun 5 2026", 5},
		{"full week skips the weekend", "Jun 1 2026", "Jun 7 2026", 5},
		{"across a weekend", "Jun 5 2026", "Jun 8 2026", 2},
		{"two full weeks", "Jun 1 2026", "Jun 14 2026", 10},
		{"weekend only", "Jun 6 2026", "Jun 7 2026", 0},
		{"across a month boundary", "Jun 29 2026", "Jul 3 2026", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := weekdaysBetween(mustDate(t, tt.from), mustDate(t, tt.to))
			if got != tt.want {
				t.Errorf("weekdaysBetween(%s, %s) = %d, want %d", tt.from, tt.to, got, tt.want)
			}
		})
	}
}
