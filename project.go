package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type itemStatus int

const (
	pending itemStatus = iota
	inProgress
	done
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
	started     time.Time
	finished    time.Time
}

// Returns pending, inProgress, or done, depending on which dates are present on the object.
func (i *item) status() itemStatus {
	if i.started.IsZero() {
		return pending
	}
	if i.finished.IsZero() {
		return inProgress
	}
	return done
}

type project struct {
	filePath     string
	name         string
	startDate    time.Time // zero value means unset
	items        []item
	usesTimeline bool
}

// Using the Go date formatting paradigm
const dateFormat = "Jan 2 2006"

func readDate(s string) (time.Time, error) {
	return time.Parse(dateFormat, s)
}

func writeDate(t time.Time) string {
	return t.Format(dateFormat)
}

// parseCodeBlock extracts key-value pairs from a fenced code block.
// Returns the map and the number of lines consumed (0 if no block found).
func parseCodeBlock(lines []string) (map[string]string, int) {
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "```" {
		return nil, 0
	}
	m := make(map[string]string)
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "```" {
			return m, i + 1
		}
		if k, v, ok := strings.Cut(line, ":"); ok {
			m[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	// Unterminated block — ignore it
	return nil, 0
}

// writeCodeBlock serializes a map into a fenced code block string.
// Keys are written in the order provided.
func writeCodeBlock(keys []string, m map[string]string) string {
	var b strings.Builder
	b.WriteString("```\n")
	for _, k := range keys {
		if v, ok := m[k]; ok {
			b.WriteString(k + ": " + v + "\n")
		}
	}
	b.WriteString("```\n")
	return b.String()
}

// Matches e.g. `abcd (1)` -> `(1)`
var durationRe = regexp.MustCompile(`\((\d+)\)\s*$`)

// parseProject parses the full file content into a project's metadata and items.
func parseProject(content string, prj *project) {
	lines := strings.Split(content, "\n")

	// Check for a leading code block with project metadata
	if meta, consumed := parseCodeBlock(lines); consumed > 0 {
		if v, ok := meta["Project Name"]; ok {
			prj.name = v
		}
		if v, ok := meta["Project Start"]; ok {
			if t, err := readDate(v); err == nil {
				prj.startDate = t
			}
		}

		// If there was a code block, use that as the cursor and parse items out of the rest of it
		lines = lines[consumed:]
	}

	// Now do the individual item parsing
	prj.items = parseItems(lines)
}

// Parses item lines into useable data
func parseItems(lines []string) []item {
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

			// Check for item metadata code block (Started / Finished dates)
			if i+1 < len(lines) {
				if meta, consumed := parseCodeBlock(lines[i+1:]); consumed > 0 {
					if v, ok := meta["Started"]; ok {
						if t, err := readDate(v); err == nil {
							cur.started = t
						}
					}
					if v, ok := meta["Finished"]; ok {
						if t, err := readDate(v); err == nil {
							cur.finished = t
						}
					}
					i += consumed
				}
			}
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

	// Write project metadata code block if any values are set
	if p.name != "" || !p.startDate.IsZero() {
		meta := make(map[string]string)
		var keys []string

		// Project name
		if p.name != "" {
			keys = append(keys, "Project Name")
			meta["Project Name"] = p.name
		}

		// Project start date
		if !p.startDate.IsZero() {
			keys = append(keys, "Project Start")
			meta["Project Start"] = writeDate(p.startDate)
		}
		b.WriteString(writeCodeBlock(keys, meta))
		b.WriteString("\n")
	}

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

		// Item metadata code block (only if any dates are set)
		if !it.started.IsZero() || !it.finished.IsZero() {
			meta := make(map[string]string)
			var keys []string
			if !it.started.IsZero() {
				keys = append(keys, "Started")
				meta["Started"] = writeDate(it.started)
			}
			if !it.finished.IsZero() {
				keys = append(keys, "Finished")
				meta["Finished"] = writeDate(it.finished)
			}
			b.WriteString(writeCodeBlock(keys, meta))
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

	prj := project{filePath: fp, usesTimeline: true}
	parseProject(string(data), &prj)

	// Default start date to today if not set
	if prj.startDate.IsZero() {
		prj.startDate = time.Now()
	}

	return &prj, nil
}
