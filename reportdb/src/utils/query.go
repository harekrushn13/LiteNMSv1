package utils

import "time"

type Query struct {
	RequestID string `json:"request_id"`

	CounterId uint16 `json:"counter_id"`

	ObjectId uint32 `json:"object_id"`

	From uint32 `json:"from"`

	To uint32 `json:"to"`

	Timestamp time.Time `json:"timestamp"`
}
