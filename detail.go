package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// @region detail:model -- DETAIL VIEWMODEL + MODES

type detailCloseMsg struct{}
type detailHelpMsg struct{}
type detailSaveMsg struct{}
type detailItemCompletedMsg struct{}
type detailItemUncompletedMsg struct{}

type detailMode int

const (
	detailNormal detailMode = iota
	detailEditingDesc
	detailEditingTask
	detailConfirming
)

var (
	completionYes = []string{"Hell yes!", "Make it so!", "Forthwith!"}
	completionNo  = []string{"Nah.", "Maybe not.", "Not today, Satan."}
)

type detailViewModel struct {
	item             *milestone
	itemStart        time.Time
	isCurrent        bool
	taskCursor       int // index into item.tasks
	subCursor        int // index into the task's subtasks, or -1 when on the task itself
	mode             detailMode
	confirm          confirm
	textarea         textarea.Model
	input            textinput.Model
	preEditTitle     string // original title for esc revert
	isNewTask        bool   // true when editing a freshly inserted task/subtask
	panelWidth       int
	completionYesIdx int
	completionNoIdx  int
}

func makeDetailViewModel(it *milestone, panelWidth int, itemStart time.Time, isCurrent bool) detailViewModel {
	ta := textarea.New()
	ta.SetHeight(5)
	ta.ShowLineNumbers = true
	ta.Prompt = ""
	ta.KeyMap.InsertNewline = key.NewBinding(key.WithKeys("alt+enter"))
	ta.Placeholder = "Description..."
	ta.SetWidth(max(10, panelWidth-6))

	ti := textinput.New()
	ti.Placeholder = "Task title..."
	ti.CharLimit = 0

	d := detailViewModel{item: it, itemStart: itemStart, isCurrent: isCurrent, taskCursor: 0, subCursor: -1, textarea: ta, input: ti, panelWidth: panelWidth}
	d.cursorToFirstUnchecked()
	return d
}

// Go to first uncompleted task, and its first uncompleted subtask.
func (d *detailViewModel) cursorToFirstUnchecked() {
	for ti := range d.item.tasks {
		if !d.item.tasks[ti].completed {
			d.taskCursor = ti
			d.subCursor = -1
			t := &d.item.tasks[ti]
			if len(t.subtasks) > 0 {
				t.open = true
				for si := range t.subtasks {
					if !t.subtasks[si].completed {
						d.subCursor = si
						break
					}
				}
			}
			return
		}
	}
}

// @region detail:cursor -- TASK/SUBTASK CURSOR NAV (flatRows)

func (d *detailViewModel) onTask() bool { return d.subCursor < 0 }

// navigable row: a task (sub == -1) or one of its subtasks.
type rowRef struct {
	task int
	sub  int
}

// all tasks + all subtasks on open tasks, in a flat list.
func (d *detailViewModel) flatRows() []rowRef {
	var rows []rowRef
	for ti := range d.item.tasks {
		rows = append(rows, rowRef{ti, -1})
		if d.item.tasks[ti].open {
			for si := range d.item.tasks[ti].subtasks {
				rows = append(rows, rowRef{ti, si})
			}
		}
	}
	return rows
}

// steps cursor through the flattened task/subtask list.
func (d *detailViewModel) moveCursor(delta int) {
	rows := d.flatRows()
	if len(rows) == 0 {
		d.taskCursor = 0
		d.subCursor = -1
		return
	}
	cur := 0
	for i, r := range rows {
		if r.task == d.taskCursor && r.sub == d.subCursor {
			cur = i
			break
		}
	}
	cur = max(0, min(cur+delta, len(rows)-1))
	d.taskCursor = rows[cur].task
	d.subCursor = rows[cur].sub
}

func (d *detailViewModel) curTitle() string {
	if d.subCursor >= 0 {
		return d.item.tasks[d.taskCursor].subtasks[d.subCursor].title
	}
	return d.item.tasks[d.taskCursor].title
}

func (d *detailViewModel) setCurTitle(s string) {
	if d.subCursor >= 0 {
		d.item.tasks[d.taskCursor].subtasks[d.subCursor].title = s
	} else {
		d.item.tasks[d.taskCursor].title = s
	}
}

// @region detail:mutate -- INSERT/DELETE/PROMOTE TASKS

