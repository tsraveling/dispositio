package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	maxWidth     = 120
	maxLogHeight = 25
)

var (
	primaryColor       = lipgloss.Color("206")
	secondaryColor     = lipgloss.Color("174") // light gray
	textColor          = lipgloss.Color("254") // white
	highlightColor     = lipgloss.Color("226")
	gradientColorLeft  = lipgloss.Color("#7b2d8b") // dusky purple
	gradientColorRight = lipgloss.Color("#2d8b4e") // dark mossy green
	errorColor         = lipgloss.Color("#cc4444") // medium red
	warningColor       = lipgloss.Color("#ccaa22") // yellow
	logColor           = lipgloss.Color("#888888") // medium gray

	primaryStyle     = lipgloss.NewStyle().Foreground(primaryColor)
	highlightedStyle = lipgloss.NewStyle().Foreground(highlightColor).Bold(true)
)

func boxWidth(termWidth int) int {
	return min(termWidth, maxWidth)
}

func errorBoxStyle(w int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(errorColor).
		Foreground(errorColor).
		Padding(1).
		Width(w - 2)
}

func outputBoxStyle(w int, done bool) lipgloss.Style {
	c := logColor
	if done {
		c = primaryColor
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(c).
		Foreground(c).
		Padding(1).
		Width(w - 2)
}

func clampLines(s string, max int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= max {
		return s
	}
	return strings.Join(lines[len(lines)-max:], "\n")
}
