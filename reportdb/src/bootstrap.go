package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	. "reportdb/datastore/reader"
	. "reportdb/datastore/writer"
	. "reportdb/server"
	. "reportdb/storage"
	. "reportdb/utils"
	"syscall"
	"time"
)

func main() {

	// For profiling

	go func() {

		http.ListenAndServe("localhost:6060", nil)
	}()

	// Handle Interrupts

	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Initialise Config

	err := InitConfig()

	if err != nil {

		log.Printf("Error initializing config: %v", err)

		return
	}

	// Initialise dataChannel with eventsBuffer

	eventsBuffer, err := GetEventsBuffer()

	if err != nil {

		log.Printf("Error initializing events buffer: %v", err)

		return
	}

	dataChannel := make(chan []Events, eventsBuffer) // buffer to receive []Events

	// pollerServer to receive data from poller

	pollerServer, err := NewPollerServer(dataChannel)

	if err != nil {

		log.Printf("Error initializing ZMQ queryServer: %v", err)

		return
	}

	// Initialise storePool for storeEngine

	storePool := NewStorePool()

	// Initialise multiple writer to handle event

	writers, err := StartWriter(storePool)

	if err != nil {

		log.Printf("Error starting writers: %v", err)

		return
	}

	// Distribute batch data among multiple writers

	DistributeData(dataChannel, writers)

	// query responseChannel

	responseChannel := make(chan Response, 3)

	// Initialise multiple readers

	readers, err := StartReaders(storePool, responseChannel)

	if err != nil {

		log.Printf("Error starting readers: %v", err)

		return
	}

	// Initialise queryChannel with buffer

	queryChannel := make(chan Query, 10)

	// Initialise queryServer to receive query from clients

	queryServer, err := NewQueryServer(queryChannel, responseChannel)

	if err != nil {

		log.Printf("Failed to start queryServer: %v", err)

		return
	}

	// Distribute query among readers

	DistributeQuery(queryChannel, readers)

	// save index file for those day who written new data

	err = storePool.SaveEngine()

	if err != nil {

		log.Printf("Error initializing store pool: %v", err)

		return
	}

	// close all resources

	<-signalChannel

	fmt.Println("\nstart shutting down : ", time.Now())

	pollerServer.Shutdown()

	queryServer.Shutdown()

	storePool.Shutdown()

	fmt.Println("\nshutdown", time.Now())
}
