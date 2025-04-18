package main

import (
	"fmt"
	"log"
	"math/rand"
	. "queryengine/client"
	"queryengine/common"
	"time"
)

func main() {
	client, err := NewClient()

	if err != nil {

		log.Fatal(err)
	}

	defer client.Close()

	ticker := time.NewTicker(3 * time.Second)

	defer ticker.Stop()

	rand.Seed(time.Now().UnixNano())

	for {

		select {

		case <-ticker.C:

			query := common.Query{

				CounterID: uint16(rand.Intn(3) + 1),

				ObjectID: uint32(rand.Intn(5) + 1),

				From: 1744636048,

				To: uint32(time.Now().Unix() + 700),
			}

			go generateQuery(client, query)
		}

	}
}

func generateQuery(client *Client, query common.Query) {

	response, err := client.RunQuery(query)

	if err != nil {

		log.Printf("Query failed: %v", err)

		return
	}

	fmt.Printf("Received response: %+v %+v\n", len(response.Data.([]interface{})), response.Data.([]interface{}))
}
