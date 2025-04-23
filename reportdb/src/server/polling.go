package server

import (
	"encoding/json"
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	. "reportdb/utils"
)

type PollingServer struct {
	pullSocket *zmq4.Socket

	context *zmq4.Context

	shutdownPull chan bool
}

func NewPollingServer(dataChannel chan []Events) (*PollingServer, error) {

	context, err := zmq4.NewContext()

	if err != nil {

		return nil, fmt.Errorf("failed to create context: %v", err)
	}

	pullSocket, err := context.NewSocket(zmq4.PULL)

	if err != nil {

		context.Term()

		return nil, fmt.Errorf("failed to create PULL socket: %v", err)
	}

	if err := pullSocket.Bind("tcp://*:6003"); err != nil {

		pullSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to bind PULL socket: %v", err)
	}

	server := &PollingServer{

		pullSocket: pullSocket,

		context: context,

		shutdownPull: make(chan bool, 1),
	}

	go server.pollingReceiver(dataChannel)

	return server, nil
}

func (server *PollingServer) pollingReceiver(dataChannel chan []Events) {

	for {

		select {

		case <-server.shutdownPull:

			server.pullSocket.Close()

			server.shutdownPull <- true

			return

		default:

			batchData, err := server.pullSocket.RecvBytes(0)

			if err != nil {

				log.Printf("pollingReceiver : Error receiving query: %v", err)

				continue
			}

			var events []Events

			err = json.Unmarshal(batchData, &events)

			if err != nil {

				log.Printf("pollingReceiver : Error unmarshalling message: %v", err)

				continue
			}

			dataChannel <- events
		}
	}
}

func (server *PollingServer) Shutdown() {

	server.shutdownPull <- true

	server.context.Term()

	<-server.shutdownPull
}
