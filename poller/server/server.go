package server

import (
	"encoding/json"
	"github.com/pebbe/zmq4"
	"log"
	. "poller/polling"
	"sync"
)

func ZMQServer(dataChannel <-chan []Events, waitGroup *sync.WaitGroup) {

	context, err := zmq4.NewContext()

	if err != nil {

		log.Fatal(err)
	}

	waitGroup.Add(1)

	go func(context *zmq4.Context, waitGroup *sync.WaitGroup) {

		defer waitGroup.Done()

		publisher, err := context.NewSocket(zmq4.PUB)

		if err != nil {

			log.Fatal(err)
		}

		publisher.Bind("tcp://*:6000")

		for batch := range dataChannel {

			jsonData, err := json.Marshal(batch)

			if err != nil {

				log.Println("Error serializing batch:", err)

				continue
			}

			_, err = publisher.SendBytes(jsonData, 0)

			if err != nil {

				log.Println("Error sending batch:", err)
			}
		}

		publisher.Close()

		context.Term()

		return

	}(context, waitGroup)

	return
}
