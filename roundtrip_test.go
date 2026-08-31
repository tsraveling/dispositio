package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata golden files")

// Parsing then rendering a fixture must produce its golden file. Parse and
// render are deliberately not symmetric — a `(1)` duration, blockquote
// metadata, blank-line spacing and text after tasks are all dropped — so the
// golden files record what dispositio actually writes back, not the input.
func TestRoundTripGolden(t *testing.T) {
	fixtures, err := filepath.Glob("testdata/*.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no fixtures found in testdata/")
	}

	for _, path := range fixtures {
		if strings.HasSuffix(path, ".golden.md") {
			continue
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			var p project
			parseProject(string(src), &p)
			got := renderProject(p)

			golden := strings.TrimSuffix(path, ".md") + ".golden.md"
			if *update {
				if err := os.WriteFile(golden, []byte(got), 0644); err != nil {
					t.Fatal(err)
				}
				t.Logf("updated %s", golden)
				return
			}

			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("%v (run: go test -update)", err)
			}
			if got != string(want) {
				t.Errorf("round trip mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
			}
		})
	}
}

// A second round trip must change nothing. Parse and render are lossy on the
// first pass, but the output they settle on has to be a fixed point, or a file
// would drift a little every time it was opened and saved.
func TestRoundTripIsIdempotent(t *testing.T) {
	fixtures, err := filepath.Glob("testdata/*.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range fixtures {
		if strings.HasSuffix(path, ".golden.md") {
			continue
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			var first project
			parseProject(string(src), &first)
			once := renderProject(first)

			var second project
			parseProject(once, &second)
			twice := renderProject(second)

			if once != twice {
				t.Errorf("second round trip differs\n--- once ---\n%s\n--- twice ---\n%s", once, twice)
			}
		})
	}
}

// Round-tripping must preserve the data the user actually typed, whatever the
// formatting does.
func TestRoundTripPreservesContent(t *testing.T) {
	src, err := os.ReadFile("testdata/finished.md")
	if err != nil {
		t.Fatal(err)
	}

	var before project
	parseProject(string(src), &before)

	var after project
	parseProject(renderProject(before), &after)

	if after.name != before.name {
		t.Errorf("name changed: %q -> %q", before.name, after.name)
	}
	if !after.startDate.Equal(before.startDate) {
		t.Errorf("startDate changed: %v -> %v", before.startDate, after.startDate)
	}
	if len(after.items) != len(before.items) {
		t.Fatalf("item count changed: %d -> %d", len(before.items), len(after.items))
	}
	for i := range before.items {
		b, a := before.items[i], after.items[i]
		if a.title != b.title {
			t.Errorf("item %d title: %q -> %q", i, b.title, a.title)
		}
		if a.duration != b.duration {
			t.Errorf("item %d duration: %d -> %d", i, b.duration, a.duration)
		}
		if !a.finished.Equal(b.finished) {
			t.Errorf("item %d finished: %v -> %v", i, b.finished, a.finished)
		}
		if a.description != b.description {
			t.Errorf("item %d description: %q -> %q", i, b.description, a.description)
		}
		if len(a.tasks) != len(b.tasks) {
			t.Fatalf("item %d task count: %d -> %d", i, len(b.tasks), len(a.tasks))
		}
		for j := range b.tasks {
			bt, at := b.tasks[j], a.tasks[j]
			if at.title != bt.title || at.completed != bt.completed {
				t.Errorf("item %d task %d: %+v -> %+v", i, j, bt, at)
			}
			if len(at.subtasks) != len(bt.subtasks) {
				t.Errorf("item %d task %d subtask count: %d -> %d", i, j, len(bt.subtasks), len(at.subtasks))
			}
		}
	}
}
