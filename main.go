package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func printHelp() {
	// TODO: Fill out help stuff here
}

func main() {

	// Load the config file
	cfg = readConfig()

	// Parse flags and positional args
	filename := "./ROADMAP.md"
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			printHelp()
			return
		default:
			filename = arg
		}
	}

	err := loadProject(filename)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}

	var m tea.Model
	m, _ = makeSomeModel()

	p := tea.NewProgram(m)
	p.Run()
}
