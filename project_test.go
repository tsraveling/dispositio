package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseItems(t *testing.T) {
	data, err := os.ReadFile("ROADMAP.md")
	if err != nil {
		t.Fatalf("failed to read ROADMAP.md: %v", err)
	}
	items := parseItems(string(data))

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// First item
	it := items[0]
	if it.title != "First Task" {
		t.Errorf("item 0 title: got %q, want %q", it.title, "First Task")
	}
	if it.duration != 3 {
		t.Errorf("item 0 duration: got %d, want 3", it.duration)
	}
	if it.description != "This is the first task we're going to tackle." {
		t.Errorf("item 0 description: got %q", it.description)
	}
	if len(it.subtasks) != 2 {
		t.Fatalf("item 0 subtasks: got %d, want 2", len(it.subtasks))
	}
	if it.subtasks[0].title != "First subtask" {
		t.Errorf("subtask 0 title: got %q", it.subtasks[0].title)
	}
	if it.subtasks[0].description != "This is the description of the first subtask" {
		t.Errorf("subtask 0 description: got %q", it.subtasks[0].description)
	}
	if it.subtasks[1].title != "Second subtask" {
		t.Errorf("subtask 1 title: got %q", it.subtasks[1].title)
	}
	if it.subtasks[1].description != "" {
		t.Errorf("subtask 1 description: got %q, want empty", it.subtasks[1].description)
	}

	// Second item
	it = items[1]
	if it.title != "Second Task" {
		t.Errorf("item 1 title: got %q", it.title)
	}
	if it.duration != 2 {
		t.Errorf("item 1 duration: got %d, want 2", it.duration)
	}
	if it.description != "This is the second roadmap item." {
		t.Errorf("item 1 description: got %q", it.description)
	}

	// Third item
	it = items[2]
	if it.title != "Third Task, without description" {
		t.Errorf("item 2 title: got %q", it.title)
	}
	if it.duration != 1 {
		t.Errorf("item 2 duration: got %d, want 1", it.duration)
	}
	if it.description != "" {
		t.Errorf("item 2 description: got %q, want empty", it.description)
	}
	if len(it.subtasks) != 0 {
		t.Errorf("item 2 subtasks: got %d, want 0", len(it.subtasks))
	}
}

func TestSaveProject(t *testing.T) {
	data, err := os.ReadFile("ROADMAP.md")
	if err != nil {
		t.Fatalf("failed to read ROADMAP.md: %v", err)
	}
	items := parseItems(string(data))

	// Save to a temp file
	tmp := filepath.Join(t.TempDir(), "out.md")
	p := project{filePath: tmp, items: items, usesTimeline: true}
	if err := saveProject(p); err != nil {
		t.Fatalf("saveProject: %v", err)
	}

	// Re-parse the saved file and verify round-trip
	saved, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("read saved: %v", err)
	}
	items2 := parseItems(string(saved))

	if len(items2) != len(items) {
		t.Fatalf("round-trip item count: got %d, want %d", len(items2), len(items))
	}

	for i := range items {
		if items2[i].title != items[i].title {
			t.Errorf("item %d title: got %q, want %q", i, items2[i].title, items[i].title)
		}
		if items2[i].duration != items[i].duration {
			t.Errorf("item %d duration: got %d, want %d", i, items2[i].duration, items[i].duration)
		}
		if items2[i].description != items[i].description {
			t.Errorf("item %d description: got %q, want %q", i, items2[i].description, items[i].description)
		}
		if len(items2[i].subtasks) != len(items[i].subtasks) {
			t.Errorf("item %d subtask count: got %d, want %d", i, len(items2[i].subtasks), len(items[i].subtasks))
			continue
		}
		for j := range items[i].subtasks {
			if items2[i].subtasks[j].title != items[i].subtasks[j].title {
				t.Errorf("item %d subtask %d title: got %q, want %q", i, j, items2[i].subtasks[j].title, items[i].subtasks[j].title)
			}
			if items2[i].subtasks[j].description != items[i].subtasks[j].description {
				t.Errorf("item %d subtask %d description: got %q, want %q", i, j, items2[i].subtasks[j].description, items[i].subtasks[j].description)
			}
		}
	}
}
