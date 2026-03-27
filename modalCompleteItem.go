package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type completeItemModal struct {
	item *item
}

func newCompleteItemModal(it *item) *completeItemModal {
	return &completeItemModal{item: it}
}

func (m *completeItemModal) Config() modalConfig {
	return modalConfig{w: 40, h: 5, xOffset: 0, yOffset: 0}
}

func (m *completeItemModal) Update(msg tea.Msg) (modal, tea.Cmd) {
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
