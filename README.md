# Dispositio

[![Tests](https://github.com/tsraveling/dispositio/actions/workflows/test.yml/badge.svg)](https://github.com/tsraveling/dispositio/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/tsraveling/dispositio/branch/main/graph/badge.svg)](https://codecov.io/gh/tsraveling/dispositio)
[![Go Report Card](https://goreportcard.com/badge/github.com/tsraveling/dispositio)](https://goreportcard.com/report/github.com/tsraveling/dispositio)
[![Go Reference](https://pkg.go.dev/badge/github.com/tsraveling/dispositio.svg)](https://pkg.go.dev/github.com/tsraveling/dispositio)

> **For more free software, or to check out my newsletter about useful tools and systems, head over to [The Systemist](https://systemist.net).**

Dispositio is a terminal tool for planning large projects in simple markdown.

![Demo](docs/demo.gif)

## Installation

### Homebrew (macOS)

```sh
brew install tsraveling/tap/dispositio
```

### Scoop (Windows)

```sh
scoop bucket add tsraveling https://github.com/tsraveling/scoop-bucket
scoop install dispositio
```

### Debian / Ubuntu / Fedora / Alpine

Download the `.deb`, `.rpm`, or `.apk` from the [latest release](https://github.com/tsraveling/dispositio/releases/latest) and install with your package manager, e.g.:

```sh
sudo dpkg -i dispositio_*.deb
```

### Go

```sh
go install github.com/tsraveling/dispositio@latest
```

### Manual

Download the archive for your platform from the [latest release](https://github.com/tsraveling/dispositio/releases/latest), extract, and place `dispositio` on your `PATH`.

## Detailed documentation

A full walkthrough with GIF examples can be found over at [The Systemist](https://systemist.net/tools/dispositio)!

## Quickstart

Simply run `dispositio` in any folder. It will prompt to create ROADMAP.md (if it doesn't already exist); your project will be stored in this file in simple Markdown.

Alternatively, you can do e.g. `dispositio OTHER_FILENAME.md` if you would like to use a different filename.

Once in dispositio, you can hit `?` to see a keymap guide wherever you are in the app.

You will see your empty project. From here:

- Hit `e` to set a project name -- this shows up at the top, and can be whatever you want.
- Hit `a` to add your first milestone.
- With that milestone selected, use shift+h/l or shift+right/left to change the duration of the milestone
- Hit enter/right/l to enter milestone details
- From there you can add tasks, subtasks, etc (see `?` popup for details).

Dispositio automatically creates a timeline. You can change the start date by selecting the first (project) row, and using h/l/right/left to change start date (hold shift to change by a week at a time).

This allows you to quickly map out a project timeline in milestones, with duration measured in weeks per milestone. Your **current milestone** is the topmost non-completed one, marked with a solid circle. The current milesone will "stretch" into overtime if you do not complete it on time, and your project timeline will update to match.
