package main

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/signal"
	. "poller/logger"
	. "poller/polling"
	. "poller/server"
	. "poller/utils"
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

	err := InitConfig()

	if err != nil {

		Logger.Error("InitConfig error", zap.Error(err))

		return
	}

	deviceChannel := make(chan []Device, GetDeviceBuffer())

	dataChannel := make(chan []Events, GetDataBuffer())

	pollingServer, err := NewPollingServer(deviceChannel, dataChannel)

	if err != nil {

		Logger.Error("NewPollingServer error", zap.Error(err))

		return
	}

	poller := NewPoller()

	poller.SetProvisionedDevices(deviceChannel)

	poller.StartPolling(dataChannel)

	<-signalChannel

	Logger.Info("Start shutting down", zap.Time("time", time.Now()))

	pollingServer.Shutdown()

	poller.ShutdownPoller()

	Logger.Info("Shutdown complete", zap.Time("time", time.Now()))
}