func (d *detailViewModel) deleteCurrent() {
	if d.subCursor >= 0 {
		t := &d.item.tasks[d.taskCursor]
		si := d.subCursor
		t.subtasks = append(t.subtasks[:si], t.subtasks[si+1:]...)
		if d.subCursor >= len(t.subtasks) {
			d.subCursor = len(t.subtasks) - 1 // -1 falls back onto the task row
		}
	} else {
		idx := d.taskCursor
		d.item.tasks = append(d.item.tasks[:idx], d.item.tasks[idx+1:]...)
		if d.taskCursor >= len(d.item.tasks) && d.taskCursor > 0 {
			d.taskCursor--
		}
	}
}

func (d *detailViewModel) insertTaskAt(idx int) tea.Cmd {
	idx = max(0, min(idx, len(d.item.tasks)))
	d.item.tasks = append(d.item.tasks, task{})
	copy(d.item.tasks[idx+1:], d.item.tasks[idx:])
	d.item.tasks[idx] = task{}
	d.taskCursor = idx
	d.subCursor = -1
	d.startEditingNew()
	return textinput.Blink
}

func (d *detailViewModel) insertSubtaskAt(taskIdx, subIdx int) tea.Cmd {
	t := &d.item.tasks[taskIdx]
	t.open = true
	subIdx = max(0, min(subIdx, len(t.subtasks)))
	t.subtasks = append(t.subtasks, subtask{})
	copy(t.subtasks[subIdx+1:], t.subtasks[subIdx:])
	t.subtasks[subIdx] = subtask{}
	d.taskCursor = taskIdx
	d.subCursor = subIdx
	d.startEditingNew()
	return textinput.Blink
}

func (d *detailViewModel) startEditingNew() {
	d.preEditTitle = ""
	d.isNewTask = true
	d.mode = detailEditingTask
	d.input.SetValue("")
	d.input.Focus()
}

func (d *detailViewModel) applyConfirm() tea.Cmd {
	switch d.confirm.kind {
	case confirmDeleteItem:
		d.deleteCurrent()
		return func() tea.Msg { return detailSaveMsg{} }
	case confirmToggleCompletion:
		if d.item.finished.IsZero() {
			d.item.finished = time.Now()
			return func() tea.Msg { return detailItemCompletedMsg{} }
		}
		d.item.finished = time.Time{}
		return func() tea.Msg { return detailItemUncompletedMsg{} }
	case confirmCompleteSubtasks:
		t := &d.item.tasks[d.taskCursor]
		t.completed = true
		for i := range t.subtasks {
			t.subtasks[i].completed = true
		}
		return func() tea.Msg { return detailSaveMsg{} }
	}
	return nil
}

// @region detail:input -- DETAIL UPDATE()

