package main

import (
	"encoding/json"
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"time"
)

type Query struct {
	CounterId uint16

	ObjectId uint32

	From uint32

	To uint32
}

func NewQuery(counterID uint16, objectID uint32, from uint32, to uint32) *Query {

	return &Query{

		CounterId: counterID,

		ObjectId: objectID,

		From: from,

		To: to,
	}
}

func createSocket() (*zmq4.Socket, *zmq4.Context, error) {

	context, err := zmq4.NewContext()

	if err != nil {

		return nil, nil, fmt.Errorf("failed to create context: %v", err)
	}

	socket, err := context.NewSocket(zmq4.REQ)

	if err != nil {

		context.Term()

		return nil, nil, fmt.Errorf("failed to create socket: %v", err)
	}

	err = socket.Connect("tcp://localhost:6001") // Changed from Bind to Connect

	if err != nil {

		socket.Close()

		context.Term()

		return nil, nil, fmt.Errorf("failed to connect socket: %v", err)
	}

	return socket, context, nil

}

func queryResult(counterID uint16, objectID uint32, from uint32, to uint32, socket *zmq4.Socket) error {

	query := NewQuery(counterID, objectID, from, to)

	jsonData, err := json.Marshal(query)

	if err != nil {

		return fmt.Errorf("error serializing query: %v", err)
	}

	fmt.Println("sendBytes")

	_, err = socket.SendBytes(jsonData, 0)

	if err != nil {

		return fmt.Errorf("error sending query: %v", err)
	}

	fmt.Println("recvBytes")

	reply, err := socket.RecvBytes(0)

	if err != nil {

		return fmt.Errorf("error receiving response: %v", err)
	}

	fmt.Printf("Reply: %s\n", string(reply))

	return nil

}

func main() {

	totalIterations := 10

	socket, context, err := createSocket()

	if err != nil {

		log.Fatal(err)
	}

	defer func() {

		socket.Close()

		context.Term()
	}()

	for i := 0; i < totalIterations; i++ {

		fmt.Printf("Iteration: %d\n", i)

		err := queryResult(3, 3, 1744342654, uint32(time.Now().Unix())+12, socket)

		if err != nil {

			log.Println(err)
		}

		time.Sleep(2 * time.Second)
	}
}
