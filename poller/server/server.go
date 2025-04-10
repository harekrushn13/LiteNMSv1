package server

import (
	"encoding/json"
	"github.com/pebbe/zmq4"
	"log"
	. "poller/polling"
)

func ZMQServer(dataChannel <-chan []Events) {

	context, err := zmq4.NewContext()

	if err != nil {

		log.Fatal(err)
	}

	go func(context *zmq4.Context) {

		publisher, err := context.NewSocket(zmq4.PUB)

		defer publisher.Close()

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

		return

	}(context)

}
