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

	dataChannel := make(chan []Events, 10) // 10 buffer to receive []Events

	pollerServer, err := ZMQPoller(dataChannel)

	if err != nil {

		log.Printf("Error initializing ZMQ queryServer: %v", err)

		return
	}

	storePool := NewStorePool()

	queryServer, err := NewQueryServer(10, storePool) // 10 workers

	if err != nil {

		log.Printf("Failed to start queryServer: %v", err)

		return
	}

	writers, err := StartWriter(storePool)

	if err != nil {

		log.Printf("Error starting writers: %v", err)

		return
	}

	DistributeData(dataChannel, writers)

	err = storePool.SaveEngine()

	if err != nil {

		log.Printf("Error initializing store pool: %v", err)

		return
	}

	<-signalChannel

	fmt.Println("\nstart shutting down : ", time.Now())

	GlobalShutdown = true

	pollerServer.Shutdown()

	queryServer.Shutdown()

	storePool.Shutdown()

	fmt.Println("\nshutdown", time.Now())
}
