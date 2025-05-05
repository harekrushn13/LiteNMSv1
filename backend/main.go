package main

import (
	. "backend/logger"
	. "backend/models"
	. "backend/routes"
	. "backend/server"
	. "backend/utils"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	if err := InitLogger(); err != nil {

		fmt.Printf("Failed to initialize logger: %v\n", err)

		return
	}

	defer Logger.Sync()

	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	if err := InitDB(); err != nil {

		Logger.Error("Failed to initialize database", zap.Error(err))

		return
	}

	defer DB.Close()

	if err := InitConfig(); err != nil {

		Logger.Error("Failed to initialize config", zap.Error(err))

		return
	}

	deviceChannel := make(chan []PollerDevice, GetDeviceBuffer())

	dataChannel := make(chan []byte, GetDataBuffer())

	queryChannel := make(chan QueryMap, GetQueryBuffer())

	queryMapping := make(map[uint64]chan Response)

	router := InitRoutes(DB, deviceChannel, queryChannel)

	pollingServer, err := NewPollingServer(deviceChannel, dataChannel)

	if err != nil {

		Logger.Error("Failed to initialize polling server", zap.Error(err))

		pollingServer.Shutdown()

		return
	}

	dbServer, err := NewDBServer(dataChannel)

	if err != nil {

		Logger.Error("Failed to initialize database server", zap.Error(err))

		pollingServer.Shutdown()

		dbServer.Shutdown()

		return
	}

	queryServer, err := NewQueryServer(queryChannel, queryMapping)

	if err != nil {

		Logger.Error("Failed to initialize query client", zap.Error(err))

		queryServer.Shutdown()

		dbServer.Shutdown()

		pollingServer.Shutdown()

		return
	}

	server := &http.Server{
		Addr: ":8080",

		Handler: router,
	}

	go func() {

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {

			Logger.Error("Failed to start server", zap.Error(err))

		}

	}()

	<-signalChannel

	Logger.Info("Start shutting down", zap.Time("time", time.Now()))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {

		log.Printf("Server shutdown failed: %v", err)
	}

	pollingServer.Shutdown()

	dbServer.Shutdown()

	queryServer.Shutdown()

	Logger.Info("Shutdown complete", zap.Time("time", time.Now()))
}
