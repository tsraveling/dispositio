package main

type subtask struct {
	title       string
	description string
}

type item struct {
	title       string
	duration    int // in weeks
	description string
	subtasks    []subtask
}

type project struct {
	filePath     string
	items        []item
	usesTimeline bool
}

var prj project

func loadProject(fp string) error {
	// data, err := os.ReadFile(fp)
	// if err != nil {
	// 	return err
	// }

	prj = project{filePath: fp, usesTimeline: true}

	return nil
}
