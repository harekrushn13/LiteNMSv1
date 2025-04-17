package server

import (
	"encoding/json"
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	. "reportdb/utils"
)

type PollerServer struct {
	context *zmq4.Context

	subscriber *zmq4.Socket

	shutdown chan bool
}

func NewPollerServer(dataChannel chan []Events) (*PollerServer, error) {

	context, err := zmq4.NewContext()

	if err != nil {

		return nil, fmt.Errorf("NewPollerServer : Error creating ZMQ context: %v", err)
	}

	subscriber, err := context.NewSocket(zmq4.SUB)

	if err != nil {

		context.Term()

		return nil, fmt.Errorf("NewPollerServer : Error creating subscriber: %v", err)
	}

	err = subscriber.Connect("tcp://localhost:6000")

	if err != nil {

		subscriber.Close()

		context.Term()

		return nil, fmt.Errorf("NewPollerServer : Error connecting to subscriber: %v", err)
	}

	err = subscriber.SetSubscribe("")

	if err != nil {

		subscriber.Close()

		context.Term()

		return nil, fmt.Errorf("NewPollerServer : Error subscribing to zmq4: %v", err)
	}

	poller := &PollerServer{

		context: context,

		subscriber: subscriber,

		shutdown: make(chan bool, 1),
	}

	go poller.listen(dataChannel)

	return poller, nil
}

func (poller *PollerServer) listen(dataChannel chan []Events) {

	for {

		select {

		case <-poller.shutdown:

			poller.subscriber.Close()

			close(dataChannel)

			poller.shutdown <- true

			return

		default:

			batchData, err := poller.subscriber.RecvBytes(0)

			fmt.Println("batchData:", len(batchData))

			if err != nil {

				log.Printf("NewPollerServer : Error receiving message: %v", err)

				continue
			}

			var events []Events

			err = json.Unmarshal(batchData, &events)

			if err != nil {

				log.Printf("NewPollerServer : Error unmarshalling message: %v", err)

				continue
			}

			dataChannel <- events
		}
	}
}

func (poller *PollerServer) Shutdown() {

	poller.shutdown <- true

	poller.context.Term()

	<-poller.shutdown
}
