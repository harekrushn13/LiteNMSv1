package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	apiURL          = "http://localhost:8080/lnms/query"
	testFile        = "./test/test.json"
	totalUsers      = 24
	requestsPerUser = 30
)

type RequestBody struct {
	CounterID      uint16   `msgpack:"counter_id" json:"counter_id" binding:"required"`
	ObjectIDs      []uint32 `msgpack:"object_ids" json:"object_ids,omitempty"`
	From           uint32   `msgpack:"from" json:"from" binding:"required"`
	To             uint32   `msgpack:"to" json:"to" binding:"required"`
	Aggregation    string   `msgpack:"aggregation" json:"aggregation,omitempty" binding:"required"`
	GroupByObjects bool     `msgpack:"group_by_objects" json:"group_by_objects,omitempty"`
	Interval       int      `msgpack:"interval" json:"interval,omitempty"`
}

var (
	totalDuration time.Duration
	totalRequests int
	mutex         sync.Mutex
)

func readTestJSON() ([]RequestBody, error) {
	file, err := os.ReadFile(testFile)
	if err != nil {
		return nil, err
	}

	var requests []RequestBody
	err = json.Unmarshal(file, &requests)
	if err != nil {
		return nil, err
	}

	if len(requests) != totalUsers {
		return nil, fmt.Errorf("expected %d request bodies, found %d", totalUsers, len(requests))
	}

	return requests, nil
}

func sendRequest(request RequestBody) {
	start := time.Now()

	reqBody, err := json.Marshal(request)
	if err != nil {
		log.Println("Error marshalling request body:", err)
		return
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return
	}

	duration := time.Since(start)

	mutex.Lock()
	totalDuration += duration
	totalRequests++
	mutex.Unlock()
}

func simulateUser(request RequestBody, userID int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < requestsPerUser; i++ {
		log.Printf("User %d sending request %d", userID, i+1)
		sendRequest(request)
		time.Sleep(2 * time.Second)
	}
}

func main() {
	requests, err := readTestJSON()
	if err != nil {
		log.Fatalf("Error reading test.json: %v", err)
	}

	var wg sync.WaitGroup

	for userID, request := range requests {
		wg.Add(1)
		go simulateUser(request, userID, &wg)
	}

	wg.Wait()

	if totalRequests > 0 {
		avgDuration := totalDuration / time.Duration(totalRequests)
		fmt.Printf("Total Requests: %d\n", totalRequests)
		fmt.Printf("Average Response Time: %v\n", avgDuration)
	} else {
		fmt.Println("No requests were made.")
	}
}
