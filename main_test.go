package main

import (
	"fmt"
	"strconv"
	"testing"
)

func TestAbs(t *testing.T) {
	log := Logger{}
	log.init()
	var info LogInfo
	info.logAction = "!@#finish"
	info.logPath = "Location test"

	go log.run()
	fmt.Println("Logger routine is running")

	for value := 0; value <= 6; value++ {
		var infoTest LogInfo
		infoTest.logAction = "novo arquivo"
		infoTest.logPath = "caminho" + " " + strconv.Itoa(value)

		log.receiver <- infoTest
	}

	log.receiver <- info //last message with secret ending message

	fmt.Println("Log output folder " + log.logOutput)
	log.outputOK()

	<-log.finishedLogging
}
