package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	// "github.com/charmbracelet/lipgloss"
)

type plannerViewModel struct {
	prj       project
	someValue string
}

func makePlannerViewModel(p *project) (plannerViewModel, tea.Cmd) {
	// Copy project into value mode so we can mutate it bubbletea-style
	m := plannerViewModel{prj: *p, someValue: "example"}
	return m, m.Init()
}

func (m plannerViewModel) Init() tea.Cmd {
	return nil
}

func (m plannerViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	//	case tea.WindowSizeMsg:
	//		m.list.SetWidth(msg.Width)

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	// Text input gets the end of it
	// var cmd tea.Cmd
	// m.input, cmd = m.input.Update(msg)
	// return m, cmd

	return m, nil
}

func (m plannerViewModel) View() string {
	now := time.Now()

	// Find the Monday of the current week
	weekday := now.Weekday()
	daysUntilMonday := (int(weekday) - int(time.Monday) + 7) % 7
	monday := now.AddDate(0, 0, -daysUntilMonday)

	var sb strings.Builder
	for i := 0; i < 20; i++ {
		weekStart := monday.AddDate(0, 0, i*7)
		_, week := weekStart.ISOWeek()
		date := fmt.Sprintf("%d.%d", int(weekStart.Month()), weekStart.Day())
		sb.WriteString(fmt.Sprintf("W%-3d %-5s ⚬\n", week, date))
	}
	return sb.String()
}

// var (
// 	titleStyle = lipgloss.NewStyle().
// 			Bold(true).
// 			Foreground(lipgloss.Color("205")).
// 			MarginLeft(2)
//
// 	itemStyle = lipgloss.NewStyle().
// 			PaddingLeft(4)
// )
