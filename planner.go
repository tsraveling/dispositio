package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type plannerMode int

const (
	normal plannerMode = iota
	editingTitle
	confirmingDeletion
)

type plannerViewModel struct {
	prj    project
	cursor int // selected item index
	mode   plannerMode
	input  textinput.Model
}

func makePlannerViewModel(p *project) (plannerViewModel, tea.Cmd) {
	ti := textinput.New()
	ti.Placeholder = "Item title..."
	ti.CharLimit = 120

	// Copy project into value mode so we can mutate it bubbletea-style
	m := plannerViewModel{prj: *p, input: ti, mode: normal}
	return m, m.Init()
}

func (m plannerViewModel) Init() tea.Cmd {
	return nil
}

// addNewAt inserts a new empty item after the given index and enters editing mode.
func (m *plannerViewModel) addNewAt(index int) tea.Cmd {
	newItem := item{title: "", duration: 1}

	var insertAt int

	if len(m.prj.items) < 1 {
		insertAt = 0
		m.prj.items = []item{newItem}
	} else {

		insertAt = index + 1

		// Insert into slice
		m.prj.items = append(m.prj.items, item{})
		copy(m.prj.items[insertAt+1:], m.prj.items[insertAt:])
		m.prj.items[insertAt] = newItem
	}

	m.cursor = insertAt
	m.mode = editingTitle
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

	switch m.mode {
	// SECTION: Title Editing Mode
	case editingTitle:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.mode = normal
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
				m.mode = normal
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

	// SECTION: Confirming Deletion Mode
	case confirmingDeletion:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y":
				m.prj.items = append(m.prj.items[:m.cursor], m.prj.items[m.cursor+1:]...)
				if m.cursor >= len(m.prj.items) && m.cursor > 0 {
					m.cursor--
				}
				m.mode = normal
				m.prj.save()
				return m, nil
			case "n", "esc":
				m.mode = normal
				return m, nil
			}
		}

	// SECTION: Normal input mode
	case normal:
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
			case "d":
				if len(m.prj.items) > 0 {
					m.mode = confirmingDeletion
				}
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
	}

	return m, nil
}

// SECTION: Rendering

func (m plannerViewModel) View() string {

	// If there's nothin' we can bail early:
	if len(m.prj.items) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(dimColor)
		return dimStyle.Render("There are no items in this plan. Press 'a' to add one.") + "\n"
	}

	// Otherwise, find the Monday of the current week
	now := time.Now()
	weekday := now.Weekday()
	daysUntilMonday := (int(weekday) - int(time.Monday) + 7) % 7
	monday := now.AddDate(0, 0, -daysUntilMonday)

	// Set up the local styles for this view
	selectedStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(textColor)
	deleteStyle := lipgloss.NewStyle().Foreground(warningColor).Bold(true)

	// Generate the lines for the whole project first
	var lines []string
	cursorRow := 0
	row := 0
	for i, it := range m.prj.items {
		style := normalStyle
		if i == m.cursor {
			style = selectedStyle
			cursorRow = row
		}
		for w := range it.duration {

			// Get the date of the first monday, and the week of the year
			// TODO: Add start of week day to config.ini
			weekStart := monday.AddDate(0, 0, row*7)
			_, week := weekStart.ISOWeek()

			// Assemble the date, MM.DD
			// TODO: Add EU-style dates to config.ini
			date := fmt.Sprintf("%d.%d", int(weekStart.Month()), weekStart.Day())

			line := fmt.Sprintf("W%-3d %-5s ", week, date)

			if w == 0 {
				if m.mode == editingTitle && i == m.cursor {
					// IF EDITING: Show input.
					line += "-" + m.input.View()
				} else if m.mode == confirmingDeletion && i == m.cursor {
					// IF DELETING: Show confirmation.
					line += "⬤  " + deleteStyle.Render("Delete? y/n")
				} else {
					// Otherwise just show the normal title.
					line += "⬤  " + it.title
				}
			} else {
				line += "⚬"
			}

			lines = append(lines, style.Render(line))
			row++
		}
	}

	// Grab only the chunk of them that are currently visible in the viewport,
	// and then just display that.
	viewHeight := cfg.wh
	if viewHeight > 0 && len(lines) > viewHeight {
		start := max(cursorRow-viewHeight/2, 0)
		end := start + viewHeight
		if end > len(lines) {
			end = len(lines)
			start = end - viewHeight
		}
		lines = lines[start:end]
	}

	return strings.Join(lines, "\n") + "\n"
}
