package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"reportdb/config"
	"strconv"
	"sync"
	"time"
)

func GetPartition(objectID uint32) uint8 {

	return uint8(objectID % uint32(config.Partitions))
}

func GetIndexFilePath(counterID uint16, t time.Time) string {

	year := strconv.Itoa(t.Year())

	month := fmt.Sprintf("%02d", t.Month())

	day := fmt.Sprintf("%02d", t.Day())

	datePath := filepath.Join(config.BaseDir, year, month, day)

	counterDir := filepath.Join(datePath, "counter_"+strconv.Itoa(int(counterID)))

	return filepath.Join(counterDir, "index.json")
}

func GetPartitionFilePath(counterID uint16, partition uint8, t time.Time) string {

	year := strconv.Itoa(t.Year())

	month := fmt.Sprintf("%02d", t.Month())

	day := fmt.Sprintf("%02d", t.Day())

	datePath := filepath.Join(config.BaseDir, year, month, day)

	counterDir := filepath.Join(datePath, "counter_"+strconv.Itoa(int(counterID)))

	partitionFile := filepath.Join(counterDir, fmt.Sprintf("partition_%d.bin", partition))

	return partitionFile
}

func InitFiles() {

	for counterID := uint16(1); counterID <= config.CounterCount; counterID++ {

		config.FileHandles[counterID] = make(map[uint8]*config.FileHandle)

		for partition := uint8(0); partition < config.Partitions; partition++ {

			config.FileHandles[counterID][partition] = &config.FileHandle{

				Lock: &sync.RWMutex{},
			}
		}
	}

	config.Mu = &sync.RWMutex{}
}

func GetDataFile(counterID uint16, partition uint8) (*config.FileHandle, error) {

	config.InitOnce.Do(InitFiles)

	handle := config.FileHandles[counterID][partition]

	if handle.File != nil {

		return handle, nil
	}

	partitionFile := GetPartitionFilePath(counterID, partition, time.Now())

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

	handle.File = file

	handle.Offset = fileInfo.Size()

	handle.AvailableSize = fileInfo.Size()

	return handle, nil
}
