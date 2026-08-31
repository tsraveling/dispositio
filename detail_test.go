package main

import "testing"

// builds a detail model over a milestone, leaving the cursor where the
// constructor put it.
func newDetail(t *testing.T, it *milestone) detailViewModel {
	t.Helper()
	return makeDetailViewModel(it, 80, mustDate(t, "Jun 1 2026"), true)
}

func tasksFixture() *milestone {
	return &milestone{
		title:    "M",
		duration: 1,
		tasks: []task{
			{title: "first", completed: true},
			{title: "second", subtasks: []subtask{
				{title: "sub a", completed: true},
				{title: "sub b"},
			}},
			{title: "third"},
		},
	}
}

// the cursor opens on the first unfinished task, and its first unfinished
// subtask when it has any
func TestCursorToFirstUnchecked(t *testing.T) {
	tests := []struct {
		name     string
		item     *milestone
		wantTask int
		wantSub  int
	}{
		{"no tasks leaves the default", &milestone{}, 0, -1},
		{
			"all complete leaves the default",
			&milestone{tasks: []task{{completed: true}, {completed: true}}},
			0, -1,
		},
		{
			"first open task",
			&milestone{tasks: []task{{completed: true}, {title: "open"}}},
			1, -1,
		},
		{
			"descends into the first open subtask",
			&milestone{tasks: []task{{subtasks: []subtask{
				{completed: true}, {title: "open sub"},
			}}}},
			0, 1,
		},
		{
			"task with all subtasks done stays on the task",
			&milestone{tasks: []task{{subtasks: []subtask{{completed: true}}}}},
			0, -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDetail(t, tt.item)
			if d.taskCursor != tt.wantTask || d.subCursor != tt.wantSub {
				t.Errorf("cursor = (%d, %d), want (%d, %d)",
					d.taskCursor, d.subCursor, tt.wantTask, tt.wantSub)
			}
		})
	}
}

func TestOnTask(t *testing.T) {
	d := newDetail(t, tasksFixture())
	d.subCursor = -1
	if !d.onTask() {
		t.Error("subCursor -1 should be on the task row")
	}
	d.subCursor = 0
	if d.onTask() {
		t.Error("subCursor 0 should be on a subtask row")
	}
}

// only open tasks contribute their subtasks to the flat list
func TestFlatRows(t *testing.T) {
	item := tasksFixture()
	d := newDetail(t, item)

	for i := range item.tasks {
		item.tasks[i].open = false
	}
	if got := len(d.flatRows()); got != 3 {
		t.Errorf("all closed: got %d rows, want 3", got)
	}

	item.tasks[1].open = true
	rows := d.flatRows()
	want := []rowRef{{0, -1}, {1, -1}, {1, 0}, {1, 1}, {2, -1}}
	if len(rows) != len(want) {
		t.Fatalf("got %d rows, want %d", len(rows), len(want))
	}
	for i := range want {
		if rows[i] != want[i] {
			t.Errorf("row %d = %+v, want %+v", i, rows[i], want[i])
		}
	}

	empty := newDetail(t, &milestone{})
	if got := len(empty.flatRows()); got != 0 {
		t.Errorf("no tasks: got %d rows, want 0", got)
	}
}

func TestMoveCursor(t *testing.T) {
	item := tasksFixture()
	item.tasks[1].open = true
	// rows: (0,-1) (1,-1) (1,0) (1,1) (2,-1)

	tests := []struct {
		name                string
		startTask, startSub int
		delta               int
		wantTask, wantSub   int
	}{
		{"down from first task", 0, -1, 1, 1, -1},
		{"down into subtasks", 1, -1, 1, 1, 0},
		{"down between subtasks", 1, 0, 1, 1, 1},
		{"down out of subtasks", 1, 1, 1, 2, -1},
		{"up from last", 2, -1, -1, 1, 1},
		{"clamps at the top", 0, -1, -1, 0, -1},
		{"clamps at the bottom", 2, -1, 1, 2, -1},
		{"large jump clamps to the end", 0, -1, 99, 2, -1},
		{"large negative jump clamps to the start", 2, -1, -99, 0, -1},
		{"zero delta stays put", 1, 0, 0, 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDetail(t, item)
			d.taskCursor, d.subCursor = tt.startTask, tt.startSub
			d.moveCursor(tt.delta)
			if d.taskCursor != tt.wantTask || d.subCursor != tt.wantSub {
				t.Errorf("cursor = (%d, %d), want (%d, %d)",
					d.taskCursor, d.subCursor, tt.wantTask, tt.wantSub)
			}
		})
	}
}

