package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
)

// renders HELP.md to a scrollable screen on a terminal, or plain text when
// piped or redirected.
func printHelp() {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Println(renderHelp(helpSource, maxWidth))
		return
	}
	p := tea.NewProgram(makeHelpScreen(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error showing help:", err)
	}
}

// @region app:files -- FILE PATH RESOLUTION + CREATE PROMPT

// turns the optional CLI arg into a .md file path.
//   - no arg: ./ROADMAP.md
//   - directory: <dir>/ROADMAP.md
//   - .md file: use as-is
func resolveFilePath(arg string) string {
	if arg == "" {
		return "./ROADMAP.md"
	}
	info, err := os.Stat(arg)
	if err == nil && info.IsDir() {
		return filepath.Join(arg, "ROADMAP.md")
	}
	return arg
}

// asks the user whether to create a missing file;
// true if confirmed and created.
func promptCreate(path string) bool {
	fmt.Printf("File %s does not exist. Create it? [y/n] ", path)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(line)) != "y" {
		return false
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Printf("Error creating directory: %s\n", err.Error())
			return false
		}
	}
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		fmt.Printf("Error creating file: %s\n", err.Error())
		return false
	}
	return true
}

// @region app:entry -- CLI ARGS + MAIN()

func main() {

	// TODO: Integrate config at some point?
	// cfg = readConfig()

	var positional string
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			printHelp()
			return
		case "-v", "--version":
			fmt.Println(Version)
			return
		default:
			positional = arg
		}
	}

	filename := resolveFilePath(positional)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if !promptCreate(filename) {
			return
		}
	}

	prj, err := loadProject(filename)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}

	var m tea.Model
	m, _ = makePlannerViewModel(prj)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
