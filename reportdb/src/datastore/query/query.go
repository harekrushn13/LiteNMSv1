package query

import (
	"encoding/json"
	"github.com/pebbe/zmq4"
	"log"
	. "reportdb/datastore/reader"
	. "reportdb/storage"
	. "reportdb/utils"
)

func HandleQuery(request *zmq4.Socket, storePool *StorePool) {

	queryBytes, err := request.RecvBytes(0)

	if err != nil {

		log.Printf("HandleQuery : Error receiving query: %v", err)

		return
	}

	var query Query

	err = json.Unmarshal(queryBytes, &query)

	if err != nil {

		log.Printf("HandleQuery : Error unmarshalling query: %v", err)

		return
	}

	results, err := FetchData(&query, storePool)

	if err != nil {

		log.Printf("HandleQuery : Error fetching data: %v", err)

		return
	}

	replyBytes, err := json.Marshal(results)

	if err != nil {

		log.Printf("HandleQuery : Error marshalling results: %v", err)

		return
	}

	_, err = request.SendBytes(replyBytes, 0)

	if err != nil {

		log.Printf("HandleQuery : Error sending reply: %v", err)
	}
}