func (d detailViewModel) Update(msg tea.Msg) (detailViewModel, tea.Cmd) {

	switch d.mode {

	case detailEditingDesc:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "enter":
				d.item.description = d.textarea.Value()
				d.mode = detailNormal
				d.textarea.Blur()
				return d, func() tea.Msg { return detailSaveMsg{} }
			case "esc":
				d.mode = detailNormal
				d.textarea.Blur()
				return d, nil
			}
		}
		var cmd tea.Cmd
		d.textarea, cmd = d.textarea.Update(msg)
		return d, cmd

	case detailEditingTask:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "enter":
				d.setCurTitle(d.input.Value())
				d.mode = detailNormal
				d.input.Blur()
				return d, func() tea.Msg { return detailSaveMsg{} }
			case "esc":
				d.mode = detailNormal
				d.input.Blur()
				d.setCurTitle(d.preEditTitle)
				if d.isNewTask {
					d.deleteCurrent()
				}
				return d, nil
			}
		}
		var cmd tea.Cmd
		d.input, cmd = d.input.Update(msg)
		return d, cmd

	case detailConfirming:
		switch d.confirm.handle(msg) {
		case confirmYes:
			cmd := d.applyConfirm()
			d.mode = detailNormal
			d.confirm = confirm{}
			return d, cmd
		case confirmNo:
			d.mode = detailNormal
			d.confirm = confirm{}
			return d, nil
		}
		return d, nil

	case detailNormal:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "q", "ctrl+c":
				return d, tea.Quit
			case "?":
				return d, func() tea.Msg { return detailHelpMsg{} }
			case "esc":
				return d, func() tea.Msg { return detailCloseMsg{} }
			case "left", "h":
				// On a subtask: close the parent task and return to it.
				// On a task: leave the detail panel.
				if !d.onTask() {
					d.item.tasks[d.taskCursor].open = false
					d.subCursor = -1
					return d, nil
				}
				return d, func() tea.Msg { return detailCloseMsg{} }
			case "right", "l":
				// On a task with subtasks: open it and drop onto the first subtask.
				if d.onTask() && len(d.item.tasks) > 0 && len(d.item.tasks[d.taskCursor].subtasks) > 0 {
					d.item.tasks[d.taskCursor].open = true
					d.subCursor = 0
				}
			case "enter":
				d.mode = detailEditingDesc
				d.textarea.SetValue(d.item.description)
				d.textarea.SetWidth(max(10, d.panelWidth-6))
				cmd := d.textarea.Focus()
				return d, cmd
			case "up", "k":
				d.moveCursor(-1)
			case "down", "j":
				d.moveCursor(1)
			case "shift+up", "K":
				if !d.onTask() {
					si := d.subCursor
					if si > 0 {
						subs := d.item.tasks[d.taskCursor].subtasks
						subs[si], subs[si-1] = subs[si-1], subs[si]
						d.subCursor--
						return d, func() tea.Msg { return detailSaveMsg{} }
					}
				} else {
					idx := d.taskCursor
					if idx > 0 {
						d.item.tasks[idx], d.item.tasks[idx-1] = d.item.tasks[idx-1], d.item.tasks[idx]
						d.taskCursor--
						return d, func() tea.Msg { return detailSaveMsg{} }
					}
				}
			case "shift+down", "J":
				if !d.onTask() {
					si := d.subCursor
					subs := d.item.tasks[d.taskCursor].subtasks
					if si < len(subs)-1 {
						subs[si], subs[si+1] = subs[si+1], subs[si]
						d.subCursor++
						return d, func() tea.Msg { return detailSaveMsg{} }
					}
				} else {
					idx := d.taskCursor
					if idx < len(d.item.tasks)-1 {
						d.item.tasks[idx], d.item.tasks[idx+1] = d.item.tasks[idx+1], d.item.tasks[idx]
						d.taskCursor++
						return d, func() tea.Msg { return detailSaveMsg{} }
					}
				}
			case "shift+left", "H":
				// Promote a subtask to a task immediately below its parent.
				if !d.onTask() {
					ti := d.taskCursor
					si := d.subCursor
					t := &d.item.tasks[ti]
					st := t.subtasks[si]
					t.subtasks = append(t.subtasks[:si], t.subtasks[si+1:]...)
					newTask := task{title: st.title, completed: st.completed}
					at := ti + 1
					d.item.tasks = append(d.item.tasks, task{})
					copy(d.item.tasks[at+1:], d.item.tasks[at:])
					d.item.tasks[at] = newTask
					d.taskCursor = at
					d.subCursor = -1
					return d, func() tea.Msg { return detailSaveMsg{} }
				}
			case "shift+right", "L":
				// Demote a childless task into a subtask of the task above it.
				if d.onTask() && len(d.item.tasks) > 0 {
					ti := d.taskCursor
					if ti == 0 || len(d.item.tasks[ti].subtasks) > 0 {
						break
					}
					cur := d.item.tasks[ti]
					above := &d.item.tasks[ti-1]
					above.open = true
					above.subtasks = append(above.subtasks, subtask{title: cur.title, completed: cur.completed})
					d.item.tasks = append(d.item.tasks[:ti], d.item.tasks[ti+1:]...)
					d.taskCursor = ti - 1
					d.subCursor = len(d.item.tasks[ti-1].subtasks) - 1
					return d, func() tea.Msg { return detailSaveMsg{} }
				}
			case " ", "x":
				if len(d.item.tasks) == 0 {
					break
				}
				if !d.onTask() {
					st := &d.item.tasks[d.taskCursor].subtasks[d.subCursor]
					st.completed = !st.completed
					return d, func() tea.Msg { return detailSaveMsg{} }
				}
				t := &d.item.tasks[d.taskCursor]
				// Completing a task with unfinished subtasks asks first.
				if !t.completed && len(t.subtasks) > 0 && !t.allSubtasksDone() {
					prompt := fmt.Sprintf("Complete %d subtasks?", t.incompleteSubtasks())
					d.confirm = newConfirm(confirmCompleteSubtasks, prompt)
					d.mode = detailConfirming
					return d, nil
				}
				t.completed = !t.completed
				return d, func() tea.Msg { return detailSaveMsg{} }
			case "e":
				if len(d.item.tasks) > 0 {
					d.preEditTitle = d.curTitle()
					d.isNewTask = false
					d.mode = detailEditingTask
					d.input.SetValue(d.curTitle())
					d.input.CursorEnd()
					d.input.Focus()
					return d, textinput.Blink
				}
			case "d":
				if len(d.item.tasks) > 0 {
					prompt := "Delete?"
					if d.onTask() {
						if n := len(d.item.tasks[d.taskCursor].subtasks); n > 0 {
							prompt = fmt.Sprintf("delete w/ %d children?", n)
						}
					}
					d.confirm = newConfirm(confirmDeleteItem, prompt)
					d.mode = detailConfirming
				}
			case "a":
				if !d.onTask() {
					return d, d.insertSubtaskAt(d.taskCursor, len(d.item.tasks[d.taskCursor].subtasks))
				}
				return d, d.insertTaskAt(len(d.item.tasks))
			case "A":
				if d.onTask() {
					return d, d.insertSubtaskAt(d.taskCursor, 0)
				}
			case "o":
				if !d.onTask() {
					return d, d.insertSubtaskAt(d.taskCursor, d.subCursor+1)
				}
				idx := d.taskCursor + 1
				if len(d.item.tasks) == 0 {
					idx = 0
				}
				return d, d.insertTaskAt(idx)
			case "O":
				if !d.onTask() {
					return d, d.insertSubtaskAt(d.taskCursor, d.subCursor)
				}
				idx := d.taskCursor
				if len(d.item.tasks) == 0 {
					idx = 0
				}
				return d, d.insertTaskAt(idx)
			case "c":
				d.confirm = newConfirm(confirmToggleCompletion, "")
				d.completionYesIdx = rand.Intn(len(completionYes))
				d.completionNoIdx = rand.Intn(len(completionNo))
				d.mode = detailConfirming
				return d, nil
			case "-":
				if !d.item.finished.IsZero() {
					d.item.finished = d.item.finished.AddDate(0, 0, -1)
					return d, func() tea.Msg { return detailSaveMsg{} }
				}
			case "=":
				if !d.item.finished.IsZero() {
					d.item.finished = d.item.finished.AddDate(0, 0, 1)
					return d, func() tea.Msg { return detailSaveMsg{} }
				}
			case "_":
				if !d.item.finished.IsZero() {
					d.item.finished = d.item.finished.AddDate(0, 0, -7)
					return d, func() tea.Msg { return detailSaveMsg{} }
				}
			case "+":
				if !d.item.finished.IsZero() {
					d.item.finished = d.item.finished.AddDate(0, 0, 7)
					return d, func() tea.Msg { return detailSaveMsg{} }
				}
			}
		}
	}

	return d, nil
}

