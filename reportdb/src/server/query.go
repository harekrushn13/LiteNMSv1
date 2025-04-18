package server

import (
	"encoding/json"
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	. "reportdb/utils"
)

type QueryServer struct {
	pullSocket *zmq4.Socket

	pubSocket *zmq4.Socket

	context *zmq4.Context

	shutdownPull chan bool

	shutdownPub chan bool
}

func NewQueryServer(queryChannel chan Query, resultChannel chan Response) (*QueryServer, error) {

	context, err := zmq4.NewContext()

	if err != nil {

		return nil, fmt.Errorf("failed to create context: %v", err)
	}

	pullSocket, err := context.NewSocket(zmq4.PULL)

	if err != nil {

		context.Term()

		return nil, fmt.Errorf("failed to create PULL socket: %v", err)
	}

	if err := pullSocket.Bind("tcp://*:6001"); err != nil {

		pullSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to bind PULL socket: %v", err)
	}

	pubSocket, err := context.NewSocket(zmq4.PUB)

	if err != nil {

		pullSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to create PUB socket: %v", err)
	}

	if err := pubSocket.Bind("tcp://*:6002"); err != nil {

		pubSocket.Close()

		pullSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to bind PUB socket: %v", err)
	}

	server := &QueryServer{

		pullSocket: pullSocket,

		pubSocket: pubSocket,

		context: context,

		shutdownPull: make(chan bool, 1),

		shutdownPub: make(chan bool, 1),
	}

	go server.queryHandler(queryChannel)

	go server.responseHandler(resultChannel)

	return server, nil
}

func (queryServer *QueryServer) queryHandler(queryChannel chan Query) {

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

				log.Printf("QueryServer : Error receiving query: %v", err)

				continue
			}

			var query Query

			if err := json.Unmarshal(msg, &query); err != nil {

				log.Printf("QueryServer : Error unmarshaling query: %v", err)

				return
			}

			queryChannel <- query
		}
	}
}

func (queryServer *QueryServer) responseHandler(resultChannel chan Response) {

	for {

		select {

		case <-queryServer.shutdownPub:

			queryServer.pubSocket.Close()

			close(resultChannel)

			queryServer.shutdownPub <- true

			return

		case response := <-resultChannel:

			responseBytes, err := json.Marshal(response)

			if err != nil {

				log.Printf("QueryServer : Error marshaling response: %v", err)

				return
			}

			if _, err := queryServer.pubSocket.SendBytes(responseBytes, 0); err != nil {

				log.Printf("QueryServer : Error sending response: %v", err)

				return
			}

			if response.Error == "" {

				fmt.Println("response :", len(response.Data.([]interface{})))
			}
		}
	}
}

func (queryServer *QueryServer) Shutdown() {

	queryServer.shutdownPull <- true

	queryServer.shutdownPub <- true

	queryServer.context.Term()

	<-queryServer.shutdownPull

	<-queryServer.shutdownPub
}
