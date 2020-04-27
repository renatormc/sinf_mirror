package main

import (
	"fmt"
	"math"
	"time"
)

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func fmtDuration(d time.Duration) string {
	secs := d.Seconds()
	hours := math.Floor(secs / 3600)
	secs = secs - hours*3600
	mins := math.Floor(secs / 60)
	secs = secs - mins*60
	return fmt.Sprintf("%dh %dmin %0.2fs", int(hours), int(mins), secs)
}
