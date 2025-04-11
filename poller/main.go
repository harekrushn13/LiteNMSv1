package main

import (
	. "poller/polling"
	. "poller/server"
	. "poller/utils"
	"sync"
)

func main() {

	InitConfig()

	var waitGroup sync.WaitGroup

	ZMQServer(PollData(&waitGroup), &waitGroup)

	waitGroup.Wait()
}
