package server

import (
	"encoding/json"
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	. "reportdb/utils"
	"sync"
)

func ZMQServer(dataChannel chan []Events, waitGroup *sync.WaitGroup) error {

	context, err := zmq4.NewContext()

	if err != nil {

		return fmt.Errorf("ZMQServer : Error creating ZMQ context: %v", err)
	}

	waitGroup.Add(1)

	go func(context *zmq4.Context, dataChannel chan []Events, waitGroup *sync.WaitGroup) {

		defer waitGroup.Done()

		subscriber, err := context.NewSocket(zmq4.SUB)

		defer subscriber.Close()

		if err != nil {

			log.Printf("ZMQServer : Error creating subscriber: %v", err)

			return
		}

		err = subscriber.Connect("tcp://localhost:6000")

		if err != nil {

			log.Printf("ZMQServer : Error connecting to subscriber: %v", err)

			return
		}

		err = subscriber.SetSubscribe("")

		if err != nil {

			log.Printf("ZMQServer : Error subscribing to zmq4: %v", err)

			return
		}

		for {

			batchData, err := subscriber.Recv(0)

			if err != nil {

				log.Printf("ZMQServer : Error receiving message: %v", err)

				return
			}

			var events []Events

			err = json.Unmarshal([]byte(batchData), &events)

			if err != nil {

				log.Printf("ZMQServer : Error unmarshalling message: %v", err)

				return
			}

			dataChannel <- events
		}

		return

	}(context, dataChannel, waitGroup)

	return nil
}
