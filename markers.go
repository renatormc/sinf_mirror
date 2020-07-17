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
	Name    string `json:"name"`
	Subtype string `json:"subtype"`
	Role    string `json:"role"`
}

// type FoundFolders struct {
// 	Sources []string
// 	Dest    string
// }

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

func isCase(folder string, caseName string) (bool, *Marker) {
	markers := getMarkers(folder)
	for _, marker := range markers {
		if marker.Type == "case" && marker.Name == caseName {
			return true, &marker
		}
	}
	return false, nil
}

func isDrive(folder string) (bool, *Marker) {
	markers := getMarkers(folder)
	for _, marker := range markers {
		if marker.Type == "disk" {
			return true, &marker
		}
	}
	return false, nil
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

func (synchronizer *Synchronizer) scanFolder(folder string, depth int) {
	if depth > synchronizer.maxDepth {
		return
	}
	res, marker := isCase(folder, synchronizer.caseName)
	if res {
		switch marker.Role {
		case "temp":
			synchronizer.sources = append(synchronizer.sources, folder)
			break

		case "final":
			synchronizer.dest = folder
			break
		}

	} else {
		items, _ := ioutil.ReadDir(folder)
		for _, item := range items {
			if item.IsDir() {
				synchronizer.scanFolder(filepath.Join(folder, item.Name()), depth+1)
			}
		}
	}
}

func (synchronizer *Synchronizer) scanDrives() {
	fmt.Printf("Iniciando vasculhamento de pastas com profundidade máxima de %d\n", synchronizer.maxDepth)
	for _, drive := range getDrives() {
		res, _ := isDrive(drive)
		if res {
			fmt.Printf("Vasculhando drive %s\n", drive)
			synchronizer.scanFolder(drive, 1)
		}
	}
}
