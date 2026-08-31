package main

import (
	"testing"
	"time"
)

func TestParseCodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		want     map[string]string
		consumed int
	}{
		{"empty input", nil, nil, 0},
		{"no leading fence", []string{"# Title"}, nil, 0},
		{"unterminated block is ignored", []string{"```", "A: 1"}, nil, 0},
		{"empty block", []string{"```", "```"}, map[string]string{}, 2},
		{
			"single pair",
			[]string{"```", "Project Name: Venice", "```"},
			map[string]string{"Project Name": "Venice"}, 3,
		},
		{
			"two pairs",
			[]string{"```", "A: 1", "B: 2", "```"},
			map[string]string{"A": "1", "B": "2"}, 4,
		},
		{
			"colons in the value are preserved",
			[]string{"```", "Note: a: b: c", "```"},
			map[string]string{"Note": "a: b: c"}, 3,
		},
		{
			"whitespace around keys and values is trimmed",
			[]string{"```", "  Key  :   value  ", "```"},
			map[string]string{"Key": "value"}, 3,
		},
		{
			"lines without a colon are skipped",
			[]string{"```", "no colon here", "A: 1", "```"},
			map[string]string{"A": "1"}, 4,
		},
		{
			"trailing lines are not consumed",
			[]string{"```", "A: 1", "```", "# Later"},
			map[string]string{"A": "1"}, 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, consumed := parseCodeBlock(tt.lines)
			if consumed != tt.consumed {
				t.Errorf("consumed = %d, want %d", consumed, tt.consumed)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("key %q = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestParseItemsMilestoneTitles(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		title    string
		duration int
	}{
		{"no duration defaults to one week", "# Proto Map", "Proto Map", 1},
		{"duration suffix is parsed and stripped", "# Proto Map (3)", "Proto Map", 3},
		{"multi-digit duration", "# Long Haul (12)", "Long Haul", 12},
		{"trailing space after duration", "# Proto Map (3)  ", "Proto Map", 3},
		{"parens mid-title are left alone", "# Ship (beta) work", "Ship (beta) work", 1},
		{"non-numeric parens are left alone", "# Ship (soon)", "Ship (soon)", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := parseItems([]string{tt.line})
			if len(items) != 1 {
				t.Fatalf("got %d items, want 1", len(items))
			}
			if items[0].title != tt.title {
				t.Errorf("title = %q, want %q", items[0].title, tt.title)
			}
			if items[0].duration != tt.duration {
				t.Errorf("duration = %d, want %d", items[0].duration, tt.duration)
			}
		})
	}
}

func TestParseItemsTasksAndSubtasks(t *testing.T) {
	items := parseItems([]string{
		"# Milestone",
		"- [ ] open task",
		"- [x] done task",
		"  - [ ] open subtask",
		"  - [x] done subtask",
		"- [ ] task with no subtasks",
	})
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	tasks := items[0].tasks
	if len(tasks) != 3 {
		t.Fatalf("got %d tasks, want 3", len(tasks))
	}
	if tasks[0].title != "open task" || tasks[0].completed {
		t.Errorf("task 0 = %+v", tasks[0])
	}
	if tasks[1].title != "done task" || !tasks[1].completed {
		t.Errorf("task 1 = %+v", tasks[1])
	}
	if len(tasks[1].subtasks) != 2 {
		t.Fatalf("task 1 got %d subtasks, want 2", len(tasks[1].subtasks))
	}
	if tasks[1].subtasks[0].title != "open subtask" || tasks[1].subtasks[0].completed {
		t.Errorf("subtask 0 = %+v", tasks[1].subtasks[0])
	}
	if !tasks[1].subtasks[1].completed {
		t.Errorf("subtask 1 should be completed")
	}
	if len(tasks[2].subtasks) != 0 {
		t.Errorf("task 2 should have no subtasks")
	}
}

func TestParseItemsSubtaskIndentation(t *testing.T) {
	tests := []struct {
		name string
		line string
		want int
	}{
		{"two spaces", "  - [ ] sub", 1},
		{"four spaces", "    - [ ] sub", 1},
		{"tab", "\t- [ ] sub", 1},
		{"unindented becomes its own task", "- [ ] sub", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := parseItems([]string{"# M", "- [ ] parent", tt.line})
			if got := len(items[0].tasks[0].subtasks); got != tt.want {
				t.Errorf("got %d subtasks, want %d", got, tt.want)
			}
		})
	}
}

