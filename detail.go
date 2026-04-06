package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type detailCloseMsg struct{}
type detailSaveMsg struct{}
type detailItemCompletedMsg struct{}
type detailItemUncompletedMsg struct{}

type detailMode int

const (
	detailNormal detailMode = iota
	detailEditingDesc
	detailEditingTask
	detailConfirmingDelete
	detailChangingCompletion
)

var (
	completionYes = []string{"Hell yes!", "Make it so!", "Forthwith!"}
	completionNo  = []string{"Nah.", "Maybe not.", "Not today, Satan."}
)

type detailViewModel struct {
	item             *item
	itemStart        time.Time
	isCurrent        bool
	taskCursor       int
	mode             detailMode
	textarea         textarea.Model
	input            textinput.Model
	preEditTitle     string // original subtask title for esc revert
	isNewSubtask     bool   // true when editing a newly added subtask
	panelWidth       int
	completionYesIdx int
	completionNoIdx  int
}

func makeDetailViewModel(it *item, panelWidth int, itemStart time.Time, isCurrent bool) detailViewModel {
	ta := textarea.New()
	ta.SetHeight(5)
	ta.ShowLineNumbers = true
	ta.Prompt = ""
	ta.KeyMap.InsertNewline = key.NewBinding(key.WithKeys("alt+enter"))
	ta.Placeholder = "Description..."
	ta.SetWidth(max(10, panelWidth-6))

	ti := textinput.New()
	ti.Placeholder = "Subtask title..."
	ti.CharLimit = 120

	return detailViewModel{item: it, itemStart: itemStart, isCurrent: isCurrent, taskCursor: 0, textarea: ta, input: ti, panelWidth: panelWidth}
}

// insertSubtaskAt inserts a new empty subtask at the given index and enters editing mode.
func (d *detailViewModel) insertSubtaskAt(idx int) tea.Cmd {
	idx = max(0, min(idx, len(d.item.subtasks)))
	d.preEditTitle = ""
	d.isNewSubtask = true
	d.mode = detailEditingTask
	d.input.SetValue("")
	d.input.Focus()
	newTask := subtask{}
	d.item.subtasks = append(d.item.subtasks, subtask{})
	copy(d.item.subtasks[idx+1:], d.item.subtasks[idx:])
	d.item.subtasks[idx] = newTask
	d.taskCursor = idx
	return textinput.Blink
}

