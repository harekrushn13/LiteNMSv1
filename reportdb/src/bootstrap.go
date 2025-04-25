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
	"runtime"
	"syscall"
	"time"
)

func main() {

	go func() {

		http.ListenAndServe("localhost:6060", nil)
	}()

	var stat runtime.MemStats

	statTicker := time.NewTicker(time.Minute)

	go func() {

		for {

			select {

			case <-statTicker.C:

				runtime.ReadMemStats(&stat)

				log.Printf("NumGC: %v  GCCPUFraction : %v", stat.NumGC, stat.GCCPUFraction)
			}
		}
	}()

	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	err := InitConfig()

	if err != nil {

		log.Printf("Error initializing config: %v", err)

		return
	}

	dataChannel := make(chan []Events, GetDataBuffer())

	pollingServer, err := NewPollingServer(dataChannel)

	if err != nil {

		log.Printf("Error initializing ZMQ queryServer: %v", err)

		return
	}

	storePool := NewStorePool()

	writers, err := StartWriter(storePool)

	if err != nil {

		log.Printf("Error starting writers: %v", err)

		return
	}

	DistributeData(dataChannel, writers)

	responseChannel := make(chan Response, GetResponseBuffer())

	readers, err := StartReaders(storePool, responseChannel)

	if err != nil {

		log.Printf("Error starting readers: %v", err)

		return
	}

	queryChannel := make(chan QueryReceive, GetQueryBuffer())

	queryServer, err := NewQueryServer(queryChannel, responseChannel)

	if err != nil {

		log.Printf("Failed to start queryServer: %v", err)

		return
	}

	DistributeQuery(queryChannel, readers)

	err = storePool.SaveEngine()

	if err != nil {

		log.Printf("Error initializing store pool: %v", err)

		return
	}

	<-signalChannel

	fmt.Println("\nstart shutting down : ", time.Now())

	pollingServer.Shutdown()

	queryServer.Shutdown()

	storePool.Shutdown()

	fmt.Println("\nshutdown", time.Now())
}
