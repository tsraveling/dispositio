package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModalUpdate(t *testing.T) {
	t.Run("nil modal stays nil", func(t *testing.T) {
		got, cmd := modalUpdate(nil, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
		if got != nil || cmd != nil {
			t.Errorf("got (%v, %v), want (nil, nil)", got, cmd)
		}
	})

	t.Run("esc closes any modal", func(t *testing.T) {
		got, cmd := modalUpdate(newPlannerHelpModal(), tea.KeyMsg{Type: tea.KeyEsc})
		if got != nil {
			t.Errorf("modal = %v, want nil", got)
		}
		if cmd != nil {
			t.Errorf("cmd = %v, want nil", cmd)
		}
	})

	t.Run("non-key messages reach the modal", func(t *testing.T) {
		m := newCompleteItemModal(&milestone{title: "M"})
		got, _ := modalUpdate(m, tea.WindowSizeMsg{Width: 80, Height: 24})
		if got == nil {
			t.Error("modal was closed by a window size message")
		}
	})
}

func TestModalView(t *testing.T) {
	t.Run("nil modal returns the background unchanged", func(t *testing.T) {
		const bg = "background text"
		if got := modalView(nil, bg); got != bg {
			t.Errorf("got %q, want %q", got, bg)
		}
	})

	t.Run("a modal is composited over the background", func(t *testing.T) {
		withWidth(t, 100)
		prev := cfg.wh
		cfg.updateWH(30)
		t.Cleanup(func() { cfg.wh = prev })

		out := plain(modalView(newPlannerHelpModal(), "background text"))
		if !strings.Contains(out, "this help") {
			t.Errorf("modal content missing from composite:\n%s", out)
		}
	})
}

// any key closes the help modal; esc is handled a level up in modalUpdate
func TestHelpModalClosesOnAnyKey(t *testing.T) {
	m := newDetailHelpModal()
	got, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")})
	if got != nil {
		t.Errorf("modal = %v, want nil", got)
	}
	if cmd != nil {
		t.Errorf("cmd = %v, want nil", cmd)
	}

	stay, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if stay == nil {
		t.Error("a non-key message should not close the help modal")
	}
}

func TestHelpModalConfigHeightTracksRowCount(t *testing.T) {
	planner := newPlannerHelpModal()
	if got, want := planner.Config().h, len(plannerHelpRows)+4; got != want {
		t.Errorf("height = %d, want %d", got, want)
	}
	detail := newDetailHelpModal()
	if got, want := detail.Config().h, len(detailHelpRows)+4; got != want {
		t.Errorf("height = %d, want %d", got, want)
	}
	if planner.Config().w != 46 {
		t.Errorf("width = %d, want 46", planner.Config().w)
	}
}

// section headers render as uppercase labels, key rows as key/description pairs
func TestHelpModalView(t *testing.T) {
	out := plain(newPlannerHelpModal().View())

	for _, want := range []string{"PLANNER", "EVERYWHERE", "move cursor", "this help", "any key to close"} {
		if !strings.Contains(out, want) {
			t.Errorf("view missing %q:\n%s", want, out)
		}
	}
}

func TestPlannerAndDetailHelpRowsDiffer(t *testing.T) {
	planner := plain(newPlannerHelpModal().View())
	detail := plain(newDetailHelpModal().View())

	if !strings.Contains(detail, "Detail View") && !strings.Contains(detail, "DETAIL VIEW") {
		t.Error("detail help missing its own section")
	}
	if planner == detail {
		t.Error("planner and detail help render identically")
	}
}

// every help row must have a description; a blank key marks a section header
func TestHelpRowsAreWellFormed(t *testing.T) {
	for name, rows := range map[string][]helpRow{
		"planner": plannerHelpRows,
		"detail":  detailHelpRows,
	} {
		for i, r := range rows {
			if strings.TrimSpace(r.desc) == "" {
				t.Errorf("%s row %d has an empty description", name, i)
			}
		}
	}
}

func TestCompleteItemModal(t *testing.T) {
	item := &milestone{title: "Proto Map"}
	m := newCompleteItemModal(item)

	if got := m.Config(); got.w != 40 || got.h != 5 {
		t.Errorf("config = %+v, want w=40 h=5", got)
	}

	got, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if got == nil {
		t.Error("modal closed itself")
	}
	if cmd != nil {
		t.Errorf("cmd = %v, want nil", cmd)
	}

	if out := plain(m.View()); !strings.Contains(out, "Proto Map") {
		t.Errorf("view missing the item title:\n%s", out)
	}
}
