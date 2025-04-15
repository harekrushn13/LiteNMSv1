package utils

import "time"

type Query struct {
	Timestamp time.Time `json:"timestamp"`

	RequestID uint64 `json:"request_id"`

	ObjectId uint32 `json:"object_id"`

	From uint32 `json:"from"`

	To uint32 `json:"to"`

	CounterId uint16 `json:"counter_id"`
}

type Response struct {
	Timestamp time.Time `json:"timestamp"`

	RequestID uint64 `json:"request_id"`

	Error string `json:"error,omitempty"`

	Data interface{} `json:"data"`
}
