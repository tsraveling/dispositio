package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

type modalType string

var modalNone modalType = ""
var modalCompleteItem modalType = "complete-item"

// modalCloseMsg is sent by any modal to request closing itself.
type modalCloseMsg struct{}

// modalConfig holds layout parameters for a modal.
type modalConfig struct {
	w       int
	h       int
	xOffset int
	yOffset int
}

// modal defines the interface all modals must satisfy.
type modal interface {
	Update(msg tea.Msg) (modal, tea.Cmd)
	View() string
	Config() modalConfig
}

// modalUpdate routes an update to the active modal.
// Returns nil modal if esc was pressed (closing it).
func modalUpdate(m modal, msg tea.Msg) (modal, tea.Cmd) {
	if m == nil {
		return nil, nil
	}

	// Any modal closes on esc
	if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "esc" {
		return nil, nil
	}

	return m.Update(msg)
}

// modalView composites the active modal over the background view string.
func modalView(m modal, bg string) string {
	if m == nil {
		return bg
	}
	c := m.Config()
	// Pad background to full terminal size so centering works against the viewport.
	padded := lipgloss.Place(cfg.ww, cfg.wh, lipgloss.Left, lipgloss.Top, bg)
	return overlay.Composite(
		m.View(), padded,
		overlay.Center, overlay.Center,
		c.xOffset, c.yOffset,
	)
}