func TestMoveCursorWithNoTasks(t *testing.T) {
	d := newDetail(t, &milestone{})
	d.taskCursor, d.subCursor = 5, 5
	d.moveCursor(1)
	if d.taskCursor != 0 || d.subCursor != -1 {
		t.Errorf("cursor = (%d, %d), want (0, -1)", d.taskCursor, d.subCursor)
	}
}

func TestCurTitleAndSetCurTitle(t *testing.T) {
	item := tasksFixture()
	d := newDetail(t, item)

	d.taskCursor, d.subCursor = 2, -1
	if got := d.curTitle(); got != "third" {
		t.Errorf("task title = %q, want %q", got, "third")
	}
	d.setCurTitle("renamed task")
	if item.tasks[2].title != "renamed task" {
		t.Errorf("task not renamed: %q", item.tasks[2].title)
	}

	d.taskCursor, d.subCursor = 1, 1
	if got := d.curTitle(); got != "sub b" {
		t.Errorf("subtask title = %q, want %q", got, "sub b")
	}
	d.setCurTitle("renamed sub")
	if item.tasks[1].subtasks[1].title != "renamed sub" {
		t.Errorf("subtask not renamed: %q", item.tasks[1].subtasks[1].title)
	}
}

func TestDeleteCurrentTask(t *testing.T) {
	tests := []struct {
		name       string
		cursor     int
		wantTitles []string
		wantCursor int
	}{
		{"first", 0, []string{"second", "third"}, 0},
		{"middle", 1, []string{"first", "third"}, 1},
		{"last shifts the cursor back", 2, []string{"first", "second"}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tasksFixture()
			d := newDetail(t, item)
			d.taskCursor, d.subCursor = tt.cursor, -1
			d.deleteCurrent()

			if len(item.tasks) != len(tt.wantTitles) {
				t.Fatalf("got %d tasks, want %d", len(item.tasks), len(tt.wantTitles))
			}
			for i, want := range tt.wantTitles {
				if item.tasks[i].title != want {
					t.Errorf("task %d = %q, want %q", i, item.tasks[i].title, want)
				}
			}
			if d.taskCursor != tt.wantCursor {
				t.Errorf("taskCursor = %d, want %d", d.taskCursor, tt.wantCursor)
			}
		})
	}
}

func TestDeleteCurrentLastRemainingTask(t *testing.T) {
	item := &milestone{tasks: []task{{title: "only"}}}
	d := newDetail(t, item)
	d.taskCursor, d.subCursor = 0, -1
	d.deleteCurrent()

	if len(item.tasks) != 0 {
		t.Fatalf("got %d tasks, want 0", len(item.tasks))
	}
	if d.taskCursor != 0 || d.subCursor != -1 {
		t.Errorf("cursor = (%d, %d), want (0, -1)", d.taskCursor, d.subCursor)
	}
}

func TestDeleteCurrentSubtask(t *testing.T) {
	tests := []struct {
		name       string
		sub        int
		wantTitles []string
		wantSub    int
	}{
		{"first of two", 0, []string{"sub b"}, 0},
		{"last of two falls back one", 1, []string{"sub a"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tasksFixture()
			d := newDetail(t, item)
			d.taskCursor, d.subCursor = 1, tt.sub
			d.deleteCurrent()

			subs := item.tasks[1].subtasks
			if len(subs) != len(tt.wantTitles) {
				t.Fatalf("got %d subtasks, want %d", len(subs), len(tt.wantTitles))
			}
			for i, want := range tt.wantTitles {
				if subs[i].title != want {
					t.Errorf("subtask %d = %q, want %q", i, subs[i].title, want)
				}
			}
			if d.subCursor != tt.wantSub {
				t.Errorf("subCursor = %d, want %d", d.subCursor, tt.wantSub)
			}
		})
	}
}

