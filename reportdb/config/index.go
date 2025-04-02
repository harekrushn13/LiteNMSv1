package config

import (
	"sync"
	"time"
)

type IndexEntry struct {
	StartTime uint32 `json:"starttime"`

	EndTime uint32 `json:"endtime"`

	Offsets []int64 `json:"offsets"`
}

type IndexHandle struct {
	Lock *sync.RWMutex
}

var (
	IndexMap = make(map[uint16]map[uint32]*IndexEntry) // IndexMap[counterID][objectID]

	Mu *sync.RWMutex

	IndexHandles = make(map[uint16]*IndexHandle) // IndexHandles[counterID]

	LastDayChecked time.Time

	InitIndexHandles sync.Once
)
