package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func printHelp() {
	// TODO: Fill out help stuff here
}

func main() {

	// Load the config file
	cfg = readConfig()

	if len(os.Args) > 1 && os.Args[1] == "init" {
		initProject()
		return
	}

	// Parse flags and positional args
	dir := "."
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			printHelp()
			return
		case "--new-only":
			sesh.NewOnly = true
		case "--nuke":
			sesh.Nuke = true
		default:
			dir = arg
		}
	}

	// Validate: --new-only and --nuke are mutually exclusive
	if sesh.NewOnly && sesh.Nuke {
		fmt.Println("Error: --new-only and --nuke cannot be used together.")
		os.Exit(1)
	}

	err := loadProject(dir)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}

	var m tea.Model
	m, _ = makeSomeModel()

	p := tea.NewProgram(m)
	p.Run()
}
