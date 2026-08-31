// Dispositio is a terminal tool for planning large projects in simple markdown.
//
// A project lives in an ordinary Markdown file, ROADMAP.md by default. Each
// milestone is an H1, its tasks are checkboxes, and indented checkboxes are
// subtasks. The file stays readable and editable outside the app, so it can sit
// in a repository alongside the work it describes.
//
// Milestones carry a duration in weeks, written as a parenthesised suffix on
// the header, and dispositio lays them out on a timeline from the project start
// date. Finishing a milestone early pulls the rest forward; overrunning pushes
// them back.
//
// Usage:
//
//	dispositio             open ./ROADMAP.md, offering to create it
//	dispositio DIR         open DIR/ROADMAP.md
//	dispositio FILE.md     open a specific file
//	dispositio -h          show help
//	dispositio -v          print version
//
// Press ? inside the app for the full keymap.
package main
