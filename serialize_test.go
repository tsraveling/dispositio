package main

import (
	"strings"
	"testing"
	"time"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse(dateFormat, s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestWriteCodeBlock(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		m    map[string]string
		want string
	}{
		{"empty", nil, nil, "```\n```\n"},
		{"single pair", []string{"A"}, map[string]string{"A": "1"}, "```\nA: 1\n```\n"},
		{
			"key order is respected, not map order",
			[]string{"B", "A"}, map[string]string{"A": "1", "B": "2"},
			"```\nB: 2\nA: 1\n```\n",
		},
		{
			"keys missing from the map are skipped",
			[]string{"A", "Absent"}, map[string]string{"A": "1"},
			"```\nA: 1\n```\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := writeCodeBlock(tt.keys, tt.m); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderProjectMetadataBlock(t *testing.T) {
	tests := []struct {
		name    string
		p       project
		wantHas []string
		wantNot []string
	}{
		{
			"no name and no start date writes no block",
			project{},
			nil, []string{"```"},
		},
		{
			"name only",
			project{name: "Venice"},
			[]string{"Project Name: Venice"}, []string{"Project Start"},
		},
		{
			"name and start date",
			project{name: "Venice", startDate: mustDate(t, "Apr 28 2026")},
			[]string{"Project Name: Venice", "Project Start: Apr 28 2026"}, nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderProject(tt.p)
			for _, s := range tt.wantHas {
				if !strings.Contains(got, s) {
					t.Errorf("output missing %q:\n%s", s, got)
				}
			}
			for _, s := range tt.wantNot {
				if strings.Contains(got, s) {
					t.Errorf("output unexpectedly contains %q:\n%s", s, got)
				}
			}
		})
	}
}

func TestRenderProjectMilestones(t *testing.T) {
	tests := []struct {
		name string
		item milestone
		want string
	}{
		{
			"duration of one is not written back",
			milestone{title: "Solo", duration: 1},
			"# Solo\n",
		},
		{
			"duration above one is written as a suffix",
			milestone{title: "Long", duration: 4},
			"# Long (4)\n",
		},
		{
			"finished date becomes a metadata block",
			milestone{title: "Done", duration: 1, finished: mustDate(t, "May 4 2026")},
			"# Done\n```\nFinished: May 4 2026\n```\n",
		},
		{
			"description follows a blank line",
			milestone{title: "M", duration: 1, description: "some text"},
			"# M\n\nsome text\n",
		},
		{
			"tasks and subtasks use checkbox syntax",
			milestone{title: "M", duration: 1, tasks: []task{
				{title: "open", completed: false},
				{title: "done", completed: true, subtasks: []subtask{
					{title: "sub open", completed: false},
					{title: "sub done", completed: true},
				}},
			}},
			"# M\n\n- [ ] open\n- [x] done\n  - [ ] sub open\n  - [x] sub done\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renderProject(project{items: []milestone{tt.item}}); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderProjectSeparatesMilestonesWithBlankLine(t *testing.T) {
	p := project{items: []milestone{
		{title: "First", duration: 1},
		{title: "Second", duration: 1},
	}}
	want := "# First\n\n# Second\n"
	if got := renderProject(p); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSaveProjectWritesRenderedOutput(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/ROADMAP.md"
	p := project{
		filePath:  path,
		name:      "Venice",
		startDate: mustDate(t, "Apr 28 2026"),
		items:     []milestone{{title: "M", duration: 2}},
	}
	if err := p.save(); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadProject(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.name != p.name {
		t.Errorf("name = %q, want %q", loaded.name, p.name)
	}
	if !loaded.startDate.Equal(p.startDate) {
		t.Errorf("startDate = %v, want %v", loaded.startDate, p.startDate)
	}
	if len(loaded.items) != 1 || loaded.items[0].duration != 2 {
		t.Errorf("items = %+v", loaded.items)
	}
}

func TestLoadProjectMissingFile(t *testing.T) {
	if _, err := loadProject(t.TempDir() + "/nope.md"); err == nil {
		t.Error("expected an error for a missing file")
	}
}

// a file with no Project Start falls back to today, so the timeline has a base
func TestLoadProjectDefaultsStartDateToNow(t *testing.T) {
	fixed := mustDate(t, "Jun 1 2026")
	defer stubNow(t, fixed)()

	dir := t.TempDir()
	path := dir + "/ROADMAP.md"
	if err := writeFile(path, "# M\n"); err != nil {
		t.Fatal(err)
	}
	p, err := loadProject(path)
	if err != nil {
		t.Fatal(err)
	}
	if !p.startDate.Equal(fixed) {
		t.Errorf("startDate = %v, want %v", p.startDate, fixed)
	}
}
