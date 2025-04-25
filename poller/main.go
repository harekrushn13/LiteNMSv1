package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	. "poller/polling"
	. "poller/server"
	. "poller/utils"
	"syscall"
	"time"
)

func main() {

	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	err := InitConfig()

	if err != nil {

		log.Printf("InitConfig error: %v", err)

		return
	}

	deviceChannel := make(chan []Device, GetDeviceBuffer())

	dataChannel := make(chan []Events, GetDataBuffer())

	pollingServer, err := NewPollingServer(deviceChannel, dataChannel)

	if err != nil {

		log.Printf("NewPollingServer error: %v", err)

		return
	}

	poller := NewPoller()

	poller.SetProvisionedDevices(deviceChannel)

	poller.StartPolling(dataChannel)

	<-signalChannel

	fmt.Println("\nstart shutting down : ", time.Now())

	pollingServer.Shutdown()

	fmt.Println("\nshutdown", time.Now())
}
