package main

import (
	"log"
	. "reportdb/datastore/reader"
	. "reportdb/datastore/writer"
	. "reportdb/polling"
	. "reportdb/storage"
	. "reportdb/utils"
	"sync"
)

func main() {

	var waitGroup sync.WaitGroup

	err := InitConfig()

	if err != nil {

		log.Fatal(err)
	}

	storePool := NewStorePool()

	pollChannel := PollData(&waitGroup)

	StartWriter(pollChannel, &waitGroup, storePool)

	StartReader(&waitGroup, storePool)

	storePool.SaveEngine(&waitGroup)

	waitGroup.Wait()
}
