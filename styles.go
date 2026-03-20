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
	primaryColor       = lipgloss.Color("#3fbf3f") // monstery green
	secondaryColor     = lipgloss.Color("#b0b0b0") // light gray
	gradientColorLeft  = lipgloss.Color("#7b2d8b") // dusky purple
	gradientColorRight = lipgloss.Color("#2d8b4e") // dark mossy green
	errorColor         = lipgloss.Color("#cc4444") // medium red
	warningColor       = lipgloss.Color("#ccaa22") // yellow
	logColor           = lipgloss.Color("#888888") // medium gray

	logoStyle         = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	phaseDoneStyle    = lipgloss.NewStyle().Foreground(logColor).Strikethrough(true)
	phaseStyle        = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	phasePendingStyle = lipgloss.NewStyle().Foreground(secondaryColor)
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
