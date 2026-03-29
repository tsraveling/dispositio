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
	editingProjectName
	confirmingDeletion
)

type plannerViewModel struct {
	prj          project
	cursor       int // 0 = meta (start date), 1..N = items[0..N-1]
	mode         plannerMode
	detail       *detailViewModel
	input        textinput.Model
	preEditTitle string // original title before editing, for esc revert
	isNewItem    bool   // true when editing a newly added item (delete on esc)
	currentModal modal
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

	// Start cursor on the current item (first non-completed),
	// or the last item if everything is completed.
	m.cursor = len(m.prj.items) // default: last item in cursor-space
	for i, it := range m.prj.items {
		if it.finished.IsZero() {
			m.cursor = i + 1
			break
		}
	}

	return m, m.Init()
}

func (m plannerViewModel) Init() tea.Cmd {
	return nil
}

func (m *plannerViewModel) detailPanelWidth() int {
	if cfg.ww >= minSideBySideWidth {
		return cfg.ww - cfg.ww/2
	}
	return cfg.ww
}

func (m *plannerViewModel) gotoDetail() {
	if !m.onMeta() {
		idx := m.itemIndex()
		dvm := makeDetailViewModel(&m.prj.items[idx], m.detailPanelWidth(), m.prj.itemStartDate(idx), m.prj.isCurrent(idx))
		m.detail = &dvm
	}
}

