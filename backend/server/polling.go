package server

import (
	. "backend/utils"
	"encoding/json"
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
)

type PollingServer struct {
	pullSocket *zmq4.Socket

	pushSocket *zmq4.Socket

	context *zmq4.Context

	shutdownPull chan bool

	shutdownPush chan bool
}

func NewPollingServer(deviceChannel chan []PollerDevice, dataChannel chan []byte) (*PollingServer, error) {

	context, err := zmq4.NewContext()

	if err != nil {

		return nil, fmt.Errorf("failed to create context: %v", err)
	}

	pullSocket, err := context.NewSocket(zmq4.PULL)

	if err != nil {

		context.Term()

		return nil, fmt.Errorf("failed to create PULL socket: %v", err)
	}

	if err := pullSocket.Connect("tcp://localhost:6001"); err != nil {

		pullSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to bind PULL socket: %v", err)
	}

	pushSocket, err := context.NewSocket(zmq4.PUSH)

	if err != nil {

		pullSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to create PUSH socket: %v", err)
	}

	if err := pushSocket.Connect("tcp://localhost:6002"); err != nil {

		pushSocket.Close()

		pullSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to bind PUSH socket: %v", err)
	}

	server := &PollingServer{

		pullSocket: pullSocket,

		pushSocket: pushSocket,

		context: context,

		shutdownPull: make(chan bool, 1),

		shutdownPush: make(chan bool, 1),
	}

	go server.pollingReceiver(dataChannel)

	go server.pollingSender(deviceChannel)

	return server, nil
}

func (server *PollingServer) pollingReceiver(dataChannel chan []byte) {

	for {

		select {

		case <-server.shutdownPull:

			server.pullSocket.Close()

			server.shutdownPull <- true

			return

		default:

			data, err := server.pullSocket.RecvBytes(0)

			if err != nil {

				log.Printf("pollingReceiver : Error receiving query: %v", err)

				continue
			}

			dataChannel <- data
		}
	}
}

func (server *PollingServer) pollingSender(deviceChannel chan []PollerDevice) {

	for {

		select {

		case <-server.shutdownPush:

			server.pushSocket.Close()

			close(deviceChannel)

			server.shutdownPush <- true

			return

		case pollerDevices := <-deviceChannel:

			jsonData, err := json.Marshal(pollerDevices)

			if err != nil {

				log.Printf("pollingSender : Error marshaling response: %v", err)

				continue
			}

			if _, err := server.pushSocket.SendBytes(jsonData, 0); err != nil {

				log.Printf("pollingSender : Error sending response: %v", err)

				continue
			}

		}
	}
}

func (server *PollingServer) Shutdown() {

	server.shutdownPull <- true

	server.shutdownPush <- true

	server.context.Term()

	<-server.shutdownPull

	<-server.shutdownPush
}
