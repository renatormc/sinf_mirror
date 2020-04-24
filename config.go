package main

import (
	"gopkg.in/ini.v1"
)

type Config struct {
	NWorkers int
}

func GetConfig() Config {

	var config Config
	cfg, err := ini.Load("config.ini")
	checkError(err)
	config.NWorkers, err = cfg.Section("").Key("n_workers").Int()
	checkError(err)
	return config
}
