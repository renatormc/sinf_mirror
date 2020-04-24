package main

import (
	"gopkg.in/ini.v1"
)

type AppConfig struct {
	NWorkersCount   int
	NWorkersProcess int
	NRepeatCopy     int
}

func GetConfig() AppConfig {

	var config AppConfig
	cfg, err := ini.Load("config.ini")
	checkError(err)
	config.NWorkersCount, err = cfg.Section("").Key("n_workers_count").Int()
	checkError(err)
	config.NWorkersProcess, err = cfg.Section("").Key("n_workers_process").Int()
	checkError(err)
	config.NRepeatCopy, err = cfg.Section("").Key("n_workers_process").Int()
	checkError(err)
	return config
}
