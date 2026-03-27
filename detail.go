package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type detailCloseMsg struct{}
type detailSaveMsg struct{}

type detailViewModel struct {
	item       *item
	taskCursor int
	editing    bool
	textarea   textarea.Model
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

	return detailViewModel{item: it, taskCursor: 0, textarea: ta, panelWidth: panelWidth}
}

func (d detailViewModel) Update(msg tea.Msg) (detailViewModel, tea.Cmd) {

	// Editing input handling
	if d.editing {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "enter":
				d.item.description = d.textarea.Value()
				d.editing = false
				d.textarea.Blur()
				return d, func() tea.Msg { return detailSaveMsg{} }
			case "esc":
				d.editing = false
				d.textarea.Blur()
				return d, nil
			}
		}
		var cmd tea.Cmd
		d.textarea, cmd = d.textarea.Update(msg)
		return d, cmd
	}

	// Main input handling
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "esc", "left", "h":
			return d, func() tea.Msg { return detailCloseMsg{} }
		case "enter":
			d.editing = true
			d.textarea.SetValue(d.item.description)
			d.textarea.SetWidth(max(10, d.panelWidth-6))
			cmd := d.textarea.Focus()
			return d, cmd
		}
	}
	return d, nil
}

func getBody(item *item, dv *detailViewModel) string {
	title := titleStyle.Render(item.title)

	var desc string
	if dv != nil && dv.editing {
		desc = dv.textarea.View()
	} else if len(item.description) == 0 {
		desc = dimStyle.Italic(true).Render("~ no description ~")
	} else {
		desc = fadeStyle.Render(item.description)
	}

	var b strings.Builder
	for _, task := range item.subtasks {
		prefix := primaryStyle.Render("- [ ] ")
		if task.completed {
			prefix = dimStyle.Render("- [x] ")
		}
		b.WriteString(prefix)
		b.WriteString(task.title)
		b.WriteString("\n")
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
