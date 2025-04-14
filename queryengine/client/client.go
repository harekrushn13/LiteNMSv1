package client

import (
	"errors"
	"github.com/google/uuid"
	. "queryengine/common"
	. "queryengine/engine"
	"time"
)

type Client struct {
	engine *QueryEngine
}

func NewClient() (*Client, error) {

	engine, err := NewQueryEngine()

	if err != nil {

		return nil, err
	}

	return &Client{engine: engine}, nil
}

func (client *Client) ExecuteQuery(query Query) (*Response, error) {

	if query.RequestID == "" {

		query.RequestID = uuid.New().String()
	}

	query.Timestamp = time.Now()

	responseChan, err := client.engine.SubmitQuery(query)

	if err != nil {

		return nil, err
	}

	select {

	case response := <-responseChan:

		if response.Error != "" {

			return nil, errors.New(response.Error)
		}

		return &response, nil

	}
}

func (client *Client) Close() {

	client.engine.Shutdown()
}
