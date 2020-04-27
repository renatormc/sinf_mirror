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
	return fmt.Sprintf("%d:%02d:%02d", int(hours), int(mins), int(secs))
}

func fmtTime(t time.Time) string {
	return fmt.Sprintf("%d/%02d/%d %d:%02d:%02d", t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), t.Second())
}
