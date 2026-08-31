package main

// @region app:config -- CONFIG + WINDOW DIMENSIONS

type config struct {
	// Window dimensions
	ww int
	wh int
}

// TODO: Make these configurable
func (c *config) updateWW(ww int) {
	c.ww = max(30, ww)
}

func (c *config) updateWH(wh int) {
	c.wh = max(10, wh)
}

var cfg config
