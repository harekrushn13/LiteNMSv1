package server

import (
	. "backend/logger"
	"fmt"
	"github.com/pebbe/zmq4"
	"go.uber.org/zap"
)

type DBServer struct {
	pushSocket *zmq4.Socket

	context *zmq4.Context

	shutdownPush chan bool
}

func NewDBServer(dataChannel chan []byte) (*DBServer, error) {

	context, err := zmq4.NewContext()

	if err != nil {

		return nil, fmt.Errorf("failed to create context: %v", err)
	}

	pushSocket, err := context.NewSocket(zmq4.PUSH)

	if err != nil {

		context.Term()

		return nil, fmt.Errorf("failed to create PUSH socket: %v", err)
	}

	pushSocket.SetLinger(0)

	if err := pushSocket.Connect("tcp://localhost:6003"); err != nil {

		pushSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to bind PUSH socket: %v", err)
	}

	server := &DBServer{

		pushSocket: pushSocket,

		context: context,

		shutdownPush: make(chan bool, 1),
	}

	go server.dbSender(dataChannel)

	return server, nil
}

func (server *DBServer) dbSender(dataChannel chan []byte) {

	for {

		select {

		case <-server.shutdownPush:

			server.pushSocket.Close()

			close(dataChannel)

			server.shutdownPush <- true

			return

		case data := <-dataChannel:

			if _, err := server.pushSocket.SendBytes(data, 0); err != nil {

				AsyncWarn("dbSender : Error sending data", zap.Error(err))

				continue
			}
		}
	}

}

func (server *DBServer) Shutdown() {

	server.shutdownPush <- true

	server.context.Term()

	<-server.shutdownPush
}
