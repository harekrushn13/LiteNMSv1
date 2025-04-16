package main

import (
	"fmt"
	"log"
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

	query := common.Query{

		CounterID: 3,

		ObjectID: 3,

		From: 1744636048,

		To: uint32(time.Now().Unix() + 500),
	}

	ticker := time.NewTicker(3 * time.Second)

	defer ticker.Stop()

	for {

		select {

		case <-ticker.C:

			go generateQuery(client, query)
		}

	}
}

func generateQuery(client *Client, query common.Query) {

	response, err := client.ExecuteQuery(query)

	if err != nil {

		log.Printf("Query failed: %v", err)

		return
	}

	fmt.Printf("Received response: %+v\n", len(response.Data.([]interface{})))
}
