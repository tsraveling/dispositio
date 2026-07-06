# dispositio

Dispositio is a terminal tool for planning large projects in simple markdown.

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

## Usage

Simply run `dispositio` in any folder. It will prompt to create ROADMAP.md (if it doesn't already exist); your project will be stored in this file in simple Markdown.

Alternatively, you can do e.g. `dispositio OTHER_FILENAME.md` if you would like to use a different filename.

Once in dispositio, you can hit `?` to see a keymap guide wherever you are in the app.