func TestParseItemsDescription(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{"single line", []string{"# M", "some text"}, "some text"},
		{"leading blanks are skipped", []string{"# M", "", "", "some text"}, "some text"},
		{"multi-line is joined", []string{"# M", "line one", "line two"}, "line one\nline two"},
		{"trailing blanks are trimmed", []string{"# M", "text", "", ""}, "text"},
		{"text after tasks is dropped", []string{"# M", "- [ ] t", "stray text"}, ""},
		{
			"blockquote metadata before description is skipped",
			[]string{"# M", "> generated: week 3", "real description"},
			"real description",
		},
		{
			"blockquote after description is kept",
			[]string{"# M", "real description", "> quoted line"},
			"real description\n> quoted line",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := parseItems(tt.lines)
			if got := items[0].description; got != tt.want {
				t.Errorf("description = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseItemsFinishedDate(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  string // empty means zero time
	}{
		{"valid finished date", []string{"# M", "```", "Finished: May 4 2026", "```"}, "May 4 2026"},
		{"malformed date leaves zero", []string{"# M", "```", "Finished: nonsense", "```"}, ""},
		{"no metadata block leaves zero", []string{"# M"}, ""},
		{"unrelated key leaves zero", []string{"# M", "```", "Other: May 4 2026", "```"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := parseItems(tt.lines)
			got := items[0].finished
			if tt.want == "" {
				if !got.IsZero() {
					t.Errorf("finished = %v, want zero", got)
				}
				return
			}
			want, err := time.Parse(dateFormat, tt.want)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Equal(want) {
				t.Errorf("finished = %v, want %v", got, want)
			}
		})
	}
}

func TestParseItemsMultipleMilestones(t *testing.T) {
	items := parseItems([]string{
		"# First (2)",
		"desc one",
		"- [x] a",
		"# Second",
		"desc two",
	})
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].title != "First" || items[0].duration != 2 || items[0].description != "desc one" {
		t.Errorf("item 0 = %+v", items[0])
	}
	if len(items[0].tasks) != 1 {
		t.Errorf("item 0 should keep its task")
	}
	if items[1].title != "Second" || items[1].description != "desc two" {
		t.Errorf("item 1 = %+v", items[1])
	}
}

func TestParseItemsIgnoresContentBeforeFirstHeader(t *testing.T) {
	items := parseItems([]string{"stray text", "- [ ] orphan task", "# Real"})
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].title != "Real" || len(items[0].tasks) != 0 {
		t.Errorf("item = %+v", items[0])
	}
}

func TestParseProject(t *testing.T) {
	t.Run("metadata block populates name and start date", func(t *testing.T) {
		var p project
		parseProject("```\nProject Name: Venice\nProject Start: Apr 28 2026\n```\n# M\n", &p)
		if p.name != "Venice" {
			t.Errorf("name = %q, want %q", p.name, "Venice")
		}
		want, _ := time.Parse(dateFormat, "Apr 28 2026")
		if !p.startDate.Equal(want) {
			t.Errorf("startDate = %v, want %v", p.startDate, want)
		}
		if len(p.items) != 1 {
			t.Errorf("got %d items, want 1", len(p.items))
		}
	})

	t.Run("no metadata block still parses items", func(t *testing.T) {
		var p project
		parseProject("# M\n", &p)
		if p.name != "" || !p.startDate.IsZero() {
			t.Errorf("expected empty metadata, got name=%q start=%v", p.name, p.startDate)
		}
		if len(p.items) != 1 {
			t.Errorf("got %d items, want 1", len(p.items))
		}
	})

	t.Run("malformed start date falls through to zero", func(t *testing.T) {
		var p project
		parseProject("```\nProject Start: not a date\n```\n", &p)
		if !p.startDate.IsZero() {
			t.Errorf("startDate = %v, want zero", p.startDate)
		}
	})

	t.Run("empty content yields no items", func(t *testing.T) {
		var p project
		parseProject("", &p)
		if len(p.items) != 0 {
			t.Errorf("got %d items, want 0", len(p.items))
		}
	})
}

func TestReadWriteDateRoundTrip(t *testing.T) {
	const s = "Jan 2 2026"
	d, err := readDate(s)
	if err != nil {
		t.Fatal(err)
	}
	if got := writeDate(d); got != s {
		t.Errorf("writeDate(readDate(%q)) = %q", s, got)
	}
	if _, err := readDate("garbage"); err == nil {
		t.Error("readDate accepted garbage")
	}
}
