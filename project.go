package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type subtask struct {
	title     string
	completed bool
}

type task struct {
	title     string
	completed bool
	subtasks  []subtask
	open      bool // ui only
}

// allSubtasksDone reports whether every subtask is completed. A task with no
// subtasks is trivially done.
func (t task) allSubtasksDone() bool {
	for _, st := range t.subtasks {
		if !st.completed {
			return false
		}
	}
	return true
}

// completedSubtasks returns the number of completed subtasks.
func (t task) completedSubtasks() int {
	n := 0
	for _, st := range t.subtasks {
		if st.completed {
			n++
		}
	}
	return n
}

// incompleteSubtasks returns the number of subtasks not yet completed.
func (t task) incompleteSubtasks() int {
	return len(t.subtasks) - t.completedSubtasks()
}

type milestone struct {
	title       string
	duration    int // in weeks
	description string
	tasks       []task
	finished    time.Time
}

// dateString returns e.g. "(3.14)" for a completed item, or "" if not finished.
func (i *milestone) dateString() string {
	if i.finished.IsZero() {
		return ""
	}
	return fmt.Sprintf("(%d.%d)", int(i.finished.Month()), i.finished.Day())
}

// actualWeeks returns the number of weeks from itemStart to the item's end
// (finished date, or now if still in progress).
func (i *milestone) actualWeeks(itemStart time.Time) int {
	end := i.finished
	if end.IsZero() {
		end = time.Now()
	}
	days := int(end.Sub(itemStart).Hours() / 24)
	weeks := days / 7
	if days%7 > 0 {
		weeks++
	}
	return weeks
}

// actualDuration returns the effective number of weeks this item occupies:
// the greater of its planned duration and the actual weeks taken.
func (i *milestone) actualDuration(itemStart time.Time) int {
	return max(i.duration, i.actualWeeks(itemStart))
}

// weeksRendered returns how many weeks this item occupies on the timeline,
// matching the planner's row count. Completed items that finished early
// occupy fewer weeks than planned; in-progress overruns extend. The week
// containing the finished date is counted, so finishing on a Monday gives
// the rest of that week as margin before the next item begins.
func (i *milestone) weeksRendered(itemStart time.Time) int {
	if i.finished.IsZero() {
		return i.actualDuration(itemStart)
	}
	days := int(i.finished.Sub(itemStart).Hours() / 24)
	if days < 0 {
		return 1
	}
	weeks := days/7 + 1
	if ad := i.actualDuration(itemStart); weeks > ad {
		weeks = ad
	}
	return weeks
}

// isCurrent returns true if the item at idx is the first non-completed item
// in the project — i.e. the one actively being worked on.
func (p *project) isCurrent(idx int) bool {
	for i, it := range p.items {
		if it.finished.IsZero() {
			return i == idx
		}
	}
	return false
}

// itemStartDate returns the Monday on which the given item (by index) begins.
// For completed items that ran past their planned duration, the actual weeks
// taken push subsequent items forward.
func (p *project) itemStartDate(idx int) time.Time {
	weekday := p.startDate.Weekday()
	daysUntilMonday := (int(weekday) - int(time.Monday) + 7) % 7
	monday := p.startDate.AddDate(0, 0, -daysUntilMonday)
	weekRow := 0
	for i, it := range p.items {
		if i == idx {
			break
		}
		start := monday.AddDate(0, 0, weekRow*7)
		weekRow += it.weeksRendered(start)
	}
	return monday.AddDate(0, 0, weekRow*7)
}

type project struct {
	filePath     string
	name         string
	startDate    time.Time // zero value means unset
	items        []milestone
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
func parseItems(lines []string) []milestone {
	var items []milestone
	var cur *milestone // currently processing this item

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

			cur = &milestone{title: title, duration: duration}

			// Check for item metadata code block (Finished date)
			if i+1 < len(lines) {
				if meta, consumed := parseCodeBlock(lines[i+1:]); consumed > 0 {
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

		// Skip blockquote lines immediately after title (before description/tasks);
		// these contain generated metadata like dates and weeks, used for reading
		// the file in e.g. Obsidian.
		if strings.HasPrefix(line, "> ") && cur.description == "" && len(cur.tasks) == 0 {
			continue
		}

		// Checklist item
		if strings.HasPrefix(line, "- [ ] ") || strings.HasPrefix(line, "- [x] ") {
			completed := strings.HasPrefix(line, "- [x] ")
			title := strings.TrimPrefix(line, "- [ ] ")
			if completed {
				title = strings.TrimPrefix(line, "- [x] ")
			}
			t := task{title: title, completed: completed}
			// Consume indented checklist lines as subtasks
			for i+1 < len(lines) {
				next := lines[i+1]
				indented := len(next) > 0 && (next[0] == ' ' || next[0] == '\t')
				trimmed := strings.TrimLeft(next, " \t")
				if indented && (strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "- [x] ")) {
					subCompleted := strings.HasPrefix(trimmed, "- [x] ")
					subTitle := strings.TrimPrefix(trimmed, "- [ ] ")
					if subCompleted {
						subTitle = strings.TrimPrefix(trimmed, "- [x] ")
					}
					t.subtasks = append(t.subtasks, subtask{title: subTitle, completed: subCompleted})
					i++
				} else {
					break
				}
			}
			cur.tasks = append(cur.tasks, t)
			continue
		}

		// Description text (only before tasks start)
		if len(cur.tasks) == 0 {
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

		// Item metadata code block (only if finished)
		if !it.finished.IsZero() {
			meta := map[string]string{"Finished": writeDate(it.finished)}
			b.WriteString(writeCodeBlock([]string{"Finished"}, meta))
		}

		// Description
		if it.description != "" {
			b.WriteString("\n" + it.description + "\n")
		}

		// Tasks
		if len(it.tasks) > 0 {
			b.WriteString("\n")
			for _, t := range it.tasks {
				checkbox := "- [ ] "
				if t.completed {
					checkbox = "- [x] "
				}
				b.WriteString(checkbox + t.title + "\n")
				// Subtasks: indented checklist items under the task
				for _, st := range t.subtasks {
					subbox := "- [ ] "
					if st.completed {
						subbox = "- [x] "
					}
					b.WriteString("  " + subbox + st.title + "\n")
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
