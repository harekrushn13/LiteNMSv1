package common

import "time"

type Query struct {
	RequestID string `json:"request_id"`

	CounterID uint16 `json:"counter_id"`

	ObjectID uint32 `json:"object_id"`

	From uint32 `json:"from"`

	To uint32 `json:"to"`

	Timestamp time.Time `json:"timestamp"`
}

type Response struct {
	RequestID string `json:"request_id"`

	Data interface{} `json:"data"`

	Error string `json:"error,omitempty"`

	Timestamp time.Time `json:"timestamp"`
}

const (
	QueryTimeout = 30 * time.Second

	QueryAddress = "tcp://localhost:6001"

	ResponseAddress = "tcp://localhost:6002"
)
