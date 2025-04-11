package server

import (
	"fmt"
	"github.com/pebbe/zmq4"
	. "reportdb/datastore/query"
	. "reportdb/storage"
	. "reportdb/utils"
)

func ZMQQuery(storePool *StorePool) (*zmq4.Context, error) {

	context, err := zmq4.NewContext()

	if err != nil {

		return nil, fmt.Errorf("ZMQQuery : failed to create ZMQ context: %v", err)
	}

	socket, err := context.NewSocket(zmq4.REP)

	if err != nil {

		context.Term()

		return nil, fmt.Errorf("ZMQQuery : failed to create ZMQ socket: %v", err)
	}

	err = socket.Bind("tcp://*:6001")

	if err != nil {

		socket.Close()

		context.Term()

		return nil, fmt.Errorf("ZMQQuery : failed to bind ZMQ socket: %v", err)
	}

	go func() {

		defer socket.Close()

		for {

			if GlobalShutdown {

				return
			}

			HandleQuery(socket, storePool)
		}
	}()

	return context, nil
}
