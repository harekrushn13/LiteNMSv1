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

	ZMQServer(PollData(&waitGroup), &waitGroup)

	waitGroup.Wait()
}
