package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

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

// Matches e.g. `abcd (1)` -> `(1)`
var durationRe = regexp.MustCompile(`\((\d+)\)\s*$`)

// Parses a project.md file into useable data
func parseItems(content string) []item {
	lines := strings.Split(content, "\n")
	var items []item
	var cur *item // currently processing this item

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// H1 starts a new item
		if strings.HasPrefix(line, "# ") {
			// Save previous item
			if cur != nil {
				cur.description = strings.TrimSpace(cur.description)
				items = append(items, *cur)
			}

			title := strings.TrimPrefix(line, "# ")
			duration := 1
			if m := durationRe.FindStringSubmatch(title); m != nil {
				duration, _ = strconv.Atoi(m[1])
				title = strings.TrimSpace(title[:len(title)-len(m[0])])
			}

			cur = &item{title: title, duration: duration}
			continue
		}

		if cur == nil {
			continue
		}

		// Skip blockquote lines immediately after title (before description/subtasks);
		// these contain generated metadata like dates and weeks, used for reading
		// the file in e.g. Obsidian.
		if strings.HasPrefix(line, "> ") && cur.description == "" && len(cur.subtasks) == 0 {
			continue
		}

		// Checklist item
		if strings.HasPrefix(line, "- [ ] ") {
			st := subtask{title: strings.TrimPrefix(line, "- [ ] ")}
			// Consume indented bullet lines as description
			for i+1 < len(lines) {
				next := lines[i+1]
				if (len(next) >= 2 && next[0] == ' ' && strings.HasPrefix(strings.TrimLeft(next, " \t"), "- ")) ||
					(len(next) >= 1 && next[0] == '\t' && strings.HasPrefix(strings.TrimLeft(next, " \t"), "- ")) {
					desc := strings.TrimLeft(next, " \t")
					desc = strings.TrimPrefix(desc, "- ")
					if st.description != "" {
						st.description += "\n"
					}
					st.description += desc
					i++
				} else {
					break
				}
			}
			cur.subtasks = append(cur.subtasks, st)
			continue
		}

		// Description text (only before subtasks start)
		if len(cur.subtasks) == 0 {
			if cur.description == "" && line == "" {
				continue // skip leading blank lines
			}
			if cur.description != "" {
				cur.description += "\n"
			}
			cur.description += line
		}
	}

	// Save last item
	if cur != nil {
		cur.description = strings.TrimSpace(cur.description)
		items = append(items, *cur)
	}

	return items
}

func saveProject(p project) error {
	var b strings.Builder

	for i, it := range p.items {
		if i > 0 {
			b.WriteString("\n")
		}

		// Title as H1 with duration
		if it.duration != 1 {
			b.WriteString("# " + it.title + " (" + strconv.Itoa(it.duration) + ")\n")
		} else {
			b.WriteString("# " + it.title + "\n")
		}

		// Description
		if it.description != "" {
			b.WriteString("\n" + it.description + "\n")
		}

		// Subtasks
		if len(it.subtasks) > 0 {
			b.WriteString("\n")
			for _, st := range it.subtasks {
				b.WriteString("- [ ] " + st.title + "\n")
				if st.description != "" {
					for _, dl := range strings.Split(st.description, "\n") {
						b.WriteString("    - " + dl + "\n")
					}
				}
			}
		}
	}

	return os.WriteFile(p.filePath, []byte(b.String()), 0644)
}

func (p *project) save() error {
	return saveProject(*p)
}

func loadProject(fp string) (*project, error) {
	data, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	items := parseItems(string(data))
	prj := project{filePath: fp, items: items, usesTimeline: true}

	return &prj, nil
}