func (d detailViewModel) Update(msg tea.Msg) (detailViewModel, tea.Cmd) {

	switch d.mode {

	// Description editing
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

	// Editing a subtask title (add or rename)
	case detailEditingTask:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "enter":
				d.item.subtasks[d.taskCursor].title = d.input.Value()
				d.mode = detailNormal
				d.input.Blur()
				return d, func() tea.Msg { return detailSaveMsg{} }
			case "esc":
				d.mode = detailNormal
				d.input.Blur()
				d.item.subtasks[d.taskCursor].title = d.preEditTitle
				if d.isNewSubtask {
					d.item.subtasks = append(d.item.subtasks[:d.taskCursor], d.item.subtasks[d.taskCursor+1:]...)
					if d.taskCursor >= len(d.item.subtasks) && d.taskCursor > 0 {
						d.taskCursor--
					}
				}
				return d, nil
			}
		}
		var cmd tea.Cmd
		d.input, cmd = d.input.Update(msg)
		return d, cmd

	// Confirming subtask deletion
	case detailConfirmingDelete:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y":
				idx := d.taskCursor
				d.item.subtasks = append(d.item.subtasks[:idx], d.item.subtasks[idx+1:]...)
				if d.taskCursor >= len(d.item.subtasks) && d.taskCursor > 0 {
					d.taskCursor--
				}
				d.mode = detailNormal
				return d, func() tea.Msg { return detailSaveMsg{} }
			case "n", "esc":
				d.mode = detailNormal
				return d, nil
			}
		}
		return d, nil

	// Changing completion status
	case detailChangingCompletion:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y":
				if d.item.finished.IsZero() {
					d.item.finished = time.Now()
					d.mode = detailNormal
					return d, func() tea.Msg { return detailItemCompletedMsg{} }
				}
				d.item.finished = time.Time{}
				d.mode = detailNormal
				return d, func() tea.Msg { return detailItemUncompletedMsg{} }
			case "n", "esc":
				d.mode = detailNormal
				return d, nil
			}
		}
		return d, nil

	// Normal detail navigation
	case detailNormal:
		if msg, ok := msg.(tea.KeyMsg); ok {
			maxTask := len(d.item.subtasks) - 1
			switch msg.String() {
			case "esc", "left", "h":
				return d, func() tea.Msg { return detailCloseMsg{} }
			case "enter":
				d.mode = detailEditingDesc
				d.textarea.SetValue(d.item.description)
				d.textarea.SetWidth(max(10, d.panelWidth-6))
				cmd := d.textarea.Focus()
				return d, cmd
			case "up", "k":
				if d.taskCursor > 0 {
					d.taskCursor--
				}
			case "down", "j":
				if d.taskCursor < maxTask {
					d.taskCursor++
				}
			case "shift+up", "K":
				idx := d.taskCursor
				if idx > 0 {
					d.item.subtasks[idx], d.item.subtasks[idx-1] = d.item.subtasks[idx-1], d.item.subtasks[idx]
					d.taskCursor--
					return d, func() tea.Msg { return detailSaveMsg{} }
				}
			case "shift+down", "J":
				idx := d.taskCursor
				if idx < maxTask {
					d.item.subtasks[idx], d.item.subtasks[idx+1] = d.item.subtasks[idx+1], d.item.subtasks[idx]
					d.taskCursor++
					return d, func() tea.Msg { return detailSaveMsg{} }
				}
			case " ":
				idx := d.taskCursor
				if len(d.item.subtasks) > idx {
					d.item.subtasks[idx].completed = !d.item.subtasks[idx].completed
					return d, func() tea.Msg { return detailSaveMsg{} }
				}
			case "e":
				if len(d.item.subtasks) > 0 {
					d.preEditTitle = d.item.subtasks[d.taskCursor].title
					d.isNewSubtask = false
					d.mode = detailEditingTask
					d.input.SetValue(d.item.subtasks[d.taskCursor].title)
					d.input.CursorEnd()
					d.input.Focus()
					return d, textinput.Blink
				}
			case "d":
				if len(d.item.subtasks) > 0 {
					d.mode = detailConfirmingDelete
				}
			case "a":
				return d, d.insertSubtaskAt(len(d.item.subtasks))
			case "o":
				idx := d.taskCursor + 1
				if len(d.item.subtasks) == 0 {
					idx = 0
				}
				return d, d.insertSubtaskAt(idx)
			case "O":
				idx := d.taskCursor
				if len(d.item.subtasks) == 0 {
					idx = 0
				}
				return d, d.insertSubtaskAt(idx)
			case "c":
				d.mode = detailChangingCompletion
				d.completionYesIdx = rand.Intn(len(completionYes))
				d.completionNoIdx = rand.Intn(len(completionNo))
				return d, nil
			case "x":
				if len(d.item.subtasks) > 0 {
					d.item.subtasks[d.taskCursor].completed = !d.item.subtasks[d.taskCursor].completed
					return d, func() tea.Msg { return detailSaveMsg{} }
				}
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

func getBody(item *item, dv *detailViewModel, itemStart time.Time, isCurrent bool) string {
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
	deleteStyle := lipgloss.NewStyle().Foreground(warningColor).Bold(true)

	var b strings.Builder
	if len(item.subtasks) == 0 {
		if active {
			b.WriteString(dimStyle.Italic(true).Render("No subtasks; a to add one"))
			b.WriteString("\n")
		}
	} else {
		for i, task := range item.subtasks {
			isSelected := active && i == dv.taskCursor

			// Checkbox prefix
			if task.completed {
				if isSelected {
					b.WriteString(selectedStyle.Render("- [x] "))
				} else {
					b.WriteString(dimStyle.Render("- [x] "))
				}
			} else {
				if isSelected {
					b.WriteString(selectedStyle.Render("- [ ] "))
				} else {
					b.WriteString(primaryStyle.Render("- [ ] "))
				}
			}

			// Title
			if active && dv.mode == detailConfirmingDelete && isSelected {
				b.WriteString(deleteStyle.Render("Delete? y/n"))
			} else if active && dv.mode == detailEditingTask && i == dv.taskCursor {
				b.WriteString(dv.input.View())
			} else if isSelected {
				b.WriteString(selectedStyle.Render(task.title))
			} else if task.completed {
				b.WriteString(dimStyle.Italic(true).Render(task.title))
			} else {
				b.WriteString(normalStyle.Render(task.title))
			}
			b.WriteString("\n")
		}
	}

	confirmStyle := lipgloss.NewStyle().Foreground(warningColor).Bold(true)

	itemStatus := ""
	if active && dv.mode == detailChangingCompletion {
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

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s", title, desc, b.String(), itemStatus)
}

func (d *detailViewModel) View(w, h int) string {
	body := getBody(d.item, d, d.itemStart, d.isCurrent)
	return detailStyle(w, h, true).Render(body)
}

func detailViewInactive(it *item, w, h int, itemStart time.Time, isCurrent bool) string {
	if it == nil {
		return ""
	}
	body := getBody(it, nil, itemStart, isCurrent)
	return detailStyle(w, h, false).Render(body)
}
