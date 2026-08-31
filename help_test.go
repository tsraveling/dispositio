package main

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

// renderHelp output depends on lipgloss's terminal profile detection, which
// differs between a TTY and a test runner. Assert on the text, not the color.
func plain(s string) string { return ansiRe.ReplaceAllString(s, "") }

func TestHelpWordsDropsBackticks(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
		code []bool // code flag of each word's first rune
	}{
		{"plain", "one two", []string{"one", "two"}, []bool{false, false}},
		{"one span", "run `dispositio` now", []string{"run", "dispositio", "now"}, []bool{false, true, false}},
		{"span with a space splits into words", "`go test`", []string{"go", "test"}, []bool{true, true}},
		{"unclosed span still yields words", "trailing `code", []string{"trailing", "code"}, []bool{false, true}},
		{"empty", "", nil, nil},
		{"only spaces", "   ", nil, nil},
		{"collapses runs of spaces", "a    b", []string{"a", "b"}, []bool{false, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helpWords(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("helpWords(%q) gave %d words, want %d", tt.in, len(got), len(tt.want))
			}
			for i, w := range got {
				if string(w.runes) != tt.want[i] {
					t.Errorf("word %d = %q, want %q", i, string(w.runes), tt.want[i])
				}
				if w.code[0] != tt.code[i] {
					t.Errorf("word %d code flag = %v, want %v", i, w.code[0], tt.code[i])
				}
			}
		})
	}
}

func TestWrapAndStyle(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		width int
		want  string
	}{
		{"empty", "", 10, ""},
		{"width below one disables wrapping", "aaa bbb ccc", 0, "aaa bbb ccc"},
		{"negative width disables wrapping", "aaa bbb ccc", -5, "aaa bbb ccc"},
		{"fits on one line", "one two", 20, "one two"},
		{"wraps at boundary", "aaa bbb ccc", 7, "aaa bbb\nccc"},
		{"word longer than width is not split", "supercalifragilistic x", 5, "supercalifragilistic\nx"},
		{"backticks do not count toward width", "`aaa` `bbb`", 7, "aaa bbb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := plain(wrapAndStyle(tt.in, tt.width)); got != tt.want {
				t.Errorf("wrapAndStyle(%q, %d) = %q, want %q", tt.in, tt.width, got, tt.want)
			}
		})
	}
}

func TestRenderHelp(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		width int
		want  string
	}{
		{"empty input", "", 40, ""},
		{"h1 loses its marker", "# Title", 40, "Title"},
		{"h2 loses its marker", "## Section", 40, "Section"},
		{"blank lines preserved", "a\n\nb", 40, "a\n\nb"},
		{"bullet gets a glyph", "- an item", 40, "  • an item"},
		{"paragraph wraps", "aaa bbb ccc", 7, "aaa bbb\nccc"},
		{"code span is unwrapped from backticks", "use `x` here", 40, "use x here"},
		{"fence content is indented verbatim", "```\n# Not A Header\n```", 40, "  # Not A Header"},
		{"unterminated fence still renders content", "```\nfoo", 40, "  foo"},
		{"header inside fence is not styled as a header", "```\n## nope\n```", 40, "  ## nope"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := plain(renderHelp(tt.src, tt.width)); got != tt.want {
				t.Errorf("renderHelp(%q, %d) =\n%q\nwant\n%q", tt.src, tt.width, got, tt.want)
			}
		})
	}
}

// a long bullet must hang-indent its continuation lines under the glyph
func TestRenderHelpBulletHangingIndent(t *testing.T) {
	got := plain(renderHelp("- aaa bbb ccc ddd", 12))
	want := "  • aaa bbb\n    ccc ddd"
	if got != want {
		t.Errorf("got\n%q\nwant\n%q", got, want)
	}
}

// styling must not count toward the wrap width; a line with code spans should
// wrap at the same point as the same line without them
func TestRenderHelpWrapIgnoresStyling(t *testing.T) {
	withCode := plain(renderHelp("`aaa` `bbb` `ccc`", 7))
	without := plain(renderHelp("aaa bbb ccc", 7))
	if withCode != without {
		t.Errorf("code spans changed wrapping:\nwith:    %q\nwithout: %q", withCode, without)
	}
}

// the embedded file must be present and non-empty, whatever its content
func TestHelpSourceIsEmbedded(t *testing.T) {
	if strings.TrimSpace(helpSource) == "" {
		t.Fatal("HELP.md embedded as empty")
	}
}

// @region test:help-screen -- SCROLLABLE HELP SCREEN

func sizedHelpScreen(t *testing.T, w, h int) helpScreen {
	t.Helper()
	m, _ := makeHelpScreen().Update(tea.WindowSizeMsg{Width: w, Height: h})
	return m.(helpScreen)
}

func TestHelpScreenRendersAfterSizing(t *testing.T) {
	if got := makeHelpScreen().View(); got != "" {
		t.Errorf("unsized screen rendered %q, want empty", got)
	}
	// Assert on structure, not content — HELP.md is human-authored prose that
	// gets wrapped and scrolled, so nothing about its text is stable here.
	out := plain(sizedHelpScreen(t, 100, 40).View())
	if !strings.Contains(out, helpFooter) {
		t.Errorf("footer missing from view:\n%s", out)
	}
	if body := strings.TrimSpace(strings.Split(out, helpFooter)[0]); body == "" {
		t.Error("help body rendered empty above the footer")
	}
}

func TestHelpScreenClampsToMaxWidth(t *testing.T) {
	if got := sizedHelpScreen(t, 500, 40).vp.Width; got != maxWidth {
		t.Errorf("viewport width = %d, want %d", got, maxWidth)
	}
}

func TestHelpScreenKeepsMinimumHeight(t *testing.T) {
	if got := sizedHelpScreen(t, 100, 1).vp.Height; got < 5 {
		t.Errorf("viewport height = %d, want at least 5", got)
	}
}

func TestHelpScreenQuitKeys(t *testing.T) {
	for _, key := range []string{"q", "esc", "ctrl+c"} {
		t.Run(key, func(t *testing.T) {
			m := sizedHelpScreen(t, 100, 40)
			_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
			if key != "q" {
				typ := map[string]tea.KeyType{"esc": tea.KeyEsc, "ctrl+c": tea.KeyCtrlC}[key]
				_, cmd = m.Update(tea.KeyMsg{Type: typ})
			}
			if cmd == nil {
				t.Fatalf("%q returned no command, want tea.Quit", key)
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Errorf("%q returned %T, want tea.QuitMsg", key, cmd())
			}
		})
	}
}

func TestHelpScreenInitIsNoop(t *testing.T) {
	if cmd := makeHelpScreen().Init(); cmd != nil {
		t.Errorf("Init returned %v, want nil", cmd)
	}
}

// non-quit keys fall through to the viewport
func TestHelpScreenPassesScrollKeysToViewport(t *testing.T) {
	m := sizedHelpScreen(t, 100, 40)
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown}); cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Error("down key quit the screen")
		}
	}
}
