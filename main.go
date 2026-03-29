package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func printHelp() {
	// TODO: Fill out help stuff here
}

// resolveFilePath turns the optional CLI arg into a .md file path.
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

// promptCreate asks the user whether to create a missing file.
// Returns true if the user confirmed and the file was created.
func promptCreate(path string) bool {
	fmt.Printf("File %s does not exist. Create it? [y/n] ", path)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(line)) != "y" {
		return false
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		os.MkdirAll(dir, 0o755)
	}
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		fmt.Printf("Error creating file: %s\n", err.Error())
		return false
	}
	return true
}

func main() {

	// Load the config file
	cfg = readConfig()

	// Parse flags and positional args
	var positional string
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			printHelp()
			return
		default:
			positional = arg
		}
	}

	filename := resolveFilePath(positional)

	// If the file doesn't exist, prompt to create it
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
	p.Run()
}
