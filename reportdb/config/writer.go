package config

import (
	"os"
	"sync"
)

type FileHandle struct {
	File *os.File

	Offset int64

	Lock *sync.RWMutex

	AvailableSize int64

	MmapData []byte
}

var (
	FileHandles = make(map[uint16]map[uint8]*FileHandle) // FileHandles[counterID][partition]

	InitOnce sync.Once
)
