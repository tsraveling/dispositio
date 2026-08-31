package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfirmActive(t *testing.T) {
	tests := []struct {
		name string
		kind confirmKind
		want bool
	}{
		{"none is inactive", confirmNone, false},
		{"delete item", confirmDeleteItem, true},
		{"toggle completion", confirmToggleCompletion, true},
		{"complete subtasks", confirmCompleteSubtasks, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newConfirm(tt.kind, "prompt text")
			if got := c.active(); got != tt.want {
				t.Errorf("active() = %v, want %v", got, tt.want)
			}
			if c.prompt != "prompt text" {
				t.Errorf("prompt = %q", c.prompt)
			}
			if c.kind != tt.kind {
				t.Errorf("kind = %v, want %v", c.kind, tt.kind)
			}
		})
	}
}

func TestConfirmHandle(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.Msg
		want confirmResult
	}{
		{"y confirms", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}, confirmYes},
		{"n declines", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}, confirmNo},
		{"esc declines", tea.KeyMsg{Type: tea.KeyEsc}, confirmNo},
		{"other keys stay pending", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}, confirmPending},
		{"enter stays pending", tea.KeyMsg{Type: tea.KeyEnter}, confirmPending},
		{"uppercase Y stays pending", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Y")}, confirmPending},
		{"non-key messages stay pending", tea.WindowSizeMsg{Width: 80, Height: 24}, confirmPending},
		{"nil message stays pending", nil, confirmPending},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newConfirm(confirmDeleteItem, "sure?")
			if got := c.handle(tt.msg); got != tt.want {
				t.Errorf("handle(%v) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}
