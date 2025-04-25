package server

import (
	. "backend/utils"
	"encoding/json"
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
)

type QueryServer struct {
	pushSocket *zmq4.Socket

	pullSocket *zmq4.Socket

	context *zmq4.Context

	shutdownPull chan bool

	shutdownPush chan bool
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
	}

	go server.querySender(queryChannel, queryMapping)

	go server.responseReceiver(queryMapping)

	return server, nil
}

func (server *QueryServer) querySender(queryChannel chan QueryMap, queryMapping map[uint64]chan Response) {

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

			queryBytes, err := json.Marshal(querySend)

			if err != nil {

				log.Printf("querySender: failed to marshal query: %v", err)

				continue
			}

			if _, err := server.pushSocket.SendBytes(queryBytes, 0); err != nil {

				log.Printf("querySender: failed to send query: %v", err)

				continue
			}

			queryMapping[querySend.RequestID] = queryMap.Response
		}
	}
}

func (server *QueryServer) responseReceiver(queryMapping map[uint64]chan Response) {

	for {

		select {

		case <-server.shutdownPull:

			server.pullSocket.Close()

			server.shutdownPull <- true

			return

		default:

			data, err := server.pullSocket.RecvBytes(0)

			if err != nil {

				log.Printf("responseReceiver: Error receiving response: %v", err)

				continue
			}

			var response Response

			if err := json.Unmarshal(data, &response); err != nil {

				log.Printf("responseReceiver: Error unmarshaling response: %v", err)

				continue
			}

			if ch, ok := queryMapping[response.RequestID]; ok {

				ch <- response

				delete(queryMapping, response.RequestID)
			}
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
