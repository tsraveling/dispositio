package main

import (
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

func (d detailViewModel) View(w, h int) string {
	title := titleStyle.Render(d.item.title)
	return detailStyle(w, h, true).Render(title)
}

func detailViewInactive(it *item, w, h int) string {
	content := ""
	if it != nil {
		content = titleStyle.Render(it.title)
	}
	return detailStyle(w, h, false).Render(content)
}