// deleting the only subtask drops the cursor back onto the task row
func TestDeleteCurrentOnlySubtask(t *testing.T) {
	item := &milestone{tasks: []task{{title: "parent", subtasks: []subtask{{title: "only"}}}}}
	d := newDetail(t, item)
	d.taskCursor, d.subCursor = 0, 0
	d.deleteCurrent()

	if len(item.tasks[0].subtasks) != 0 {
		t.Fatalf("subtask not removed")
	}
	if d.subCursor != -1 {
		t.Errorf("subCursor = %d, want -1 (back on the task row)", d.subCursor)
	}
	if !d.onTask() {
		t.Error("cursor should be on the task row")
	}
}

func TestInsertTaskAt(t *testing.T) {
	tests := []struct {
		name       string
		idx        int
		wantTitles []string
		wantCursor int
	}{
		{"at the start", 0, []string{"", "first", "second", "third"}, 0},
		{"in the middle", 1, []string{"first", "", "second", "third"}, 1},
		{"at the end", 3, []string{"first", "second", "third", ""}, 3},
		{"past the end clamps", 99, []string{"first", "second", "third", ""}, 3},
		{"negative clamps to the start", -5, []string{"", "first", "second", "third"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tasksFixture()
			d := newDetail(t, item)
			cmd := d.insertTaskAt(tt.idx)

			if len(item.tasks) != len(tt.wantTitles) {
				t.Fatalf("got %d tasks, want %d", len(item.tasks), len(tt.wantTitles))
			}
			for i, want := range tt.wantTitles {
				if item.tasks[i].title != want {
					t.Errorf("task %d = %q, want %q", i, item.tasks[i].title, want)
				}
			}
			if d.taskCursor != tt.wantCursor || d.subCursor != -1 {
				t.Errorf("cursor = (%d, %d), want (%d, -1)", d.taskCursor, d.subCursor, tt.wantCursor)
			}
			if d.mode != detailEditingTask || !d.isNewTask {
				t.Errorf("expected to enter new-task editing, got mode=%v isNew=%v", d.mode, d.isNewTask)
			}
			if cmd == nil {
				t.Error("expected a blink command")
			}
		})
	}
}

func TestInsertSubtaskAt(t *testing.T) {
	tests := []struct {
		name       string
		idx        int
		wantTitles []string
		wantSub    int
	}{
		{"at the start", 0, []string{"", "sub a", "sub b"}, 0},
		{"in the middle", 1, []string{"sub a", "", "sub b"}, 1},
		{"at the end", 2, []string{"sub a", "sub b", ""}, 2},
		{"past the end clamps", 99, []string{"sub a", "sub b", ""}, 2},
		{"negative clamps to the start", -5, []string{"", "sub a", "sub b"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tasksFixture()
			d := newDetail(t, item)
			cmd := d.insertSubtaskAt(1, tt.idx)

			subs := item.tasks[1].subtasks
			if len(subs) != len(tt.wantTitles) {
				t.Fatalf("got %d subtasks, want %d", len(subs), len(tt.wantTitles))
			}
			for i, want := range tt.wantTitles {
				if subs[i].title != want {
					t.Errorf("subtask %d = %q, want %q", i, subs[i].title, want)
				}
			}
			if !item.tasks[1].open {
				t.Error("inserting a subtask should open the task")
			}
			if d.taskCursor != 1 || d.subCursor != tt.wantSub {
				t.Errorf("cursor = (%d, %d), want (1, %d)", d.taskCursor, d.subCursor, tt.wantSub)
			}
			if cmd == nil {
				t.Error("expected a blink command")
			}
		})
	}
}

