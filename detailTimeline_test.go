package main

import (
	"strings"
	"testing"
)

func bodyFor(t *testing.T, it *milestone, active bool, mode timelineMode) string {
	t.Helper()
	start := mustDate(t, "Jun 1 2026")
	var dv *detailViewModel
	if active {
		d := makeDetailViewModel(it, 80, start, true, mode)
		dv = &d
	}
	return plain(getBody(it, dv, 80, 40, start, true, mode))
}

// None mode drops every date-derived line for an unfinished item
func TestDetailBodyTimelineNoneUnfinished(t *testing.T) {
	defer stubNow(t, mustDate(t, "Sep 1 2026"))() // well past due
	it := tasksFixture()
	it.duration = 2

	out := bodyFor(t, it, true, timelineNone)
	for _, bad := range []string{"Due:", "Actual", "past", "weekdays per", "tasks per"} {
		if strings.Contains(out, bad) {
			t.Errorf("None mode body contains %q:\n%s", bad, out)
		}
	}
	for _, want := range []string{"Estimated: 2w", "hit c to mark", "2 tasks remaining"} {
		if !strings.Contains(out, want) {
			t.Errorf("None mode body missing %q:\n%s", want, out)
		}
	}

	// inactive panel drops the hint but keeps the estimate
	inactive := bodyFor(t, it, false, timelineNone)
	if strings.Contains(inactive, "hit c") || !strings.Contains(inactive, "Estimated: 2w") {
		t.Errorf("inactive None body wrong:\n%s", inactive)
	}

	// weeks mode on the same item still shows dates
	weeks := bodyFor(t, it, true, timelineWeeks)
	for _, want := range []string{"Due:", "Actual", "per"} {
		if !strings.Contains(weeks, want) {
			t.Errorf("weeks mode body missing %q:\n%s", want, weeks)
		}
	}
}

func TestDetailBodyTimelineNoneSingularRemaining(t *testing.T) {
	it := &milestone{title: "M", duration: 1, tasks: []task{{title: "a", completed: true}, {title: "b"}}}
	if out := bodyFor(t, it, true, timelineNone); !strings.Contains(out, "1 task remaining") {
		t.Errorf("singular form missing:\n%s", out)
	}
	it.tasks[1].completed = true
	if out := bodyFor(t, it, true, timelineNone); !strings.Contains(out, "All subtasks complete!") {
		t.Errorf("complete message missing:\n%s", out)
	}
}

// None mode keeps completion date, estimate and the date-change hint, drops Actual
func TestDetailBodyTimelineNoneFinished(t *testing.T) {
	it := tasksFixture()
	it.duration = 2
	it.finished = mustDate(t, "Jun 20 2026")

	out := bodyFor(t, it, true, timelineNone)
	for _, want := range []string{"Completed on Jun 20, 2026", "Estimated: 2w", "-+ change date"} {
		if !strings.Contains(out, want) {
			t.Errorf("None mode finished body missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Actual") {
		t.Errorf("None mode finished body shows Actual:\n%s", out)
	}
	if weeks := bodyFor(t, it, true, timelineWeeks); !strings.Contains(weeks, "Actual: 3w") {
		t.Errorf("weeks mode finished body missing Actual:\n%s", weeks)
	}
}
