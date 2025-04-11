package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	. "reportdb/datastore/writer"
	. "reportdb/server"
	. "reportdb/storage"
	. "reportdb/utils"
	"syscall"
	"time"
)

func main() {

	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	err := InitConfig()

	if err != nil {

		log.Printf("Error initializing config: %v", err)

		return
	}

	dataChannel := make(chan []Events, 10)

	pollerContext, err := ZMQPoller(dataChannel)

	if err != nil {

		log.Printf("Error initializing ZMQ server: %v", err)

		GlobalShutdown = true

		pollerContext.Term()

		return
	}

	storePool := NewStorePool()

	queryContext, err := ZMQQuery(storePool)

	if err != nil {

		log.Printf("Error initializing ZMQ query: %v", err)

		GlobalShutdown = true

		pollerContext.Term()

		queryContext.Term()

		return
	}

	writers, err := StartWriter(storePool)

	if err != nil {

		log.Printf("Error starting writers: %v", err)

		GlobalShutdown = true

		pollerContext.Term()

		queryContext.Term()

		return
	}

	DistributeData(dataChannel, writers)

	ticker, err := storePool.SaveEngine()

	if err != nil {

		log.Printf("Error initializing store pool: %v", err)

		GlobalShutdown = true

		pollerContext.Term()

		queryContext.Term()

		ticker.Stop()

		return
	}

	<-signalChannel

	fmt.Println("\nstart shutting down : ", time.Now())

	GlobalShutdown = true

	pollerContext.Term()

	queryContext.Term()

	ticker.Stop()

	fmt.Println("\n shutdown", time.Now())

}
