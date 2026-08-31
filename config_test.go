package main

import "testing"

func TestConfigClampsWindowDimensions(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
		wantW, wantH  int
	}{
		{"typical terminal", 120, 40, 120, 40},
		{"narrow width clamps up", 10, 40, 30, 40},
		{"short height clamps up", 120, 2, 120, 10},
		{"zero clamps both", 0, 0, 30, 10},
		{"negative clamps both", -5, -5, 30, 10},
		{"exactly at the minimums", 30, 10, 30, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c config
			c.updateWW(tt.width)
			c.updateWH(tt.height)
			if c.ww != tt.wantW {
				t.Errorf("ww = %d, want %d", c.ww, tt.wantW)
			}
			if c.wh != tt.wantH {
				t.Errorf("wh = %d, want %d", c.wh, tt.wantH)
			}
		})
	}
}
