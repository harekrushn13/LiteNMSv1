package main

import (
	. "backend/models"
	. "backend/routes"
	. "backend/server"
	. "backend/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	if err := InitDB(); err != nil {

		log.Printf("Failed to initialize database: %v", err)

		return
	}

	defer DB.Close()

	deviceChannel := make(chan []PollerDevice, 5)

	dataChannel := make(chan []byte, 5)

	queryChannel := make(chan QueryMap, 5)

	queryMapping := make(map[uint64]chan Response)

	router := InitRoutes(DB, deviceChannel, queryChannel)

	pollingServer, err := NewPollingServer(deviceChannel, dataChannel)

	if err != nil {

		log.Printf("Failed to initialize polling server: %v", err)

		pollingServer.Shutdown()

		return
	}

	dbServer, err := NewDBServer(dataChannel)

	if err != nil {

		log.Printf("Failed to initialize database server: %v", err)

		pollingServer.Shutdown()

		dbServer.Shutdown()

		return
	}

	queryServer, err := NewQueryServer(queryChannel, queryMapping)

	if err != nil {

		log.Printf("Failed to initialize query client: %v", err)

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

			log.Printf("Failed to start server: %v", err)

		}
	}()

	<-signalChannel

	fmt.Println("\nstart shutting down : ", time.Now())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {

		log.Printf("Server shutdown failed: %v", err)
	}

	pollingServer.Shutdown()

	fmt.Println("1")

	dbServer.Shutdown()

	fmt.Println("2")

	queryServer.Shutdown()

	fmt.Println("\nshutdown", time.Now())
}
