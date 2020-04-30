package main

import (
	"fmt"
	"math"
	"os"
	"syscall"
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

func bitsToDrives(bitMap uint32) (drives []string) {
	availableDrives := []string{"A:\\.", "B:\\.", "C:\\.", "D:\\.", "E:\\.", "F:\\.", "G:\\.", "H:\\.", "I:\\.", "J:\\.", "K:\\.", "L:\\.", "M:\\.", "N:\\.", "O:\\.", "P:\\.", "Q:\\.", "R:\\.", "S:\\.", "T:\\.", "U:\\.", "V:\\.", "W:\\.", "X:\\.", "Y:\\.", "Z:\\."}

	for i := range availableDrives {
		if bitMap&1 == 1 {
			drives = append(drives, availableDrives[i])
		}
		bitMap >>= 1
	}

	return
}

func getDrives() []string {
	var drives []string
	kernel32, _ := syscall.LoadLibrary("kernel32.dll")
	getLogicalDrivesHandle, _ := syscall.GetProcAddress(kernel32, "GetLogicalDrives")
	if ret, _, callErr := syscall.Syscall(uintptr(getLogicalDrivesHandle), 0, 0, 0, 0); callErr != 0 {
		checkError(callErr)
	} else {
		drives = bitsToDrives(uint32(ret))
	}
	return drives
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
