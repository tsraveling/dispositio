package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Idea was to have a modal interface for completing milestones
// right from the planner; currently unused.

// Idea felt cute, might delete later.

type completeItemModal struct {
	item *milestone
}

func newCompleteItemModal(it *milestone) *completeItemModal {
	return &completeItemModal{item: it}
}

func (m *completeItemModal) Config() modalConfig {
	return modalConfig{w: 40, h: 5, xOffset: 0, yOffset: 0}
}

func (m *completeItemModal) Update(_ tea.Msg) (modal, tea.Cmd) {
	return m, nil
}

func (m *completeItemModal) View() string {
	c := m.Config()
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Width(c.w).
		Height(c.h)
	return style.Render("hello world: " + m.item.title)
}
