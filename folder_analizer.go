package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// FolderMarker armazena informações sobre a pasta
type FolderMarker struct {
	Type string `json:"type"`
	Role string `json:"role"`
	Name string `json:"name"`
}

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

func (folderAnalyzer *FolderAnalyzer) checkFolder(path string) bool {
	pathJSON := filepath.Join(path, ".sinf_mark.json")
	_, err := os.Stat(pathJSON)
	if err != nil {
		return false
	}
	file, err := ioutil.ReadFile(pathJSON)
	checkError(err)
	data := FolderMarker{}

	err = json.Unmarshal([]byte(file), &data)
	checkError(err)
	if data.Name == folderAnalyzer.Name {
		switch data.Role {
		case "processamento":
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
		folderAnalyzer.Deep = 0
		folderAnalyzer.findFoldersRecursively(drive)
	}
}

func (folderAnalyzer *FolderAnalyzer) findFoldersRecursively(folder string) {
	if folderAnalyzer.checkFolder(folder) {
		return
	}
	folderAnalyzer.Deep++
	if folderAnalyzer.Deep >= folderAnalyzer.MaxDeep {
		return
	}
	items, _ := ioutil.ReadDir(folder)
	for _, item := range items {

		if item.IsDir() && !strings.HasPrefix(item.Name(), "$") {
			absolute := filepath.Join(folder, item.Name())
			folderAnalyzer.findFoldersRecursively(absolute)
		}
	}
}
