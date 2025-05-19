package server

import (
	"encoding/json"
	"fmt"
	"github.com/pebbe/zmq4"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	. "reportdb/logger"
	. "reportdb/utils"
)

type QueryServer struct {
	pullSocket *zmq4.Socket

	pushSocket *zmq4.Socket

	context *zmq4.Context

	shutdownPull chan bool

	shutdownPush chan bool
}

func NewQueryServer(queryChannel chan QueryReceive, resultChannel chan Response) (*QueryServer, error) {

	context, err := zmq4.NewContext()

	if err != nil {

		return nil, fmt.Errorf("failed to create context: %v", err)
	}

	pullSocket, err := context.NewSocket(zmq4.PULL)

	if err != nil {

		context.Term()

		return nil, fmt.Errorf("failed to create PULL socket: %v", err)
	}

	if err := pullSocket.Bind("tcp://*:6004"); err != nil {

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

	pushSocket.SetLinger(0)

	if err := pushSocket.Bind("tcp://*:6005"); err != nil {

		pushSocket.Close()

		pullSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to bind PUSH socket: %v", err)
	}

	server := &QueryServer{

		pullSocket: pullSocket,

		pushSocket: pushSocket,

		context: context,

		shutdownPull: make(chan bool, 1),

		shutdownPush: make(chan bool, 1),
	}

	go server.queryReceiver(queryChannel)

	go server.responseSender(resultChannel)

	return server, nil
}

func (queryServer *QueryServer) queryReceiver(queryChannel chan QueryReceive) {

	for {

		select {

		case <-queryServer.shutdownPull:

			queryServer.pullSocket.Close()

			close(queryChannel)

			queryServer.shutdownPull <- true

			return

		default:

			msg, err := queryServer.pullSocket.RecvBytes(0)

			if err != nil {

				Logger.Warn("queryReceiver : Error receiving query", zap.Error(err))

				continue
			}

			var query QueryReceive

			if err := msgpack.Unmarshal(msg, &query); err != nil {

				Logger.Warn("queryReceiver : Error unmarshalling query", zap.Error(err))

				return
			}

			queryChannel <- query
		}
	}
}

func (queryServer *QueryServer) responseSender(resultChannel chan Response) {

	for {

		select {

		case <-queryServer.shutdownPush:

			queryServer.pushSocket.Close()

			close(resultChannel)

			queryServer.shutdownPush <- true

			return

		case response := <-resultChannel:

			responseBytes, err := json.Marshal(response)

			if err != nil {

				Logger.Warn("responseSender : Error marshaling response", zap.Error(err))

				return
			}

			if _, err := queryServer.pushSocket.SendBytes(responseBytes, 0); err != nil {

				Logger.Warn("responseSender : Error sending response", zap.Error(err))

				return
			}

		}
	}
}

func (queryServer *QueryServer) Shutdown() {

	queryServer.shutdownPull <- true

	queryServer.shutdownPush <- true

	queryServer.context.Term()

	<-queryServer.shutdownPull

	<-queryServer.shutdownPush
}
