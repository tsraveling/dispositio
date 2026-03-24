package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type plannerViewModel struct {
	prj          project
	cursor       int  // selected item index
	editingTitle bool // true if we are in edit title mode
	input        textinput.Model
}

func makePlannerViewModel(p *project) (plannerViewModel, tea.Cmd) {
	ti := textinput.New()
	ti.Placeholder = "Item title..."
	ti.CharLimit = 120

	// Copy project into value mode so we can mutate it bubbletea-style
	m := plannerViewModel{prj: *p, input: ti}
	return m, m.Init()
}

func (m plannerViewModel) Init() tea.Cmd {
	return nil
}

// addNewAt inserts a new empty item after the given index and enters editing mode.
func (m *plannerViewModel) addNewAt(index int) tea.Cmd {
	newItem := item{title: "", duration: 1}
	insertAt := index + 1

	// Insert into slice
	m.prj.items = append(m.prj.items, item{})
	copy(m.prj.items[insertAt+1:], m.prj.items[insertAt:])
	m.prj.items[insertAt] = newItem

	m.cursor = insertAt
	m.editingTitle = true
	m.input.SetValue("")
	m.input.Focus()
	return textinput.Blink
}

func (m plannerViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cfg.updateWW(msg.Width)
		cfg.updateWH(msg.Height)
	default:
		_ = msg
	}

	// SECTION: Editing input

	if m.editingTitle {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.editingTitle = false
				m.input.Blur()
				// If title is empty, remove the item
				if m.prj.items[m.cursor].title == "" {
					m.prj.items = append(m.prj.items[:m.cursor], m.prj.items[m.cursor+1:]...)
					if m.cursor > 0 && m.cursor >= len(m.prj.items) {
						m.cursor--
					}
				}
				return m, nil
			case "enter":
				m.prj.items[m.cursor].title = m.input.Value()
				m.editingTitle = false
				m.input.Blur()
				err := m.prj.save()
				if err != nil {
					panic(err)
				}
				return m, nil
			}
		}

		// Forward to text input
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m.prj.items[m.cursor].title = m.input.Value()
		return m, cmd
	}

	// SECTION: Normal input

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.prj.items)-1 {
				m.cursor++
			}
		case "a":
			cmd := m.addNewAt(m.cursor)
			return m, cmd
		case "shift+up", "K":
			if m.cursor > 0 {
				m.prj.items[m.cursor], m.prj.items[m.cursor-1] = m.prj.items[m.cursor-1], m.prj.items[m.cursor]
				m.cursor--
				m.prj.save()
			}
		case "shift+down", "J":
			if m.cursor < len(m.prj.items)-1 {
				m.prj.items[m.cursor], m.prj.items[m.cursor+1] = m.prj.items[m.cursor+1], m.prj.items[m.cursor]
				m.cursor++
				m.prj.save()
			}
		case "shift+left", "H":
			if m.prj.items[m.cursor].duration > 1 {
				m.prj.items[m.cursor].duration--
				m.prj.save()
			}
		case "shift+right", "L":
			m.prj.items[m.cursor].duration++
			m.prj.save()
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

// SECTION: Rendering

func (m plannerViewModel) View() string {
	now := time.Now()

	// Find the Monday of the current week
	weekday := now.Weekday()
	daysUntilMonday := (int(weekday) - int(time.Monday) + 7) % 7
	monday := now.AddDate(0, 0, -daysUntilMonday)

	selectedStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(textColor)

	var sb strings.Builder
	row := 0
	for i, it := range m.prj.items {
		style := normalStyle
		if i == m.cursor {
			style = selectedStyle
		}
		for w := range it.duration {
			weekStart := monday.AddDate(0, 0, row*7)
			_, week := weekStart.ISOWeek()
			date := fmt.Sprintf("%d.%d", int(weekStart.Month()), weekStart.Day())

			line := fmt.Sprintf("W%-3d %-5s ", week, date)
			if w == 0 {
				if m.editingTitle && i == m.cursor {
					line += "-" + m.input.View()
				} else {
					line += "⬤  " + it.title
				}
			} else {
				line += "⚬"
			}

			sb.WriteString(style.Render(line) + "\n")
			row++
		}
	}
	return sb.String()
}
