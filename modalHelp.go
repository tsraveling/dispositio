package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// @region modal:help -- HELP MODAL + KEYMAP TABLES

var modalHelpType modalType = "help"

// a key/description pair, or a section header when key is "".
type helpRow struct {
	key  string
	desc string
}

type helpModal struct {
	rows []helpRow
}

var everywhereRows = []helpRow{
	{"", "Everywhere"},
	{"?", "this help"},
	{"esc", "close panel / modal"},
	{"q / ctrl+c", "quit"},
}

var plannerHelpRows = append([]helpRow{
	{"", "Planner"},
	{"↑/↓ j/k", "move cursor"},
	{"a / o / O", "add item (end / below / above)"},
	{"e", "rename item or project"},
	{"d", "delete item"},
	{"←/→ h/l", "open / close item detail"},
	{"K/J", "reorder milestone"},
	{"H/L", "change duration"},
}, everywhereRows...)

var detailHelpRows = append([]helpRow{
	{"", "Detail View"},
	{"↑/↓ j/k", "move between tasks & subtasks"},
	{"c", "mark milestone complete"},
	{"enter", "edit milestone description"},
	{"", "Tasks"},
	{"← / h", "return to planner"},
	{"→ / l", "expand subtasks"},
	{"shift→ / L", "make subtask"},
	{"", "Subtasks"},
	{"← / h", "collapse subtasks"},
	{"shift← / H", "make task"},
	{"", "Both"},
	{"a / o / O", "add to end of list"},
	{"e", "rename"},
	{"space / x", "toggle complete"},
	{"d", "delete"},
	{"K/J", "reorder"},
}, everywhereRows...)

func newPlannerHelpModal() *helpModal { return &helpModal{rows: plannerHelpRows} }
func newDetailHelpModal() *helpModal  { return &helpModal{rows: detailHelpRows} }

func (m *helpModal) Config() modalConfig {
	return modalConfig{w: 46, h: len(m.rows) + 4, xOffset: 0, yOffset: 0}
}

func (m *helpModal) Update(msg tea.Msg) (modal, tea.Cmd) {
	// Any key other than esc (handled by modalUpdate) closes the help.
	if _, ok := msg.(tea.KeyMsg); ok {
		return nil, nil
	}
	return m, nil
}

func (m *helpModal) View() string {
	keyStyle := lipgloss.NewStyle().Foreground(highlightColor)
	descStyle := lipgloss.NewStyle().Foreground(textColor)

	var b strings.Builder
	for i, r := range m.rows {
		if r.key == "" {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(dimStyle.Render(strings.ToUpper(r.desc)) + "\n")
			continue
		}
		b.WriteString(fmt.Sprintf("%s  %s\n",
			keyStyle.Render(fmt.Sprintf("%-12s", r.key)),
			descStyle.Render(r.desc)))
	}
	b.WriteString("\n" + dimStyle.Italic(true).Render("any key to close"))

	c := m.Config()
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Width(c.w).
		Render(b.String())
}
