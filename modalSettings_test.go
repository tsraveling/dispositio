package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keyRunes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestAdjustDate(t *testing.T) {
	base := mustDate(t, "Jun 10 2026")
	tests := []struct {
		key    string
		want   string
		wantOK bool
	}{
		{"left", "Jun 9 2026", true},
		{"h", "Jun 9 2026", true},
		{"right", "Jun 11 2026", true},
		{"l", "Jun 11 2026", true},
		{"shift+left", "Jun 3 2026", true},
		{"H", "Jun 3 2026", true},
		{"shift+right", "Jun 17 2026", true},
		{"L", "Jun 17 2026", true},
		{"x", "Jun 10 2026", false},
	}
	for _, tt := range tests {
		got, ok := adjustDate(base, tt.key)
		if ok != tt.wantOK || !got.Equal(mustDate(t, tt.want)) {
			t.Errorf("adjustDate(%q) = (%v, %v), want (%s, %v)", tt.key, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestSettingsModalCopiesProject(t *testing.T) {
	p := &project{startDate: mustDate(t, "Jun 1 2026"), timeline: timelineNone}
	m := newSettingsModal(p)
	m.Update(keyRunes("l"))
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(keyRunes("l"))
	if !p.startDate.Equal(mustDate(t, "Jun 1 2026")) || p.timeline != timelineNone {
		t.Errorf("project mutated before save: %+v", p)
	}
}

func TestSettingsModalRowNav(t *testing.T) {
	m := newSettingsModal(&project{})
	if m.row != settingsRowStart {
		t.Fatalf("initial row = %v", m.row)
	}
	m.Update(keyRunes("k"))
	if m.row != settingsRowStart {
		t.Errorf("up on first row moved to %v", m.row)
	}
	m.Update(keyRunes("j"))
	if m.row != settingsRowTimeline {
		t.Errorf("down = %v, want timeline row", m.row)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.row != settingsRowTimeline {
		t.Errorf("down on last row moved to %v", m.row)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.row != settingsRowStart {
		t.Errorf("up = %v, want start row", m.row)
	}
}

func TestSettingsModalEditsDateOnStartRow(t *testing.T) {
	m := newSettingsModal(&project{startDate: mustDate(t, "Jun 10 2026")})
	m.Update(keyRunes("L"))
	m.Update(keyRunes("h"))
	if want := mustDate(t, "Jun 16 2026"); !m.startDate.Equal(want) {
		t.Errorf("startDate = %v, want %v", m.startDate, want)
	}
	if m.timeline != timelineWeeks {
		t.Errorf("timeline changed on the date row: %v", m.timeline)
	}
}

func TestSettingsModalTogglesTimeline(t *testing.T) {
	m := newSettingsModal(&project{startDate: mustDate(t, "Jun 10 2026")})
	m.Update(keyRunes("j"))

	m.Update(keyRunes("l"))
	if m.timeline != timelineNone {
		t.Errorf("after l: %v, want None", m.timeline)
	}
	m.Update(keyRunes("l"))
	if m.timeline != timelineWeeks {
		t.Errorf("after second l: %v, want Weeks (wrap)", m.timeline)
	}
	m.Update(keyRunes("h"))
	if m.timeline != timelineNone {
		t.Errorf("after h: %v, want None (wrap backwards)", m.timeline)
	}
	if !m.startDate.Equal(mustDate(t, "Jun 10 2026")) {
		t.Errorf("date changed on the timeline row: %v", m.startDate)
	}
}

func TestSettingsModalEnterEmitsSaved(t *testing.T) {
	m := newSettingsModal(&project{startDate: mustDate(t, "Jun 10 2026")})
	m.Update(keyRunes("l"))
	m.Update(keyRunes("j"))
	m.Update(keyRunes("l"))

	got, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if got != nil {
		t.Error("enter should close the modal")
	}
	if cmd == nil {
		t.Fatal("enter should emit a command")
	}
	saved, ok := cmd().(settingsSavedMsg)
	if !ok {
		t.Fatalf("cmd yielded %T, want settingsSavedMsg", cmd())
	}
	if !saved.startDate.Equal(mustDate(t, "Jun 11 2026")) || saved.timeline != timelineNone {
		t.Errorf("saved = %+v", saved)
	}
}

func TestSettingsModalUnknownKeyIsNoop(t *testing.T) {
	m := newSettingsModal(&project{startDate: mustDate(t, "Jun 10 2026")})
	got, cmd := m.Update(keyRunes("x"))
	if got == nil || cmd != nil {
		t.Errorf("got (%v, %v), want (modal, nil)", got, cmd)
	}
	if !m.startDate.Equal(mustDate(t, "Jun 10 2026")) || m.timeline != timelineWeeks || m.row != settingsRowStart {
		t.Errorf("state changed: %+v", m)
	}
}

func TestSettingsModalView(t *testing.T) {
	m := newSettingsModal(&project{startDate: mustDate(t, "Jun 10 2026"), timeline: timelineNone})
	out := plain(m.View())
	for _, want := range []string{"PROJECT SETTINGS", "> Start date", "Wed, Jun 10, 2026", "  Timeline", "None", "enter save"} {
		if !strings.Contains(out, want) {
			t.Errorf("view missing %q:\n%s", want, out)
		}
	}
	m.Update(keyRunes("j"))
	out = plain(m.View())
	if !strings.Contains(out, "> Timeline") || strings.Contains(out, "> Start date") {
		t.Errorf("marker did not follow the row:\n%s", out)
	}
}

func TestPlannerOpensSettingsWithS(t *testing.T) {
	m, _ := makePlannerViewModel(projectFixture(t))
	for _, cursor := range []int{0, 2} {
		m.cursor = cursor
		m.currentModal = nil
		got, _ := m.Update(keyRunes("s"))
		pm := got.(plannerViewModel)
		if _, ok := pm.currentModal.(*settingsModal); !ok {
			t.Errorf("cursor %d: modal = %T, want *settingsModal", cursor, pm.currentModal)
		}
	}
}

func TestPlannerAppliesSettingsSaved(t *testing.T) {
	m, _ := makePlannerViewModel(projectFixture(t))
	m.prj.filePath = filepath.Join(t.TempDir(), "ROADMAP.md")

	got, _ := m.Update(settingsSavedMsg{startDate: mustDate(t, "Jul 1 2026"), timeline: timelineNone})
	pm := got.(plannerViewModel)
	if !pm.prj.startDate.Equal(mustDate(t, "Jul 1 2026")) || pm.prj.timeline != timelineNone {
		t.Errorf("project = start %v timeline %v", pm.prj.startDate, pm.prj.timeline)
	}
	if pm.saveErr != nil {
		t.Fatalf("save failed: %v", pm.saveErr)
	}
	data, err := os.ReadFile(pm.prj.filePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Project Start: Jul 1 2026\nTimeline: None\n") {
		t.Errorf("file missing saved settings:\n%s", data)
	}
}

// inline meta-row date editing moved into the settings modal
func TestPlannerMetaRowNoLongerEditsDate(t *testing.T) {
	m, _ := makePlannerViewModel(projectFixture(t))
	m.cursor = 0
	before := m.prj.startDate
	for _, k := range []string{"h", "l", "H", "L"} {
		got, _ := m.Update(keyRunes(k))
		m = got.(plannerViewModel)
	}
	if !m.prj.startDate.Equal(before) {
		t.Errorf("startDate changed to %v", m.prj.startDate)
	}
	withWidth(t, 100)
	if out := plain(m.plannerView()); !strings.Contains(out, "s: project settings") {
		t.Errorf("meta hint missing:\n%s", out)
	}
}
