package main

import (
	"errors"
	"path/filepath"
	"testing"
)

// the planner reads window size from the package-level cfg, so tests that care
// about width must set it explicitly.
func withWidth(t *testing.T, w int) {
	t.Helper()
	prev := cfg
	cfg.updateWW(w)
	t.Cleanup(func() { cfg = prev })
}

func projectFixture(t *testing.T) *project {
	t.Helper()
	return &project{
		name:      "Venice",
		startDate: mustDate(t, "Jun 1 2026"),
		items: []milestone{
			{title: "first", duration: 2, finished: mustDate(t, "Jun 8 2026")},
			{title: "second", duration: 1},
			{title: "third", duration: 3},
		},
	}
}

// cursor 0 is the project meta row; items start at 1
func TestPlannerCursorSpace(t *testing.T) {
	tests := []struct {
		name      string
		cursor    int
		wantIndex int
		wantMeta  bool
	}{
		{"meta row", 0, -1, true},
		{"first item", 1, 0, false},
		{"third item", 3, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := plannerViewModel{cursor: tt.cursor}
			if got := m.itemIndex(); got != tt.wantIndex {
				t.Errorf("itemIndex() = %d, want %d", got, tt.wantIndex)
			}
			if got := m.onMeta(); got != tt.wantMeta {
				t.Errorf("onMeta() = %v, want %v", got, tt.wantMeta)
			}
		})
	}
}

