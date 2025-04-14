package engine

import (
	"encoding/json"
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	. "queryengine/common"
	"sync"
)

type QueryEngine struct {
	pushSocket *zmq4.Socket

	subSocket *zmq4.Socket

	context *zmq4.Context

	pending sync.Map

	shutdown chan bool
}

func NewQueryEngine() (*QueryEngine, error) {

	context, err := zmq4.NewContext()

	if err != nil {

		return nil, fmt.Errorf("failed to create context: %v", err)
	}

	pushSocket, err := context.NewSocket(zmq4.PUSH)

	if err != nil {

		context.Term()

		return nil, fmt.Errorf("failed to create PUSH socket: %v", err)
	}

	if err := pushSocket.Connect(QueryAddress); err != nil {

		pushSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to connect PUSH socket: %v", err)
	}

	subSocket, err := context.NewSocket(zmq4.SUB)

	if err != nil {

		pushSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to create SUB socket: %v", err)
	}

	if err := subSocket.Connect(ResponseAddress); err != nil {

		subSocket.Close()

		pushSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to connect SUB socket: %v", err)
	}

	if err := subSocket.SetSubscribe(""); err != nil {

		subSocket.Close()

		pushSocket.Close()

		context.Term()

		return nil, fmt.Errorf("failed to subscribe: %v", err)
	}

	engine := &QueryEngine{

		pushSocket: pushSocket,

		subSocket: subSocket,

		context: context,

		shutdown: make(chan bool, 1),
	}

	go engine.responseHandler()

	return engine, nil
}

func (qe *QueryEngine) SubmitQuery(query Query) (chan Response, error) {

	responseChan := make(chan Response, 1)

	qe.pending.Store(query.RequestID, responseChan)

	queryBytes, err := json.Marshal(query)

	if err != nil {

		qe.pending.Delete(query.RequestID)

		return nil, fmt.Errorf("failed to marshal query: %v", err)
	}

	if _, err := qe.pushSocket.SendBytes(queryBytes, 0); err != nil {

		qe.pending.Delete(query.RequestID)

		return nil, fmt.Errorf("failed to send query: %v", err)
	}

	return responseChan, nil
}

func (qe *QueryEngine) responseHandler() {

	for {

		select {

		case <-qe.shutdown:

			qe.subSocket.Close()

			qe.pushSocket.Close()

			qe.shutdown <- true

			return

		default:

			msg, err := qe.subSocket.RecvBytes(0)

			if err != nil {

				log.Printf("Error receiving response: %v", err)

				continue
			}

			var response Response

			if err := json.Unmarshal(msg, &response); err != nil {

				log.Printf("Error unmarshaling response: %v", err)

				continue
			}

			if ch, ok := qe.pending.Load(response.RequestID); ok {

				ch.(chan Response) <- response

				qe.pending.Delete(response.RequestID)
			}
		}
	}
}

func (qe *QueryEngine) Shutdown() {

	qe.shutdown <- true

	qe.context.Term()

	<-qe.shutdown
}
