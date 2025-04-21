package main

import (
	"fmt"
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

	deviceChan := make(chan []Device, 1)

	go StartProvisionListener(deviceChan, &waitGroup)

	go func() {

		for devices := range deviceChan {

			fmt.Println("setProvisionDevices", devices)

			SetProvisionedDevices(devices)

		}

	}()

	pollChan := PollData(&waitGroup)

	ZMQServer(pollChan, &waitGroup)

	waitGroup.Wait()
}
