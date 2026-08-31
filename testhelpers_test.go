package main

import (
	"os"
	"testing"
	"time"
)

// pins the package clock for the duration of a test; call the returned func to
// restore it.
func stubNow(t *testing.T, at time.Time) func() {
	t.Helper()
	prev := now
	now = func() time.Time { return at }
	return func() { now = prev }
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
