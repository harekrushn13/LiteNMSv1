package config

import "time"

type RowData struct {
	ObjectId uint32

	CounterId uint16

	Timestamp uint32

	Value interface{}
}

const (
	PollingInterval time.Duration = 1 * time.Second

	BatchTime time.Duration = 2500 * time.Millisecond

	StopTime time.Duration = 10 * time.Second
)
