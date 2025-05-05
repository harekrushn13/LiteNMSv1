package main

import (
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	. "reportdb/cache"
	. "reportdb/datastore/reader"
	. "reportdb/datastore/writer"
	. "reportdb/logger"
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

				hit, missed, hitratio := GetMetrics()

				log.Printf("hit: %v, missed: %v, hitratio: %v", hit, missed, hitratio)
			}
		}
	}()

	if err := InitLogger(); err != nil {

		fmt.Printf("Failed to initialize logger: %v\n", err)

		return
	}

	defer Logger.Sync()

	if err := InitCache(); err != nil {

		Logger.Error("Failed to initialize cache", zap.Error(err))

		return
	}

	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	err := InitConfig()

	if err != nil {

		Logger.Error("Error initializing config", zap.Error(err))

		return
	}

	dataChannel := make(chan []Events, GetDataBuffer())

	pollingServer, err := NewPollingServer(dataChannel)

	if err != nil {

		Logger.Error("Error initializing ZMQ queryServer", zap.Error(err))

		return
	}

	storePool := NewStorePool()

	writers, err := StartWriter(storePool)

	if err != nil {

		Logger.Error("Error starting writers", zap.Error(err))

		return
	}

	DistributeData(dataChannel, writers)

	responseChannel := make(chan Response, GetResponseBuffer())

	readers, err := StartReaders(storePool, responseChannel)

	if err != nil {

		Logger.Error("Error starting readers", zap.Error(err))

		return
	}

	queryChannel := make(chan QueryReceive, GetQueryBuffer())

	queryServer, err := NewQueryServer(queryChannel, responseChannel)

	if err != nil {

		Logger.Error("Failed to start queryServer", zap.Error(err))

		return
	}

	DistributeQuery(queryChannel, readers)

	err = storePool.SaveEngine()

	if err != nil {

		Logger.Error("Error initializing store pool", zap.Error(err))

		return
	}

	<-signalChannel

	Logger.Info("Start shutting down", zap.Time("time", time.Now()))

	pollingServer.Shutdown()

	queryServer.Shutdown()

	storePool.Shutdown()

	Logger.Info("Shutdown complete", zap.Time("time", time.Now()))

}
