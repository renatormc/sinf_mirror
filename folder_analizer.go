package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// FolderMarker armazena informações sobre a pasta
// type FolderMarker struct {
// 	Type string `json:"type"`
// 	Role string `json:"role"`
// 	Name string `json:"name"`
// }

// FolderAnalyzer _
type FolderAnalyzer struct {
	Name        string
	Sources     []string
	Destination string
	Deep        int
	MaxDeep     int
}

func (folderAnalyzer *FolderAnalyzer) init() {
	folderAnalyzer.Sources = make([]string, 0, 10)
	folderAnalyzer.Destination = ""
}

func checkFolder(path string) (map[string]interface{}, bool) {
	var result map[string]interface{}
	pathJSON := filepath.Join(path, ".sinf_mark.json")
	_, err := os.Stat(pathJSON)
	if err != nil {
		return result, false
	}
	file, err := ioutil.ReadFile(pathJSON)
	checkError(err)
	err = json.Unmarshal([]byte(file), &result)
	checkError(err)
	return result, true
}

func (folderAnalyzer *FolderAnalyzer) checkFolder(path string) bool {

	marker, found := checkFolder(path)
	if !found || marker["type"] != "case" {
		return false
	}

	pathJSON := filepath.Join(path, ".sinf_mark.json")
	fmt.Printf("Encontrado marcação, arquivo: %s\n", pathJSON)
	if marker["name"] == folderAnalyzer.Name {
		switch marker["role"] {
		case "temp":
			folderAnalyzer.Sources = append(folderAnalyzer.Sources, path)
		case "final":
			folderAnalyzer.Destination = path
		}
	}
	return true
}

func (folderAnalyzer *FolderAnalyzer) findFolders(name string) {
	folderAnalyzer.Name = name
	drives := getDrives()
	for _, drive := range drives {
		marker, found := checkFolder(drive)
		if !found || marker["type"] != "disk" {
			continue
		}

		fmt.Printf("Vasculhando drive %s\n", drive)
		folderAnalyzer.findFoldersRecursively(drive, 0)
	}
}

func (folderAnalyzer *FolderAnalyzer) findFoldersRecursively(folder string, depth int) {
	if depth >= folderAnalyzer.MaxDeep {
		return
	}
	if folderAnalyzer.checkFolder(folder) {
		return
	}

	items, _ := ioutil.ReadDir(folder)
	for _, item := range items {

		if item.IsDir() && !strings.HasPrefix(item.Name(), "$") {
			absolute := filepath.Join(folder, item.Name())
			folderAnalyzer.findFoldersRecursively(absolute, depth+1)
		}
	}
}