// meta controls are only live on the meta row in normal mode
func TestPlannerIsHoveringMeta(t *testing.T) {
	tests := []struct {
		name   string
		cursor int
		mode   plannerMode
		want   bool
	}{
		{"meta row in normal mode", 0, normal, true},
		{"meta row while editing the project name", 0, editingProjectName, false},
		{"meta row while editing a title", 0, editingTitle, false},
		{"item row in normal mode", 1, normal, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := plannerViewModel{cursor: tt.cursor, mode: tt.mode}
			if got := m.isHoveringMeta(); got != tt.want {
				t.Errorf("isHoveringMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}

// the cursor opens on the current item, or the last one when all are finished
func TestMakePlannerViewModelInitialCursor(t *testing.T) {
	done := mustDate(t, "Jun 8 2026")

	tests := []struct {
		name  string
		items []milestone
		want  int
	}{
		{"empty project sits on meta", nil, 0},
		{"first unfinished item", []milestone{{}, {}}, 1},
		{"skips finished items", []milestone{{finished: done}, {finished: done}, {}}, 3},
		{"all finished lands on the last", []milestone{{finished: done}, {finished: done}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &project{items: tt.items}
			m, cmd := makePlannerViewModel(p)
			if m.cursor != tt.want {
				t.Errorf("cursor = %d, want %d", m.cursor, tt.want)
			}
			if cmd != nil {
				t.Errorf("Init returned %v, want nil", cmd)
			}
			if m.mode != normal {
				t.Errorf("mode = %v, want normal", m.mode)
			}
		})
	}
}

// The project is copied by value, but that is a shallow copy: the items slice
// still shares its backing array with the caller's project, so item edits are
// visible through the original pointer. Harmless today because main() drops
// that pointer immediately, but this test pins the real behavior so a future
// "discard changes" path does not quietly rely on isolation it does not have.
// See _spec/surfaced-bugs.md.
func TestMakePlannerViewModelSharesItemStorage(t *testing.T) {
	p := projectFixture(t)
	m, _ := makePlannerViewModel(p)

	m.prj.name = "renamed project"
	if p.name != "Venice" {
		t.Errorf("scalar field leaked to the source project: %q", p.name)
	}

	m.prj.items[0].title = "mutated"
	if p.items[0].title != "mutated" {
		t.Errorf("items are expected to share storage; source shows %q", p.items[0].title)
	}
}

func TestDetailPanelWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
		want  int
	}{
		{"narrow uses the full width", 80, 80},
		{"just below the split threshold", minSideBySideWidth - 1, minSideBySideWidth - 1},
		{"at the split threshold takes the larger half", minSideBySideWidth, minSideBySideWidth - minSideBySideWidth/2},
		{"wide takes the larger half", 121, 61},
		{"even width splits evenly", 120, 60},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withWidth(t, tt.width)
			var m plannerViewModel
			if got := m.detailPanelWidth(); got != tt.want {
				t.Errorf("detailPanelWidth() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGotoDetail(t *testing.T) {
	withWidth(t, 120)

	t.Run("opens the detail model for the selected item", func(t *testing.T) {
		m, _ := makePlannerViewModel(projectFixture(t))
		m.cursor = 2 // second item
		m.gotoDetail()

		if m.detail == nil {
			t.Fatal("detail model not created")
		}
		if m.detail.item.title != "second" {
			t.Errorf("detail item = %q, want %q", m.detail.item.title, "second")
		}
		if !m.detail.isCurrent {
			t.Error("second item should be current: the first is finished")
		}
	})

	t.Run("does nothing on the meta row", func(t *testing.T) {
		m, _ := makePlannerViewModel(projectFixture(t))
		m.cursor = 0
		m.gotoDetail()

		if m.detail != nil {
			t.Error("detail model created from the meta row")
		}
	})

	t.Run("detail edits reach the planner's copy of the item", func(t *testing.T) {
		m, _ := makePlannerViewModel(projectFixture(t))
		m.cursor = 3
		m.gotoDetail()
		m.detail.item.title = "renamed"

		if m.prj.items[2].title != "renamed" {
			t.Errorf("planner item = %q, want %q", m.prj.items[2].title, "renamed")
		}
	})
}

func TestInsertItemAt(t *testing.T) {
	tests := []struct {
		name       string
		idx        int
		wantTitles []string
		wantCursor int
	}{
		{"at the start", 0, []string{"", "first", "second", "third"}, 1},
		{"in the middle", 1, []string{"first", "", "second", "third"}, 2},
		{"at the end", 3, []string{"first", "second", "third", ""}, 4},
		{"past the end clamps", 99, []string{"first", "second", "third", ""}, 4},
		{"negative clamps to the start", -5, []string{"", "first", "second", "third"}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := makePlannerViewModel(projectFixture(t))
			cmd := m.insertItemAt(tt.idx)

			if len(m.prj.items) != len(tt.wantTitles) {
				t.Fatalf("got %d items, want %d", len(m.prj.items), len(tt.wantTitles))
			}
			for i, want := range tt.wantTitles {
				if m.prj.items[i].title != want {
					t.Errorf("item %d = %q, want %q", i, m.prj.items[i].title, want)
				}
			}
			if m.cursor != tt.wantCursor {
				t.Errorf("cursor = %d, want %d", m.cursor, tt.wantCursor)
			}
			if m.mode != editingTitle || !m.isNewItem {
				t.Errorf("expected new-item editing, got mode=%v isNew=%v", m.mode, m.isNewItem)
			}
			if m.input.Value() != "" || !m.input.Focused() {
				t.Error("input should be empty and focused")
			}
			if cmd == nil {
				t.Error("expected a blink command")
			}
		})
	}
}

// a new item defaults to a one-week duration
func TestInsertItemDefaultDuration(t *testing.T) {
	m, _ := makePlannerViewModel(projectFixture(t))
	m.insertItemAt(1)
	if got := m.prj.items[1].duration; got != 1 {
		t.Errorf("duration = %d, want 1", got)
	}
}

func TestInsertItemIntoEmptyProject(t *testing.T) {
	m, _ := makePlannerViewModel(&project{})
	m.insertItemAt(5) // index is ignored when there is nothing to insert into

	if len(m.prj.items) != 1 {
		t.Fatalf("got %d items, want 1", len(m.prj.items))
	}
	if m.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.cursor)
	}
	if m.prj.items[0].duration != 1 {
		t.Errorf("duration = %d, want 1", m.prj.items[0].duration)
	}
}

// a failing save must be recorded rather than dropped or panicked on: the file
// is the only copy of the user's work
func TestPersistRecordsSaveFailure(t *testing.T) {
	m, _ := makePlannerViewModel(projectFixture(t))

	m.prj.filePath = filepath.Join(t.TempDir(), "ROADMAP.md")
	m.persist()
	if m.saveErr != nil {
		t.Fatalf("valid path should save cleanly, got %v", m.saveErr)
	}

	// a path whose parent is a file, not a directory, cannot be written
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := writeFile(blocker, "x"); err != nil {
		t.Fatal(err)
	}
	m.prj.filePath = filepath.Join(blocker, "ROADMAP.md")
	m.persist()
	if m.saveErr == nil {
		t.Error("expected a save error for an unwritable path")
	}
}

// a later successful save clears a previously recorded failure
func TestPersistClearsPreviousError(t *testing.T) {
	m, _ := makePlannerViewModel(projectFixture(t))
	m.saveErr = errors.New("stale failure")

	m.prj.filePath = filepath.Join(t.TempDir(), "ROADMAP.md")
	m.persist()
	if m.saveErr != nil {
		t.Errorf("successful save left an error: %v", m.saveErr)
	}
}
