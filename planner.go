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
	cursor int // 0 = meta (start date), 1..N = items[0..N-1]
	mode   plannerMode
	input  textinput.Model
}

// itemIndex returns the index into prj.items for the current cursor,
// or -1 if the cursor is on the meta item.
func (m *plannerViewModel) itemIndex() int {
	return m.cursor - 1
}

// onMeta returns true if the cursor is on the meta start-date item.
func (m *plannerViewModel) onMeta() bool {
	return m.cursor == 0
}

// isHoveringMeta returns true if the cursor is currently over the meta item
// and we're in normal mode (i.e. meta controls should be shown/active).
func (m *plannerViewModel) isHoveringMeta() bool {
	return m.cursor == 0 && m.mode == normal
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

// addNewAt inserts a new empty item after the current cursor position and enters editing mode.
// cursor is in cursor-space (0=meta, 1+=items).
func (m *plannerViewModel) addNewAt(cursor int) tea.Cmd {
	newItem := item{title: "", duration: 1}

	// Convert cursor to item index; if on meta (0), insert at position 0
	itemIdx := cursor // item insert position in items slice
	if cursor == 0 {
		itemIdx = 0
	} else {
		itemIdx = cursor // cursor 1 -> after items[0], etc.
	}

	if len(m.prj.items) < 1 {
		m.prj.items = []item{newItem}
		itemIdx = 0
	} else {
		m.prj.items = append(m.prj.items, item{})
		copy(m.prj.items[itemIdx+1:], m.prj.items[itemIdx:])
		m.prj.items[itemIdx] = newItem
	}

	m.cursor = itemIdx + 1 // convert back to cursor-space
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
		idx := m.itemIndex()
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.mode = normal
				m.input.Blur()
				// If title is empty, remove the item
				if m.prj.items[idx].title == "" {
					m.prj.items = append(m.prj.items[:idx], m.prj.items[idx+1:]...)
					if idx >= len(m.prj.items) && m.cursor > 1 {
						m.cursor--
					}
				}
				return m, nil
			case "enter":
				m.prj.items[idx].title = m.input.Value()
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
		m.prj.items[idx].title = m.input.Value()
		return m, cmd

	// SECTION: Confirming Deletion Mode
	case confirmingDeletion:
		idx := m.itemIndex()
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y":
				m.prj.items = append(m.prj.items[:idx], m.prj.items[idx+1:]...)
				if m.cursor > len(m.prj.items) {
					m.cursor = max(1, len(m.prj.items))
				}
				if len(m.prj.items) == 0 {
					m.cursor = 0
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
		maxCursor := len(m.prj.items) // 0=meta, 1..N=items
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch keypress := msg.String(); keypress {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < maxCursor {
					m.cursor++
				}
			case "a":
				cmd := m.addNewAt(m.cursor)
				return m, cmd
			case "d":
				if !m.onMeta() && len(m.prj.items) > 0 {
					m.mode = confirmingDeletion
				}
			case "shift+up", "K":
				idx := m.itemIndex()
				if !m.onMeta() && idx > 0 {
					m.prj.items[idx], m.prj.items[idx-1] = m.prj.items[idx-1], m.prj.items[idx]
					m.cursor--
					m.prj.save()
				}
			case "shift+down", "J":
				idx := m.itemIndex()
				if !m.onMeta() && idx < len(m.prj.items)-1 {
					m.prj.items[idx], m.prj.items[idx+1] = m.prj.items[idx+1], m.prj.items[idx]
					m.cursor++
					m.prj.save()
				}
			case "left", "h":
				if m.isHoveringMeta() {
					m.prj.startDate = m.prj.startDate.AddDate(0, 0, -1)
					m.prj.save()
				}
			case "right", "l":
				if m.isHoveringMeta() {
					m.prj.startDate = m.prj.startDate.AddDate(0, 0, 1)
					m.prj.save()
				}
			case "shift+left", "H":
				if m.isHoveringMeta() {
					m.prj.startDate = m.prj.startDate.AddDate(0, 0, -7)
					m.prj.save()
				} else if !m.onMeta() {
					idx := m.itemIndex()
					if m.prj.items[idx].duration > 1 {
						m.prj.items[idx].duration--
						m.prj.save()
					}
				}
			case "shift+right", "L":
				if m.isHoveringMeta() {
					m.prj.startDate = m.prj.startDate.AddDate(0, 0, 7)
					m.prj.save()
				} else if !m.onMeta() {
					idx := m.itemIndex()
					m.prj.items[idx].duration++
					m.prj.save()
				}
			case "esc", "ctrl+c":
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// SECTION: Rendering

func (m plannerViewModel) View() string {

	// Set up the local styles for this view
	selectedStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(textColor)
	fadeStyle := lipgloss.NewStyle().Foreground(fadeColor)
	deleteStyle := lipgloss.NewStyle().Foreground(warningColor).Bold(true)

	// Use the project start date as the base Monday
	startDate := m.prj.startDate
	weekday := startDate.Weekday()
	daysUntilMonday := (int(weekday) - int(time.Monday) + 7) % 7
	monday := startDate.AddDate(0, 0, -daysUntilMonday)

	// Generate the lines for the whole project first.
	const metaHeight = 2
	var lines []string
	cursorRow := 0
	row := 0

	// Meta item: project start date (2 rows)
	{
		style := normalStyle
		if m.onMeta() {
			style = selectedStyle
			cursorRow = 0
		}
		label := "Project started: " + startDate.Format("Mon, Jan 2, 2006")
		lines = append(lines, style.Render(label))
		if m.isHoveringMeta() {
			lines = append(lines, dimStyle.Render("◀▶ h/l: ±1 day   ◀▶ H/L: ±1 week"))
		} else {
			lines = append(lines, style.Render(""))
		}
	}

	// If there are no items, show a hint
	if len(m.prj.items) == 0 {
		lines = append(lines, dimStyle.Render("There are no items in this plan. Press 'a' to add one."))
		return strings.Join(lines, "\n") + "\n"
	}

	for i, it := range m.prj.items {
		rightStyle := normalStyle
		leftStyle := fadeStyle
		if i == m.itemIndex() {
			rightStyle = selectedStyle
			leftStyle = normalStyle
			cursorRow = row + metaHeight
		}
		for w := range it.duration {

			// Get the date of the first monday, and the week of the year
			// TODO: Add start of week day to config.ini
			weekStart := monday.AddDate(0, 0, row*7)
			_, week := weekStart.ISOWeek()

			// Assemble the date, MM.DD
			// TODO: Add EU-style dates to config.ini
			date := fmt.Sprintf("%d.%d", int(weekStart.Month()), weekStart.Day())

			leftSide := fmt.Sprintf("W%-3d %-5s ", week, date)
			var rightSide string

			if w == 0 {
				if m.mode == editingTitle && i == m.itemIndex() {
					// IF EDITING: Show input.
					rightSide = "-" + m.input.View()
				} else if m.mode == confirmingDeletion && i == m.itemIndex() {
					// IF DELETING: Show confirmation.
					rightSide = "⬤  " + deleteStyle.Render("Delete? y/n")
				} else {
					// Otherwise just show the normal title.
					rightSide = "⬤  " + it.title
				}
			} else {
				rightSide += "⚬"
			}
			line := leftStyle.Render(leftSide) + rightStyle.Render(rightSide)

			lines = append(lines, line)
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
