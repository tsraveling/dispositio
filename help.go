package main

import (
	_ "embed"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//go:embed HELP.md
var helpSource string

var (
	helpH1     = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	helpH2     = lipgloss.NewStyle().Bold(true).Foreground(secondaryColor)
	helpCode   = lipgloss.NewStyle().Foreground(highlightColor)
	helpBullet = lipgloss.NewStyle().Foreground(dimColor)
	helpText   = lipgloss.NewStyle().Foreground(textColor)
)

// @region cli:markdown -- HELP MARKDOWN RENDERER

// a word carries a code-span flag per rune, so styling survives wrapping.
type helpWord struct {
	runes []rune
	code  []bool
}

func (w helpWord) render() string {
	var b strings.Builder
	for i := 0; i < len(w.runes); {
		j := i
		for j < len(w.runes) && w.code[j] == w.code[i] {
			j++
		}
		seg := string(w.runes[i:j])
		if w.code[i] {
			b.WriteString(helpCode.Render(seg))
		} else {
			b.WriteString(helpText.Render(seg))
		}
		i = j
	}
	return b.String()
}

// splits text into words, dropping backticks and recording which runes fell
// inside a `code` span. Dropping them here rather than at render time keeps
// them from counting toward the wrap width.
func helpWords(s string) []helpWord {
	var words []helpWord
	var cur helpWord
	inCode := false
	flush := func() {
		if len(cur.runes) > 0 {
			words = append(words, cur)
			cur = helpWord{}
		}
	}
	for _, r := range s {
		switch {
		case r == '`':
			inCode = !inCode
		case unicode.IsSpace(r):
			flush()
		default:
			cur.runes = append(cur.runes, r)
			cur.code = append(cur.code, inCode)
		}
	}
	flush()
	return words
}

// wraps to width on word boundaries, styling code spans as it goes. A width
// below 1 disables wrapping. Words longer than width are never split.
func wrapAndStyle(s string, width int) string {
	words := helpWords(s)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	var line strings.Builder
	lineWidth := 0

	for _, w := range words {
		switch {
		case lineWidth == 0:
			line.WriteString(w.render())
			lineWidth = len(w.runes)
		case width >= 1 && lineWidth+1+len(w.runes) > width:
			lines = append(lines, line.String())
			line.Reset()
			line.WriteString(w.render())
			lineWidth = len(w.runes)
		default:
			line.WriteString(helpText.Render(" ") + w.render())
			lineWidth += 1 + len(w.runes)
		}
	}
	lines = append(lines, line.String())
	return strings.Join(lines, "\n")
}

// minimal markdown renderer: headers, bullets, fenced blocks, and code spans.
func renderHelp(src string, width int) string {
	var out []string
	inFence := false

	for line := range strings.SplitSeq(src, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			out = append(out, helpCode.Render("  "+line))
			continue
		}

		switch {
		case trimmed == "":
			out = append(out, "")
		case strings.HasPrefix(trimmed, "## "):
			out = append(out, helpH2.Render(strings.TrimPrefix(trimmed, "## ")))
		case strings.HasPrefix(trimmed, "# "):
			out = append(out, helpH1.Render(strings.TrimPrefix(trimmed, "# ")))
		case strings.HasPrefix(trimmed, "- "):
			body := wrapAndStyle(strings.TrimPrefix(trimmed, "- "), width-4)
			indented := strings.ReplaceAll(body, "\n", "\n    ")
			out = append(out, helpBullet.Render("  • ")+indented)
		default:
			out = append(out, wrapAndStyle(trimmed, width))
		}
	}

	return strings.Join(out, "\n")
}

// @region cli:help -- SCROLLABLE HELP SCREEN

const helpFooter = "↑↓ scroll  q quit  —  press ? inside dispositio for the keymap"

type helpScreen struct {
	vp    viewport.Model
	ready bool
}

func makeHelpScreen() helpScreen {
	return helpScreen{vp: viewport.New(maxWidth, 24)}
}

func (m helpScreen) Init() tea.Cmd { return nil }

func (m helpScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := boxWidth(msg.Width)
		m.vp.Width = w
		m.vp.Height = max(msg.Height-4, 5)
		m.vp.SetContent(renderHelp(helpSource, w-4))
		m.ready = true
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m helpScreen) View() string {
	if !m.ready {
		return ""
	}
	body := m.vp.View() + "\n" + dimStyle.Render(helpFooter)
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
}
