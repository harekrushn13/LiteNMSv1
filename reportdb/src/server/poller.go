package server

import (
	"encoding/json"
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	. "reportdb/utils"
)

func ZMQPoller(dataChannel chan []Events) (*zmq4.Context, error) {

	context, err := zmq4.NewContext()

	if err != nil {

		return nil, fmt.Errorf("ZMQPoller : Error creating ZMQ context: %v", err)
	}

	go func(context *zmq4.Context, dataChannel chan []Events) {

		subscriber, err := context.NewSocket(zmq4.SUB)

		defer subscriber.Close()

		if err != nil {

			log.Printf("ZMQPoller : Error creating subscriber: %v", err)

			return
		}

		err = subscriber.Connect("tcp://localhost:6000")

		if err != nil {

			log.Printf("ZMQPoller : Error connecting to subscriber: %v", err)

			return
		}

		err = subscriber.SetSubscribe("")

		if err != nil {

			log.Printf("ZMQPoller : Error subscribing to zmq4: %v", err)

			return
		}

		for {

			if GlobalShutdown {

				close(dataChannel)

				return
			}

			batchData, err := subscriber.Recv(0)

			if err != nil {

				log.Printf("ZMQPoller : Error receiving message: %v", err)

				return
			}

			var events []Events

			err = json.Unmarshal([]byte(batchData), &events)

			if err != nil {

				log.Printf("ZMQPoller : Error unmarshalling message: %v", err)

				return
			}

			dataChannel <- events
		}

	}(context, dataChannel)

	return context, nil
}
