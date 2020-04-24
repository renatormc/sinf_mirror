package main

import (
	"io/ioutil"
	"path/filepath"
)

type Counter struct {
	config      *Config
	totalNumber int64
	totalSize   int64
}

func (counter *Counter) countFiles() {
	counter.totalNumber = 0
	counter.totalSize = 0
	counter.countFilesRecursively(counter.config.source)
}

func (counter *Counter) countFilesRecursively(path string) {

	items, _ := ioutil.ReadDir(path)
	for _, item := range items {
		if item.IsDir() {
			counter.countFilesRecursively(filepath.Join(path, item.Name()))
		} else {
			counter.totalNumber++
			counter.totalSize += item.Size()
		}
	}
}
