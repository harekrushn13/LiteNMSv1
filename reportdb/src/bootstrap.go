package main

import (
	"log"
	. "reportdb/datastore/reader"
	. "reportdb/datastore/writer"
	. "reportdb/server"
	. "reportdb/storage"
	. "reportdb/utils"
	"sync"
)

func main() {

	//signalChannel := make(chan os.Signal, 1)
	//
	//signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	var waitGroup sync.WaitGroup

	err := InitConfig()

	if err != nil {

		log.Fatal(err)
	}

	dataChannel := make(chan []Events, 10)

	err = ZMQServer(dataChannel, &waitGroup)

	if err != nil {

		log.Fatal(err)
	}

	storePool := NewStorePool()

	writers, err := StartWriter(&waitGroup, storePool)

	if err != nil {

		log.Fatal(err)
	}

	DistributeData(dataChannel, writers, &waitGroup)

	StartReader(&waitGroup, storePool)

	err = storePool.SaveEngine(&waitGroup)

	if err != nil {

		log.Fatal(err)
	}

	waitGroup.Wait()
}
