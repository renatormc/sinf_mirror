package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"
	"syscall"
	"time"
)

// Win32FileAttributeData mantém atributos de arquivo do windows
type Win32FileAttributeData struct {
	FileAttributes uint32
	CreationTime   syscall.Filetime
	LastAccessTime syscall.Filetime
	LastWriteTime  syscall.Filetime
	FileSizeHigh   uint32
	FileSizeLow    uint32
}

func checkError(e error) {
	if e != nil {
		// panic(e)
		log.Fatal(e)
	}
}

func copyTimes(src string, dst string) {
	fi, err := os.Lstat(src)
	checkError(err)
	d := fi.Sys().(*syscall.Win32FileAttributeData)
	mTime := time.Unix(0, d.LastWriteTime.Nanoseconds())
	aTime := time.Unix(0, d.LastAccessTime.Nanoseconds())
	err = os.Chtimes(dst, aTime, mTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Não foi possível mudar os carimbos de hora do arquivo \"%s\"\n", dst)

	}
	// fmt.Println(d.CreationTime.Nanoseconds())
}

// func compareFiles(src string, dst string) bool {
// 	fiSrc, err := os.Lstat(src)
// 	checkError(err)
// 	fiDst, err := os.Lstat(dst)
// 	checkError(err)
// 	if fiSrc.FileSizeHigh != fiDst.FileSizeHigh ||  fiSrc.FileSizeLow != fiDst.FileSizeLow {
// 		return false
// 	}
// 	if fiSrc.LastWriteTime != fiDst.LastWriteTime {
// 		return false
// 	}
// 	return true
// }

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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getInputsFromFile(path string) ([]string, string) {
	file, err := os.Open(path)
	checkError(err)
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	text := strings.TrimSpace(string(b))
	lines := strings.Split(text, "\n")
	sources := strings.Split(lines[0], ",")
	for i := 0; i < len(sources); i++ {
		sources[i] = strings.TrimSpace(sources[i])
	}
	dest := strings.TrimSpace(lines[1])
	return sources, dest
}
