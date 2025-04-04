package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type FileHandle struct {
	File *os.File

	Offset int64

	Lock *sync.RWMutex

	AvailableSize int64

	MmapData []byte
}

type FileManager struct {
	handles map[uint16]map[uint8]*FileHandle // handles map[counterID][partition]

	handlesMux sync.RWMutex

	BaseDir string

	PartitionCount uint8

	FileGrowth int64 // byte
}

func NewFileManager(baseDir string) *FileManager {

	return &FileManager{

		handles: make(map[uint16]map[uint8]*FileHandle),

		handlesMux: sync.RWMutex{},

		BaseDir: baseDir,

		PartitionCount: 3,

		FileGrowth: 64,
	}
}

func (fm *FileManager) GetHandle(counterID uint16, partition uint8) (*FileHandle, error) {

	fm.handlesMux.Lock()

	defer fm.handlesMux.Unlock()

	if _, exists := fm.handles[counterID]; !exists {

		fm.handles[counterID] = make(map[uint8]*FileHandle)
	}

	if handle, exists := fm.handles[counterID][partition]; exists {

		return handle, nil
	}

	partitionFile := fm.GetPartitionFilePath(counterID, partition, time.Now())

	counterDir := filepath.Dir(partitionFile)

	if err := os.MkdirAll(counterDir, 0755); err != nil {

		return nil, err
	}

	file, err := os.OpenFile(partitionFile, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {

		return nil, err
	}

	fileInfo, err := file.Stat()

	if err != nil {

		file.Close()

		return nil, err
	}

	handle := &FileHandle{

		File: file,

		Offset: fileInfo.Size(),

		AvailableSize: fileInfo.Size(),

		Lock: &sync.RWMutex{},
	}

	fm.handles[counterID][partition] = handle

	return handle, nil
}

func (fm *FileManager) GetPartitionFilePath(counterID uint16, partition uint8, t time.Time) string {

	year := strconv.Itoa(t.Year())

	month := fmt.Sprintf("%02d", t.Month())

	day := fmt.Sprintf("%02d", t.Day())

	datePath := filepath.Join(fm.BaseDir, year, month, day)

	counterDir := filepath.Join(datePath, "counter_"+strconv.Itoa(int(counterID)))

	partitionFile := filepath.Join(counterDir, fmt.Sprintf("partition_%d.bin", partition))

	return partitionFile
}

func (fm *FileManager) GetPartition(objectID uint32) uint8 {

	return uint8(objectID % uint32(fm.PartitionCount))
}
