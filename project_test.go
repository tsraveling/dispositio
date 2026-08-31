package main

import (
	"testing"
	"time"
)

func TestTaskSubtaskCounts(t *testing.T) {
	tests := []struct {
		name       string
		subtasks   []subtask
		allDone    bool
		completed  int
		incomplete int
	}{
		{"no subtasks is trivially done", nil, true, 0, 0},
		{"one open", []subtask{{completed: false}}, false, 0, 1},
		{"one done", []subtask{{completed: true}}, true, 1, 0},
		{
			"mixed",
			[]subtask{{completed: true}, {completed: false}, {completed: true}},
			false, 2, 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := task{subtasks: tt.subtasks}
			if got := task.allSubtasksDone(); got != tt.allDone {
				t.Errorf("allSubtasksDone = %v, want %v", got, tt.allDone)
			}
			if got := task.completedSubtasks(); got != tt.completed {
				t.Errorf("completedSubtasks = %d, want %d", got, tt.completed)
			}
			if got := task.incompleteSubtasks(); got != tt.incomplete {
				t.Errorf("incompleteSubtasks = %d, want %d", got, tt.incomplete)
			}
		})
	}
}

func TestMilestoneDateString(t *testing.T) {
	tests := []struct {
		name     string
		finished time.Time
		want     string
	}{
		{"unfinished is empty", time.Time{}, ""},
		{"single-digit month and day", mustDate(t, "May 4 2026"), "(5.4)"},
		{"double-digit month and day", mustDate(t, "Dec 25 2026"), "(12.25)"},
		{"january first", mustDate(t, "Jan 1 2026"), "(1.1)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := milestone{finished: tt.finished}
			if got := m.dateString(); got != tt.want {
				t.Errorf("dateString = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMilestoneActualWeeks(t *testing.T) {
	start := mustDate(t, "Jun 1 2026") // a Monday

	tests := []struct {
		name     string
		finished string // empty means in progress
		nowAt    string
		want     int
	}{
		{"finished same day", "Jun 1 2026", "Jun 1 2026", 0},
		{"finished mid-week rounds up", "Jun 3 2026", "Jun 8 2026", 1},
		{"finished after exactly one week", "Jun 8 2026", "Jun 8 2026", 1},
		{"finished after eight days rounds up", "Jun 9 2026", "Jun 9 2026", 2},
		{"in progress uses the clock", "", "Jun 15 2026", 2},
		{"in progress mid-week rounds up", "", "Jun 10 2026", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer stubNow(t, mustDate(t, tt.nowAt))()
			m := milestone{}
			if tt.finished != "" {
				m.finished = mustDate(t, tt.finished)
			}
			if got := m.actualWeeks(start); got != tt.want {
				t.Errorf("actualWeeks = %d, want %d", got, tt.want)
			}
		})
	}
}

// planned duration is a floor: an item that finishes early still occupies its
// planned weeks, and one that overruns extends
func TestMilestoneActualDuration(t *testing.T) {
	start := mustDate(t, "Jun 1 2026")
	defer stubNow(t, mustDate(t, "Jun 1 2026"))()

	tests := []struct {
		name     string
		duration int
		finished string
		want     int
	}{
		{"finished early keeps planned duration", 4, "Jun 8 2026", 4},
		{"finished on time", 1, "Jun 8 2026", 1},
		{"overrun extends past the plan", 1, "Jun 22 2026", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := milestone{duration: tt.duration, finished: mustDate(t, tt.finished)}
			if got := m.actualDuration(start); got != tt.want {
				t.Errorf("actualDuration = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMilestoneWeeksRendered(t *testing.T) {
	start := mustDate(t, "Jun 1 2026") // a Monday
	defer stubNow(t, mustDate(t, "Jun 1 2026"))()

	tests := []struct {
		name     string
		duration int
		finished string
		want     int
	}{
		{"unfinished falls back to planned duration", 3, "", 3},
		{"finished in week one occupies one week", 4, "Jun 1 2026", 1},
		{"finishing on a Sunday still counts that week", 4, "Jun 7 2026", 1},
		{"finishing the next Monday starts week two", 4, "Jun 8 2026", 2},
		{"cannot exceed actual duration", 1, "Jun 22 2026", 3},
		{"finished before it started clamps to one", 4, "May 1 2026", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := milestone{duration: tt.duration}
			if tt.finished != "" {
				m.finished = mustDate(t, tt.finished)
			}
			if got := m.weeksRendered(start); got != tt.want {
				t.Errorf("weeksRendered = %d, want %d", got, tt.want)
			}
		})
	}
}

// the current item is the first one not yet finished
func TestProjectIsCurrent(t *testing.T) {
	done := mustDate(t, "May 4 2026")

	tests := []struct {
		name  string
		items []milestone
		want  int // index expected to be current, -1 for none
	}{
		{"empty project has none", nil, -1},
		{"first item when nothing is finished", []milestone{{}, {}}, 0},
		{"first unfinished after a finished one", []milestone{{finished: done}, {}, {}}, 1},
		{"all finished has none", []milestone{{finished: done}, {finished: done}}, -1},
		{
			"a finished item later in the list does not steal current",
			[]milestone{{}, {finished: done}},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := project{items: tt.items}
			for i := range tt.items {
				if got := p.isCurrent(i); got != (i == tt.want) {
					t.Errorf("isCurrent(%d) = %v, want %v", i, got, i == tt.want)
				}
			}
		})
	}
}

func TestProjectItemStartDate(t *testing.T) {
	defer stubNow(t, mustDate(t, "Jun 1 2026"))()

	t.Run("snaps a mid-week start date back to Monday", func(t *testing.T) {
		// Jun 3 2026 is a Wednesday; the timeline should begin Jun 1
		p := project{startDate: mustDate(t, "Jun 3 2026"), items: []milestone{{duration: 1}}}
		want := mustDate(t, "Jun 1 2026")
		if got := p.itemStartDate(0); !got.Equal(want) {
			t.Errorf("itemStartDate(0) = %v, want %v", got, want)
		}
	})

	t.Run("a Monday start date is left alone", func(t *testing.T) {
		p := project{startDate: mustDate(t, "Jun 1 2026"), items: []milestone{{duration: 1}}}
		want := mustDate(t, "Jun 1 2026")
		if got := p.itemStartDate(0); !got.Equal(want) {
			t.Errorf("itemStartDate(0) = %v, want %v", got, want)
		}
	})

	t.Run("each item starts after the previous one's rendered weeks", func(t *testing.T) {
		p := project{
			startDate: mustDate(t, "Jun 1 2026"),
			items: []milestone{
				{duration: 2},
				{duration: 1},
				{duration: 3},
			},
		}
		wants := []string{"Jun 1 2026", "Jun 15 2026", "Jun 22 2026"}
		for i, w := range wants {
			want := mustDate(t, w)
			if got := p.itemStartDate(i); !got.Equal(want) {
				t.Errorf("itemStartDate(%d) = %v, want %v", i, got, want)
			}
		}
	})

	t.Run("an early finish pulls later items forward", func(t *testing.T) {
		// planned 4 weeks but finished in week one, so the next item starts
		// a week in rather than four
		p := project{
			startDate: mustDate(t, "Jun 1 2026"),
			items: []milestone{
				{duration: 4, finished: mustDate(t, "Jun 3 2026")},
				{duration: 1},
			},
		}
		want := mustDate(t, "Jun 8 2026")
		if got := p.itemStartDate(1); !got.Equal(want) {
			t.Errorf("itemStartDate(1) = %v, want %v", got, want)
		}
	})

	t.Run("an overrun pushes later items back", func(t *testing.T) {
		p := project{
			startDate: mustDate(t, "Jun 1 2026"),
			items: []milestone{
				{duration: 1, finished: mustDate(t, "Jun 22 2026")},
				{duration: 1},
			},
		}
		want := mustDate(t, "Jun 22 2026")
		if got := p.itemStartDate(1); !got.Equal(want) {
			t.Errorf("itemStartDate(1) = %v, want %v", got, want)
		}
	})
}
