package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

//Marker stores a marker info
type Marker struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	Role    string `json:"role"`
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

func checkFolder(path string) *Marker {
	path = filepath.Join(path, ".sinf_mark.json")
	jsonFile, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var markers []Marker
	json.Unmarshal([]byte(byteValue), &markers)
	return nil
}

func getMarkers(folder string) []Marker {
	path := filepath.Join(folder, ".sinf_mark.json")
	jsonFile, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var markers []Marker
	json.Unmarshal([]byte(byteValue), &markers)
	return markers
}

func getFoldersFromCaseName(casename string) ([]string, string) {
	for _, drive := range getDrives() {
		fmt.Println((drive))
	}
	return nil, ""
}
