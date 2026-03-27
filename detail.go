package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type detailCloseMsg struct{}

type detailViewModel struct {
	item       *item
	taskCursor int
}

func makeDetailViewModel(it *item) detailViewModel {
	return detailViewModel{item: it, taskCursor: 0}
}

func (d detailViewModel) Update(msg tea.Msg) (detailViewModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "esc", "left", "h":
			return d, func() tea.Msg { return detailCloseMsg{} }
		}
	}
	return d, nil
}

func getBody(item *item, dv *detailViewModel) string {
	title := titleStyle.Render(item.title)
	desc := item.description
	if len(desc) == 0 {
		desc = dimStyle.Italic(true).Render("~ no description ~")
	} else {
		desc = fadeStyle.Render(desc)
	}

	// Do the tasks section
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

	// Put it all together
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
