package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// @region modal:settings -- PROJECT SETTINGS MODAL

// emitted when the user saves; the planner applies the values and persists.
type settingsSavedMsg struct {
	startDate time.Time
	timeline  timelineMode
}

type settingsRow int

const (
	settingsRowStart settingsRow = iota
	settingsRowTimeline
	settingsRowCount
)

// edits copies of the project settings; nothing touches the project until
// enter emits settingsSavedMsg. Esc is handled by modalUpdate (cancel).
type settingsModal struct {
	startDate time.Time
	timeline  timelineMode
	row       settingsRow
}

func newSettingsModal(p *project) *settingsModal {
	return &settingsModal{startDate: p.startDate, timeline: p.timeline}
}

func (m *settingsModal) Config() modalConfig {
	return modalConfig{w: 46, h: 9, xOffset: 0, yOffset: 0}
}

func (m *settingsModal) Update(msg tea.Msg) (modal, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch k := key.String(); k {
	case "up", "k":
		if m.row > 0 {
			m.row--
		}
	case "down", "j":
		if m.row < settingsRowCount-1 {
			m.row++
		}
	case "enter":
		saved := settingsSavedMsg{startDate: m.startDate, timeline: m.timeline}
		return nil, func() tea.Msg { return saved }
	default:
		switch m.row {
		case settingsRowStart:
			if d, ok := adjustDate(m.startDate, k); ok {
				m.startDate = d
			}
		case settingsRowTimeline:
			switch k {
			case "right", "l", "shift+right", "L":
				m.timeline = (m.timeline + 1) % timelineModeCount
			case "left", "h", "shift+left", "H":
				m.timeline = (m.timeline + timelineModeCount - 1) % timelineModeCount
			}
		}
	}
	return m, nil
}

// nudges a date by key: h/l ±1 day, H/L (shift) ±1 week. ok is false for
// any other key.
func adjustDate(t time.Time, key string) (time.Time, bool) {
	switch key {
	case "left", "h":
		return t.AddDate(0, 0, -1), true
	case "right", "l":
		return t.AddDate(0, 0, 1), true
	case "shift+left", "H":
		return t.AddDate(0, 0, -7), true
	case "shift+right", "L":
		return t.AddDate(0, 0, 7), true
	}
	return t, false
}

func (m *settingsModal) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("PROJECT SETTINGS") + "\n\n")

	rows := []struct {
		label, value string
	}{
		{"Start date", fmtFullDate(m.startDate)},
		{"Timeline", m.timeline.String()},
	}
	for i, r := range rows {
		style := lipgloss.NewStyle().Foreground(textColor)
		marker := "  "
		if settingsRow(i) == m.row {
			style = highlightedStyle
			marker = "> "
		}
		fmt.Fprintf(&b, "%s%s\n", marker, style.Render(fmt.Sprintf("%-12s %s", r.label, r.value)))
	}

	b.WriteString("\n" + dimStyle.Render("↑↓ select  ←→ change  enter save  esc cancel"))

	c := m.Config()
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Width(c.w).
		Render(b.String())
}
