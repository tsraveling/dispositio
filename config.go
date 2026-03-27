package main

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

type config struct {

	// Config file
	someValue string

	// Window dimensions
	ww int
	wh int
}

func (c *config) fullWidth() int {
	return c.ww - 8
}

// TODO: Make these configurable
func (c *config) updateWW(ww int) {
	c.ww = max(30, ww)
}

func (c *config) updateWH(wh int) {
	c.wh = max(10, wh)
}

var cfg config

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func readConfig() config {
	// Get home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	// Path to config
	configPath := filepath.Join(homeDir, ".config", "dispositio", "config.ini")

	// Load the INI file
	cfg_file, err := ini.Load(configPath)
	if err != nil {
		panic(err)
	}

	// Read values
	ret := config{ww: 30}
	section := cfg_file.Section("general")

	// FIXME: If we don't have anything to stick in here, just delete the whole thing
	ret.someValue = expandPath(section.Key("someValue").String())

	return ret
}
