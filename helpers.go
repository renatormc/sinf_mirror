package main

import (
	"fmt"
	"time"
)

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%02d:%02d", h, m)
}
