package server

import (
	. "backend/logger"
	. "backend/utils"
	"encoding/json"
	"fmt"
	"github.com/pebbe/zmq4"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"sync"
)

type QueryServer struct {
	pushSocket *zmq4.Socket

	pullSocket *zmq4.Socket

	context *zmq4.Context

	shutdownPull chan bool

	shutdownPush chan bool

	queryMapping map[uint64]chan Response

	lock sync.RWMutex
}

func NewQueryServer(queryChannel chan QueryMap, queryMapping map[uint64]chan Response) (*QueryServer, error) {

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

	if err := pushSocket.Connect("tcp://localhost:6004"); err != nil {

		pushSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to connect PUSH socket: %v", err)
	}

	pullSocket, err := context.NewSocket(zmq4.PULL)

	if err != nil {

		pushSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to create PULL socket: %v", err)
	}

	if err := pullSocket.Connect("tcp://localhost:6005"); err != nil {

		pullSocket.Close()

		pushSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to connect PULL socket: %v", err)
	}

	server := &QueryServer{

		pushSocket: pushSocket,

		pullSocket: pullSocket,

		context: context,

		shutdownPull: make(chan bool, 1),

		shutdownPush: make(chan bool, 1),

		queryMapping: make(map[uint64]chan Response),

		lock: sync.RWMutex{},
	}

	go server.querySender(queryChannel)

	go server.responseReceiver()

	return server, nil
}

func (server *QueryServer) querySender(queryChannel chan QueryMap) {

	for {

		select {

		case <-server.shutdownPush:

			server.pushSocket.Close()

			close(queryChannel)

			server.shutdownPush <- true

			return

		case queryMap := <-queryChannel:

			querySend := QuerySend{

				RequestID: queryMap.RequestID,

				QueryRequest: queryMap.QueryRequest,
			}

			queryBytes, err := msgpack.Marshal(querySend)

			if err != nil {

				AsyncWarn("querySender: failed to marshal query", zap.Error(err))

				continue
			}

			if _, err := server.pushSocket.SendBytes(queryBytes, 0); err != nil {

				AsyncWarn("querySender: failed to send query", zap.Error(err))

				continue
			}

			server.lock.Lock()

			server.queryMapping[querySend.RequestID] = queryMap.Response

			server.lock.Unlock()
		}
	}
}

func (server *QueryServer) responseReceiver() {

	for {

		select {

		case <-server.shutdownPull:

			server.pullSocket.Close()

			server.lock.Lock()

			for id, ch := range server.queryMapping {

				ch <- Response{RequestID: id, Error: "Server shutdown"}

				close(ch)
			}

			server.lock.Unlock()

			server.shutdownPull <- true

			return

		default:

			data, err := server.pullSocket.RecvBytes(0)

			if err != nil {

				AsyncWarn("responseReceiver: Error receiving response", zap.Error(err))

				continue
			}

			var response Response

			if err := json.Unmarshal(data, &response); err != nil {

				AsyncWarn("responseReceiver: Error unmarshalling response", zap.Error(err))

				continue
			}

			server.lock.Lock()

			if ch, ok := server.queryMapping[response.RequestID]; ok {

				ch <- response

				delete(server.queryMapping, response.RequestID)

				close(ch)
			}

			server.lock.Unlock()
		}
	}
}

func (server *QueryServer) Shutdown() {

	server.shutdownPull <- true

	server.shutdownPush <- true

	server.context.Term()

	<-server.shutdownPull

	<-server.shutdownPush

}
