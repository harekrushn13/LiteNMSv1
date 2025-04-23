package main

import (
	"log"
	. "poller/polling"
	. "poller/server"
	. "poller/utils"
	"sync"
)

func main() {

	err := InitConfig()

	if err != nil {

		log.Fatal("InitConfig error: ", err)

		return
	}

	var waitGroup sync.WaitGroup

	deviceChannel := make(chan []Device, 1)

	dataChannel := make(chan []Events, 20)

	_, err = NewPollingServer(deviceChannel, dataChannel)

	if err != nil {

		log.Fatal("NewPollingServer error: ", err)
	}

	go func() {

		for device := range deviceChannel {

			SetProvisionedDevices(device)
		}

	}()

	PollData(dataChannel, &waitGroup)

	waitGroup.Wait()
}