// inserting a subtask under a task that had none opens it and seeds the list
func TestInsertSubtaskOnTaskWithNone(t *testing.T) {
	item := tasksFixture()
	d := newDetail(t, item)
	d.insertSubtaskAt(2, 0)

	if len(item.tasks[2].subtasks) != 1 {
		t.Fatalf("got %d subtasks, want 1", len(item.tasks[2].subtasks))
	}
	if !item.tasks[2].open {
		t.Error("task should be open")
	}
}

func TestApplyConfirmDeleteItem(t *testing.T) {
	item := tasksFixture()
	d := newDetail(t, item)
	d.taskCursor, d.subCursor = 0, -1
	d.confirm = newConfirm(confirmDeleteItem, "delete?")

	cmd := d.applyConfirm()
	if len(item.tasks) != 2 {
		t.Errorf("got %d tasks, want 2", len(item.tasks))
	}
	if cmd == nil {
		t.Fatal("expected a save command")
	}
	if _, ok := cmd().(detailSaveMsg); !ok {
		t.Errorf("got %T, want detailSaveMsg", cmd())
	}
}

func TestApplyConfirmToggleCompletion(t *testing.T) {
	fixed := mustDate(t, "Jun 15 2026")
	defer stubNow(t, fixed)()

	t.Run("marks an unfinished item complete as of now", func(t *testing.T) {
		item := tasksFixture()
		d := newDetail(t, item)
		d.confirm = newConfirm(confirmToggleCompletion, "complete?")

		cmd := d.applyConfirm()
		if !item.finished.Equal(fixed) {
			t.Errorf("finished = %v, want %v", item.finished, fixed)
		}
		if _, ok := cmd().(detailItemCompletedMsg); !ok {
			t.Errorf("got %T, want detailItemCompletedMsg", cmd())
		}
	})

	t.Run("clears the date on an already finished item", func(t *testing.T) {
		item := tasksFixture()
		item.finished = mustDate(t, "May 4 2026")
		d := newDetail(t, item)
		d.confirm = newConfirm(confirmToggleCompletion, "reopen?")

		cmd := d.applyConfirm()
		if !item.finished.IsZero() {
			t.Errorf("finished = %v, want zero", item.finished)
		}
		if _, ok := cmd().(detailItemUncompletedMsg); !ok {
			t.Errorf("got %T, want detailItemUncompletedMsg", cmd())
		}
	})
}

// completing a task sweeps all of its subtasks complete too
func TestApplyConfirmCompleteSubtasks(t *testing.T) {
	item := tasksFixture()
	d := newDetail(t, item)
	d.taskCursor, d.subCursor = 1, -1
	d.confirm = newConfirm(confirmCompleteSubtasks, "complete all?")

	cmd := d.applyConfirm()
	if !item.tasks[1].completed {
		t.Error("task not marked complete")
	}
	for i, st := range item.tasks[1].subtasks {
		if !st.completed {
			t.Errorf("subtask %d not marked complete", i)
		}
	}
	if _, ok := cmd().(detailSaveMsg); !ok {
		t.Errorf("got %T, want detailSaveMsg", cmd())
	}
}

func TestApplyConfirmNoneIsNoop(t *testing.T) {
	item := tasksFixture()
	d := newDetail(t, item)
	d.confirm = newConfirm(confirmNone, "")

	if cmd := d.applyConfirm(); cmd != nil {
		t.Errorf("expected nil command, got %T", cmd())
	}
	if len(item.tasks) != 3 || !item.finished.IsZero() {
		t.Error("state changed on a no-op confirm")
	}
}

func TestStartEditingNew(t *testing.T) {
	d := newDetail(t, tasksFixture())
	d.preEditTitle = "leftover"
	d.startEditingNew()

	if d.mode != detailEditingTask {
		t.Errorf("mode = %v, want detailEditingTask", d.mode)
	}
	if !d.isNewTask {
		t.Error("isNewTask should be set")
	}
	if d.preEditTitle != "" {
		t.Errorf("preEditTitle = %q, want empty", d.preEditTitle)
	}
	if d.input.Value() != "" {
		t.Errorf("input = %q, want empty", d.input.Value())
	}
	if !d.input.Focused() {
		t.Error("input should be focused")
	}
}
