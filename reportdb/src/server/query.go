package server

import (
	"encoding/json"
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	. "reportdb/datastore/reader"
	. "reportdb/storage"
	. "reportdb/utils"
	"sync"
	"time"
)

type QueryServer struct {
	pullSocket *zmq4.Socket

	pubSocket *zmq4.Socket

	context *zmq4.Context

	storePool *StorePool

	workerPool chan struct{}

	shutdown chan bool

	waitGroup *sync.WaitGroup
}

type Response struct {
	RequestID string `json:"request_id"`

	Data interface{} `json:"data"`

	Error string `json:"error,omitempty"`

	Timestamp time.Time `json:"timestamp"`
}

func NewQueryServer(workerCount int, storePool *StorePool) (*QueryServer, error) {

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

		storePool: storePool,

		workerPool: make(chan struct{}, workerCount),

		shutdown: make(chan bool, 1),

		waitGroup: &sync.WaitGroup{},
	}

	for i := 0; i < workerCount; i++ {

		server.workerPool <- struct{}{}
	}

	go server.queryHandler()

	return server, nil
}

func (queryServer *QueryServer) queryHandler() {

	for {

		select {

		case <-queryServer.shutdown:

			queryServer.pubSocket.Close()

			queryServer.pullSocket.Close()

			//queryServer.waitGroup.Wait()

			queryServer.shutdown <- true

			return

		default:

			msg, err := queryServer.pullSocket.RecvBytes(0)

			if err != nil {

				log.Printf("QueryServer : Error receiving query: %v", err)

				continue
			}

			<-queryServer.workerPool

			//queryServer.waitGroup.Add(1)

			go func(msg []byte) {

				defer func() {

					queryServer.workerPool <- struct{}{}

					//queryServer.waitGroup.Done()

				}()

				var query Query

				if err := json.Unmarshal(msg, &query); err != nil {

					log.Printf("QueryServer : Error unmarshaling query: %v", err)

					return
				}

				response := queryServer.processQuery(query)

				responseBytes, err := json.Marshal(response)

				if err != nil {

					log.Printf("QueryServer : Error marshaling response: %v", err)

					return
				}

				if _, err := queryServer.pubSocket.SendBytes(responseBytes, 0); err != nil {

					log.Printf("QueryServer : Error sending response: %v", err)

					return
				}

			}(msg)
		}
	}
}

func (queryServer *QueryServer) processQuery(query Query) Response {

	response, err := FetchData(query, queryServer.storePool)

	if err != nil {

		return Response{

			RequestID: query.RequestID,

			Error: err.Error(),

			Timestamp: time.Now(),
		}

	}

	return Response{

		RequestID: query.RequestID,

		Data: response,

		Timestamp: time.Now(),
	}
}

func (queryServer *QueryServer) Shutdown() {

	queryServer.shutdown <- true

	queryServer.context.Term()

	<-queryServer.shutdown
}