// @region detail:body -- DETAIL BODY RENDERING (getBody)

func renderProgressBar(width int, ratio float64, active bool) string {
	opts := []progress.Option{progress.WithWidth(width)}
	if active {
		opts = append(opts, progress.WithScaledGradient("#7d3483", "#ff5fd7")) // darker purple -> primary (206)
	} else {
		opts = append(opts, progress.WithSolidFill("#767676")) // dim gray (243)
	}
	return progress.New(opts...).ViewAs(ratio)
}

func getBody(item *milestone, dv *detailViewModel, width, height int, itemStart time.Time, isCurrent bool) string {
	title := titleStyle.Render(item.title)
	active := dv != nil

	var desc string
	if active && dv.mode == detailEditingDesc {
		desc = dv.textarea.View()
	} else if len(item.description) == 0 {
		desc = dimStyle.Italic(true).Render("~ no description ~")
	} else {
		desc = fadeStyle.Render(item.description)
	}

	selectedStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(textColor)
	confirmStyle := lipgloss.NewStyle().Foreground(warningColor).Bold(true)

	checkbox := func(completed, selected bool) string {
		box := "- [ ] "
		if completed {
			box = "- [x] "
		}
		switch {
		case selected:
			return selectedStyle.Render(box)
		case completed:
			return dimStyle.Render(box)
		default:
			return primaryStyle.Render(box)
		}
	}

	// renders the title, or the inline edit/confirm widget when the
	// cursor is on this row.
	titleCell := func(t string, completed, selected bool, ti, si int) string {
		if active && selected && dv.mode == detailConfirming &&
			(dv.confirm.kind == confirmDeleteItem || dv.confirm.kind == confirmCompleteSubtasks) {
			return confirmStyle.Render(dv.confirm.prompt + " y/n")
		}
		if active && dv.mode == detailEditingTask && dv.taskCursor == ti && dv.subCursor == si {
			return dv.input.View()
		}
		switch {
		case selected:
			return selectedStyle.Render(t)
		case completed:
			return dimStyle.Italic(true).Render(t)
		default:
			return normalStyle.Render(t)
		}
	}

	// Build the task list as discrete logical lines so the body can be
	// windowed around the cursor (see assembly + clipping below). cursorTaskLine
	// is the index, within taskLines, of the currently selected row.
	var taskLines []string
	cursorTaskLine := -1
	if len(item.tasks) == 0 {
		if active {
			taskLines = append(taskLines, dimStyle.Italic(true).Render("No tasks; a to add one"))
		}
	} else {
		for ti := range item.tasks {
			t := &item.tasks[ti]
			taskSelected := active && dv.taskCursor == ti && dv.subCursor == -1

			var line strings.Builder
			line.WriteString(checkbox(t.completed, taskSelected))

			if len(t.subtasks) > 0 {
				count := fmt.Sprintf("(%d/%d) ", t.completedSubtasks(), len(t.subtasks))
				if t.open {
					line.WriteString(primaryStyle.Render(count))
				} else {
					line.WriteString(dimStyle.Render(count))
				}
			}

			line.WriteString(titleCell(t.title, t.completed, taskSelected, ti, -1))
			if taskSelected {
				cursorTaskLine = len(taskLines)
			}
			taskLines = append(taskLines, line.String())

			if t.open {
				for si := range t.subtasks {
					st := &t.subtasks[si]
					subSelected := active && dv.taskCursor == ti && dv.subCursor == si
					var sl strings.Builder
					sl.WriteString("  ")
					sl.WriteString(checkbox(st.completed, subSelected))
					sl.WriteString(titleCell(st.title, st.completed, subSelected, ti, si))
					if subSelected {
						cursorTaskLine = len(taskLines)
					}
					taskLines = append(taskLines, sl.String())
				}
			}
		}
	}

	itemStatus := ""
	if active && dv.mode == detailConfirming && dv.confirm.kind == confirmToggleCompletion {
		if !item.finished.IsZero() {
			itemStatus = confirmStyle.Render("Unmark complete? y/n")
		} else {
			line1 := lipgloss.NewStyle().Foreground(doneColor).Render("Mark this item complete as of today?")
			line2 := confirmStyle.Render("y. " + completionYes[dv.completionYesIdx])
			line3 := dimStyle.Render("n. " + completionNo[dv.completionNoIdx])
			itemStatus = fmt.Sprintf("%s\n\n%s\n%s", line1, line2, line3)
		}
	} else if item.finished.IsZero() {
		endDate := itemStart.AddDate(0, 0, item.duration*7-1)
		daysUntil := int(time.Until(endDate).Hours() / 24)
		endStyle := dimStyle
		if endDate.Before(time.Now()) {
			endStyle = warningStyle
		}
		itemStatus = endStyle.Render("Due: " + fmtFullDate(endDate))
		if isCurrent {
			dU := fmt.Sprintf("%dd", daysUntil)
			if daysUntil == 0 {
				dU = "Today"
			} else if daysUntil < 0 {
				dU = fmt.Sprintf("%d past", -daysUntil)
			}
			itemStatus += dimStyle.Render(fmt.Sprintf(" (%s)", dU))
			aw := item.actualWeeks(itemStart)
			if aw > item.duration {
				overdueStyle := lipgloss.NewStyle().Foreground(errorColor)
				itemStatus += "\n" + dimStyle.Render(fmt.Sprintf("Estimated: %dw", item.duration))
				itemStatus += "\n" + overdueStyle.Render(fmt.Sprintf("Actual: %dw", aw))
			}
		}
		if active {
			itemStatus += "\n\n" + dimStyle.Render("~ hit c to mark this item complete. ~")
		}
	} else {
		itemStatus = doneStyle.Render(checkmark + " Completed on " + item.finished.Format("Jan 2, 2006"))

		estimated := fmt.Sprintf("Estimated: %dw", item.duration)
		aw := item.actualWeeks(itemStart)
		var actual string
		if aw < 1 {
			actual = "Actual: <1w"
		} else {
			actual = fmt.Sprintf("Actual: %dw", aw)
		}
		itemStatus += "\n" + dimStyle.Render(estimated)
		itemStatus += "\n" + dimStyle.Render(actual)
		if active {
			itemStatus += "\n\n" + dimStyle.Render("-+ change date, shift: by week")
		}
	}

	// Progress block: shown between the description and the task list when the
	// item has tasks and is not yet complete.
	var progressBlock string
	if len(item.tasks) > 0 && item.finished.IsZero() {
		done := 0
		for i := range item.tasks {
			if item.tasks[i].completed {
				done++
			}
		}
		incomplete := len(item.tasks) - done
		bar := renderProgressBar(max(10, width-6), float64(done)/float64(len(item.tasks)), active)

		// The rate/status line under the bar is only meaningful for the active
		// (current) milestone.
		progressBlock = bar
		var rate string
		if incomplete == 0 {
			rate = "All subtasks complete!"
		} else {
			endDate := itemStart.AddDate(0, 0, item.duration*7-1)
			weekdaysLeft := weekdaysBetween(time.Now(), endDate)
			if weekdaysLeft < 1 {
				weekdaysLeft = 1
			}
			perDay := float64(incomplete) / float64(weekdaysLeft)
			if perDay < 1 {
				rate = fmt.Sprintf("%.1f weekdays per task", 1/perDay)
			} else {
				rate = fmt.Sprintf("%.1f tasks per weekday", perDay)
			}
			if isCurrent {
				rate += " remaining"
			}
			progressBlock = bar + "\n" + dimStyle.Render(rate)
		}
	}

	// Assemble the body as a flat list of logical lines, tracking which line
	// holds the cursor. A blank string entry is a spacer (matches the old
	// "\n\n" separators).
	var logical []string
	cursorLogical := -1
	addBlock := func(s string) { logical = append(logical, strings.Split(s, "\n")...) }

	addBlock(title)
	logical = append(logical, "")
	addBlock(desc)
	if progressBlock != "" {
		logical = append(logical, "")
		addBlock(progressBlock)
	}
	logical = append(logical, "")
	taskStart := len(logical)
	logical = append(logical, taskLines...)
	if cursorTaskLine >= 0 {
		cursorLogical = taskStart + cursorTaskLine
	}
	logical = append(logical, "")
	addBlock(itemStatus)

	// Expand each logical line into the display lines it actually occupies once
	// wrapped at the panel's content width, so long/multiline rows are counted
	// correctly. contentWidth mirrors detailStyle: Width(w-2) minus horizontal
	// padding of 2 on each side.
	contentWidth := max(10, width-6)
	wrap := lipgloss.NewStyle().Width(contentWidth)
	var display []string
	cursorDisplay := -1
	for i, ll := range logical {
		if i == cursorLogical {
			cursorDisplay = len(display)
		}
		display = append(display, strings.Split(wrap.Render(ll), "\n")...)
	}

	// Window the display lines so the cursor row stays visible and the content
	// never overflows the panel (which would push the header off-screen).
	// viewHeight matches detailStyle's Height(h-5).
	viewHeight := height - 5
	if viewHeight > 0 && len(display) > viewHeight {
		start := 0
		if cursorDisplay >= 0 {
			start = max(cursorDisplay-viewHeight/2, 0)
		}
		end := start + viewHeight
		if end > len(display) {
			end = len(display)
			start = max(0, end-viewHeight)
		}
		display = display[start:end]
	}

	return strings.Join(display, "\n")
}

func (d *detailViewModel) View(w, h int) string {
	body := getBody(d.item, d, w, h, d.itemStart, d.isCurrent)
	return detailStyle(w, h, true).Render(body)
}

func detailViewInactive(it *milestone, w, h int, itemStart time.Time, isCurrent bool) string {
	if it == nil {
		return ""
	}
	body := getBody(it, nil, w, h, itemStart, isCurrent)
	return detailStyle(w, h, false).Render(body)
}
