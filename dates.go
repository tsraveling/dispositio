package main

import "time"

// weekdaysBetween counts the weekdays (Mon–Fri) in the inclusive range
// [from, to], skipping Saturdays and Sundays. Returns 0 if to is before from.
func weekdaysBetween(from, to time.Time) int {
	n := 0
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		if wd := d.Weekday(); wd != time.Saturday && wd != time.Sunday {
			n++
		}
	}
	return n
}
