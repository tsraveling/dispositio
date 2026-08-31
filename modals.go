package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

// @region modal:core -- MODAL INTERFACE + OVERLAY

type modalConfig struct {
	w       int
	h       int
	xOffset int
	yOffset int
}

type modal interface {
	Update(msg tea.Msg) (modal, tea.Cmd)
	View() string
	Config() modalConfig
}

// routes an update to the active modal; returns nil modal on esc,
// which closes any modal.
func modalUpdate(m modal, msg tea.Msg) (modal, tea.Cmd) {
	if m == nil {
		return nil, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "esc" {
		return nil, nil
	}

	return m.Update(msg)
}

// composites the active modal over the background view string.
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

// overlays a save failure over the composed view. The user keeps their work on
// screen; the message clears on the next successful save.
func saveErrorView(view string, err error) string {
	if err == nil {
		return view
	}
	box := errorBoxStyle(boxWidth(cfg.ww) - 8).Render("Could not save: " + err.Error())
	padded := lipgloss.Place(cfg.ww, cfg.wh, lipgloss.Left, lipgloss.Top, view)
	return overlay.Composite(box, padded, overlay.Center, overlay.Center, 0, 0)
}