// insertItemAt inserts a new empty item at the given index in prj.items
// and enters editing mode.
func (m *plannerViewModel) insertItemAt(idx int) tea.Cmd {
	newItem := item{title: "", duration: 1}

	if len(m.prj.items) == 0 {
		m.prj.items = []item{newItem}
		idx = 0
	} else {
		idx = max(0, min(idx, len(m.prj.items)))
		m.prj.items = append(m.prj.items, item{})
		copy(m.prj.items[idx+1:], m.prj.items[idx:])
		m.prj.items[idx] = newItem
	}

	m.cursor = idx + 1 // convert to cursor-space
	m.preEditTitle = ""
	m.isNewItem = true
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
		if m.detail != nil {
			m.detail.panelWidth = m.detailPanelWidth()
		}
	default:
		_ = msg
	}

	// Handle detail messages.
	if _, ok := msg.(detailCloseMsg); ok {
		m.detail = nil
		return m, nil
	}
	if _, ok := msg.(detailSaveMsg); ok {
		m.prj.save()
		return m, nil
	}

	if _, ok := msg.(detailItemCompletedMsg); ok {
		// Find the index of the completed item
		completedIdx := -1
		for i := range m.prj.items {
			if &m.prj.items[i] == m.detail.item {
				completedIdx = i
				break
			}
		}

		if completedIdx >= 0 {
			// Find the first active (non-finished) item
			activeIdx := -1
			for i, it := range m.prj.items {
				if it.finished.IsZero() {
					activeIdx = i
					break
				}
			}

			// If the completed item is after the active item, move it just before
			if activeIdx >= 0 && completedIdx > activeIdx {
				completed := m.prj.items[completedIdx]
				m.prj.items = append(m.prj.items[:completedIdx], m.prj.items[completedIdx+1:]...)
				m.prj.items = append(m.prj.items[:activeIdx], append([]item{completed}, m.prj.items[activeIdx:]...)...)
				m.cursor = activeIdx + 1 // cursor-space
			}
		}

		m.detail = nil
		m.prj.save()
		return m, nil
	}

	if _, ok := msg.(detailItemUncompletedMsg); ok {
		// When un-completing an item that sits before other completed items,
		// move it so it comes after the last completed item. This keeps all
		// finished items grouped at the top of the list.
		uncompletedIdx := -1
		for i := range m.prj.items {
			if &m.prj.items[i] == m.detail.item {
				uncompletedIdx = i
				break
			}
		}

		if uncompletedIdx >= 0 {
			// Find the last completed item
			lastFinishedIdx := -1
			for i, it := range m.prj.items {
				if !it.finished.IsZero() {
					lastFinishedIdx = i
				}
			}

			// If there are completed items after this one, move it after them
			if lastFinishedIdx > uncompletedIdx {
				uncompleted := m.prj.items[uncompletedIdx]
				m.prj.items = append(m.prj.items[:uncompletedIdx], m.prj.items[uncompletedIdx+1:]...)
				m.prj.items = append(m.prj.items[:lastFinishedIdx], append([]item{uncompleted}, m.prj.items[lastFinishedIdx:]...)...)
				m.cursor = lastFinishedIdx + 1 // cursor-space
			}
		}

		m.detail = nil
		m.prj.save()
		return m, nil
	}

	// Route to active modal; always consumes input when present.
	if m.currentModal != nil {
		var cmd tea.Cmd
		m.currentModal, cmd = modalUpdate(m.currentModal, msg)
		return m, cmd
	}

	// Route to detail panel if active.
	if m.detail != nil {
		var cmd tea.Cmd
		*m.detail, cmd = m.detail.Update(msg)
		return m, cmd
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
				m.prj.items[idx].title = m.preEditTitle
				if m.isNewItem {
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

	// SECTION: Project Name Editing Mode
	case editingProjectName:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.mode = normal
				m.input.Blur()
				return m, nil
			case "enter":
				m.prj.name = m.input.Value()
				m.mode = normal
				m.input.Blur()
				m.prj.save()
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
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
				return m, m.insertItemAt(len(m.prj.items))
			case "o":
				idx := m.itemIndex() + 1 // after current; if on meta, inserts at 0
				if m.onMeta() {
					idx = 0
				}
				return m, m.insertItemAt(idx)
			case "O":
				idx := m.itemIndex() // before current
				if m.onMeta() {
					idx = 0
				}
				return m, m.insertItemAt(idx)
			case "e":
				if m.onMeta() {
					m.mode = editingProjectName
					m.input.SetValue(m.prj.name)
					m.input.CursorEnd()
					m.input.Focus()
					return m, textinput.Blink
				} else {
					idx := m.itemIndex()
					m.preEditTitle = m.prj.items[idx].title
					m.isNewItem = false
					m.mode = editingTitle
					m.input.SetValue(m.prj.items[idx].title)
					m.input.CursorEnd()
					m.input.Focus()
					return m, textinput.Blink
				}
			case "d":
				if !m.onMeta() && len(m.prj.items) > 0 {
					m.mode = confirmingDeletion
				}
			case "M":
				if !m.onMeta() {
					m.currentModal = newCompleteItemModal(&m.prj.items[m.itemIndex()])
					return m, nil
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
			case "enter":
				m.gotoDetail()
			case "right", "l":
				if m.isHoveringMeta() {
					m.prj.startDate = m.prj.startDate.AddDate(0, 0, 1)
					m.prj.save()
				} else if !m.onMeta() {
					m.gotoDetail()
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

func (m plannerViewModel) plannerView() string {

	// Set up the local styles for this view
	selectedStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(textColor)
	fadeStyle := lipgloss.NewStyle().Foreground(fadeColor)
	deleteStyle := lipgloss.NewStyle().Foreground(warningColor).Bold(true)
	nowYear, nowWeek := time.Now().ISOWeek()

	// Use the project start date as the base Monday
	startDate := m.prj.startDate
	weekday := startDate.Weekday()
	daysUntilMonday := (int(weekday) - int(time.Monday) + 7) % 7
	monday := startDate.AddDate(0, 0, -daysUntilMonday)

	// Generate the lines for the whole project first.
	lines := []string{"", ""}
	cursorRow := 0

	row := 0     // Row for cursor
	weekRow := 0 // Row for date calc

	// Meta item: project title + start date
	panelWidth := cfg.ww/2 - 4 // account for padding; narrow mode uses full width
	if cfg.ww < minSideBySideWidth {
		panelWidth = cfg.ww - 4
	}
	{
		style := normalStyle
		if m.onMeta() {
			style = selectedStyle
			cursorRow = len(lines)
		}

		// Project title
		titleText := strings.ToUpper(m.prj.name)
		if m.mode == editingProjectName {
			lines = append(lines, m.input.View())
		} else if titleText != "" {
			pad := max(0, (panelWidth-len(titleText))/2)
			lines = append(lines, strings.Repeat(" ", pad)+primaryStyle.Render(titleText))
			lines = append(lines, strings.Repeat(" ", pad)+primaryStyle.Render(strings.Repeat("=", len(titleText))))
		} else if m.isHoveringMeta() {
			lines = append(lines, dimStyle.Render("e to set project name"))
		}

		// Extra space between title and start date
		lines = append(lines, "")

		// Start date
		label := "Project started: " + startDate.Format("Mon, Jan 2, 2006")
		lines = append(lines, style.Render(label))
		if m.isHoveringMeta() {
			lines = append(lines, dimStyle.Render("◀▶ h/l: ±1 day   ◀▶ H/L: ±1 week"))
		} else {
			lines = append(lines, "")
		}
		lines = append(lines, "")
	}
	itemsStart := len(lines) // line index where items begin

	// If there are no items, show a hint
	if len(m.prj.items) == 0 {
		lines = append(lines, dimStyle.Render("There are no items in this plan. Press 'a' to add one."))
		return strings.Join(lines, "\n") + "\n"
	}

	// Iterate through the items in the project
	for i, it := range m.prj.items {
		if i == m.itemIndex() {
			cursorRow = row + itemsStart
		}

		isCurrent := m.prj.isCurrent(i)
		itemStart := monday.AddDate(0, 0, weekRow*7)
		renderWeeks := it.actualDuration(itemStart)

		// Iterate for as many weeks as the item should be rendered
		for w := range renderWeeks {

			// Get the date of the first monday, and the week of the year
			// TODO: Add start of week day to config.ini
			weekStart := monday.AddDate(0, 0, weekRow*7)
			wsYear, week := weekStart.ISOWeek()

			// If item is finished, stop after the week it was completed
			sameWeekFinish := false
			if !it.finished.IsZero() && weekStart.After(it.finished) {
				if w == 0 {
					// Multiple milestones finished in the same week:
					// render a single collapsed row instead of skipping.
					sameWeekFinish = true
				} else {
					break
				}
			}

			// Assemble the date, MM.DD
			// TODO: Add EU-style dates to config.ini
			date := fmt.Sprintf("%d.%d", int(weekStart.Month()), weekStart.Day())

			var leftSide string
			rightStyle := normalStyle
			leftStyle := fadeStyle
			if i == m.itemIndex() {
				rightStyle = selectedStyle
				leftStyle = normalStyle
			}

			if sameWeekFinish {
				leftSide = fmt.Sprintf("--+ %7s", "")
			} else {
				leftSide = fmt.Sprintf("W%-3d %-5s ", week, date)
				if wsYear == nowYear && week == nowWeek {
					leftStyle = highlightedStyle
				}
			}

			var rightSide string

			// Symbols: ✓ done, ⬤ current, ◯ pending, ⚠ overdue
			overdue := isCurrent && w >= it.duration
			symbol := "◯"
			if !it.finished.IsZero() {
				symbol = "✓"
				rightStyle = rightStyle.Italic(true)
			} else if isCurrent {
				symbol = "⬤"
			}

			if w == 0 {
				if m.mode == editingTitle && i == m.itemIndex() {
					// IF EDITING: Show input.
					rightSide = rightStyle.Render("-") + m.input.View()
				} else if m.mode == confirmingDeletion && i == m.itemIndex() {
					// IF DELETING: Show confirmation.
					rightSide = rightStyle.Render(symbol+"  ") + deleteStyle.Render("Delete? y/n")
				} else {
					// Otherwise just show the normal title.
					dateStr := ""
					if ds := it.dateString(); ds != "" {
						dateStr = " " + dimStyle.Render(ds)
					}
					rightSide = rightStyle.Render(symbol+"  "+it.title) + dateStr
				}
			} else if overdue {
				rightSide = rightStyle.Render("⚠")
			} else {
				rightSide = rightStyle.Render("⚬")
			}
			line := leftStyle.Render(leftSide) + rightSide

			lines = append(lines, line)
			row++
			if !sameWeekFinish {
				weekRow++
			}
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

func (m plannerViewModel) View() string {
	plannerStr := m.plannerView()

	if cfg.ww >= minSideBySideWidth {
		// Panel mode: always show both panels
		plannerWidth := cfg.ww / 2
		detailWidth := cfg.ww - plannerWidth
		plannerCol := lipgloss.NewStyle().Width(plannerWidth).Height(cfg.wh).PaddingLeft(2).Render(plannerStr)

		var detailCol string
		if m.detail != nil {
			detailCol = m.detail.View(detailWidth, cfg.wh)
		} else {
			var it *item
			var itemStart time.Time
			var isCurrent bool
			if !m.onMeta() {
				idx := m.itemIndex()
				it = &m.prj.items[idx]
				itemStart = m.prj.itemStartDate(idx)
				isCurrent = m.prj.isCurrent(idx)
			}
			detailCol = detailViewInactive(it, detailWidth, cfg.wh, itemStart, isCurrent)
		}

		combined := lipgloss.JoinHorizontal(lipgloss.Top, plannerCol, detailCol)
		return modalView(m.currentModal, combined)
	}

	// Narrow mode: swap views
	if m.detail != nil {
		detailStr := m.detail.View(cfg.ww, cfg.wh)
		return modalView(m.currentModal, detailStr)
	}

	paddedPlanner := lipgloss.NewStyle().PaddingLeft(2).Render(plannerStr)
	return modalView(m.currentModal, paddedPlanner)
}
