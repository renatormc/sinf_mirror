package main

import (
	"time"
)

const (
	day  = time.Minute * 60 * 24
	year = 365 * day
)

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

// func fmtDuration(d time.Duration) string {
// 	d = d.Round(time.Minute)
// 	h := d / time.Hour
// 	d -= h * time.Hour
// 	m := d / time.Minute
// 	return fmt.Sprintf("%02d:%02d", h, m)
// }

// func humanizeDuration(d time.Duration) string {
// 	if d < day {
// 		return d.String()
// 	}

// 	var b strings.Builder
// 	if d >= year {
// 		years := d / year
// 		fmt.Fprintf(&b, "%dy", years)
// 		d -= years * year
// 	}

// 	days := d / day
// 	d -= days * day
// 	fmt.Fprintf(&b, "%dd%s", days, d)

// 	return b.String()
// }
