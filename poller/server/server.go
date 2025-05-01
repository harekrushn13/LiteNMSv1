package server

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	. "poller/logger"
	. "poller/utils"
)

type PollingServer struct {
	pullSocket *zmq4.Socket

	pushSocket *zmq4.Socket

	context *zmq4.Context

	shutdownPull chan bool

	shutdownPush chan bool
}

func NewPollingServer(deviceChannel chan []Device, dataChannel chan []Events) (*PollingServer, error) {

	context, err := zmq4.NewContext()

	if err != nil {

		return nil, fmt.Errorf("failed to create context: %v", err)
	}

	pullSocket, err := context.NewSocket(zmq4.PULL)

	if err != nil {

		context.Term()

		return nil, fmt.Errorf("failed to create PULL socket: %v", err)
	}

	if err := pullSocket.Bind("tcp://*:6002"); err != nil {

		pullSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to bind PULL socket: %v", err)
	}

	pushSocket, err := context.NewSocket(zmq4.PUSH)

	if err != nil {

		pullSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to create PUB socket: %v", err)
	}

	if err := pushSocket.Bind("tcp://*:6001"); err != nil {

		pushSocket.Close()

		pullSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to bind PUB socket: %v", err)
	}

	server := &PollingServer{

		pullSocket: pullSocket,

		pushSocket: pushSocket,

		context: context,

		shutdownPull: make(chan bool, 1),

		shutdownPush: make(chan bool, 1),
	}

	go server.pollingReceiver(deviceChannel)

	go server.pollingSender(dataChannel)

	return server, nil
}

func (server *PollingServer) pollingReceiver(deviceChannel chan []Device) {

	for {

		select {

		case <-server.shutdownPull:

			server.pullSocket.Close()

			close(deviceChannel)

			server.shutdownPull <- true

			return

		default:

			msg, err := server.pullSocket.RecvBytes(0)

			if err != nil {

				Logger.Warn("pollingReceiver:Error receiving query", zap.Error(err))

				continue
			}

			var devices []Device

			if err := msgpack.Unmarshal(msg, &devices); err != nil {

				Logger.Warn("pollingReceiver: Error unmarshalling query", zap.Error(err))

				continue
			}

			deviceChannel <- devices
		}
	}
}

func (server *PollingServer) pollingSender(dataChannel chan []Events) {

	for {

		select {

		case <-server.shutdownPush:

			server.pushSocket.Close()

			close(dataChannel)

			server.shutdownPush <- true

			return

		case events := <-dataChannel:

			fmt.Println("sent : ", len(events))

			data, err := msgpack.Marshal(events)

			if err != nil {

				Logger.Error("pollingSender: Error marshaling response", zap.Error(err))

				continue
			}

			if _, err = server.pushSocket.SendBytes(data, 0); err != nil {

				Logger.Error("pollingSender: Error sending response", zap.Error(err))

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
