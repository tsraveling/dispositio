package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type detailCloseMsg struct{}
type detailSaveMsg struct{}

type detailMode int

const (
	detailNormal detailMode = iota
	detailEditingDesc
	detailEditingTask
	detailConfirmingDelete
)

type detailViewModel struct {
	item       *item
	taskCursor int
	mode       detailMode
	textarea   textarea.Model
	input      textinput.Model
	panelWidth int
}

func makeDetailViewModel(it *item, panelWidth int) detailViewModel {
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

	return detailViewModel{item: it, taskCursor: 0, textarea: ta, input: ti, panelWidth: panelWidth}
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
			case "d":
				if len(d.item.subtasks) > 0 {
					d.mode = detailConfirmingDelete
				}
			case "a":
				d.mode = detailEditingTask
				d.input.SetValue("")
				d.input.Focus()
				// Insert a placeholder subtask after current cursor
				insertIdx := d.taskCursor + 1
				if len(d.item.subtasks) == 0 {
					insertIdx = 0
				}
				newTask := subtask{}
				d.item.subtasks = append(d.item.subtasks, subtask{})
				copy(d.item.subtasks[insertIdx+1:], d.item.subtasks[insertIdx:])
				d.item.subtasks[insertIdx] = newTask
				d.taskCursor = insertIdx
				return d, textinput.Blink
			case "x":
				if len(d.item.subtasks) > 0 {
					d.item.subtasks[d.taskCursor].completed = !d.item.subtasks[d.taskCursor].completed
					return d, func() tea.Msg { return detailSaveMsg{} }
				}
			}
		}
	}

	return d, nil
}

func getBody(item *item, dv *detailViewModel) string {
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

	return fmt.Sprintf("%s\n\n%s\n\n%s", title, desc, b.String())
}

func (d *detailViewModel) View(w, h int) string {
	body := getBody(d.item, d)
	return detailStyle(w, h, true).Render(body)
}

func detailViewInactive(it *item, w, h int) string {
	if it == nil {
		return ""
	}
	body := getBody(it, nil)
	return detailStyle(w, h, false).Render(body)
}
