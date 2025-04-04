package engine

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"reportdb/config"
	"reportdb/src/storage/helper"
	"syscall"
	"time"
)

type StorageEngine struct {
	fileCfg *helper.FileManager

	indexCfg *helper.IndexManager

	baseDir string
}

func NewStorageEngine(fileCfg *helper.FileManager, indexCfg *helper.IndexManager, baseDir string) *StorageEngine {

	return &StorageEngine{
		fileCfg: fileCfg,

		indexCfg: indexCfg,

		baseDir: baseDir,
	}
}

func (store *StorageEngine) Save(row config.RowData) error {

	partition := store.fileCfg.GetPartition(row.ObjectId)

	handle, err := store.fileCfg.GetHandle(row.CounterId, partition)

	if err != nil {

		return err
	}

	handle.Lock.Lock()

	defer handle.Lock.Unlock()

	dataType, exists := config.CounterTypeMapping[row.CounterId]

	if !exists {

		return fmt.Errorf("unknown counter ID: %d", row.CounterId)
	}

	var requiredSize int64

	switch dataType {

	case config.TypeUint64, config.TypeFloat64:

		requiredSize = 8

	case config.TypeString:

		str, ok := row.Value.(string)

		if !ok {

			return fmt.Errorf("invalid string value for counter %d", row.CounterId)
		}

		requiredSize = 4 + int64(len(str))
	}

	if handle.Offset+requiredSize > handle.AvailableSize {

		if handle.MmapData != nil {

			if err := syscall.Munmap(handle.MmapData); err != nil {

				return fmt.Errorf("munmap failed: %v", err)
			}
		}

		handle.AvailableSize = handle.Offset + store.fileCfg.FileGrowth // just 64 byte adding

		if err := handle.File.Truncate(handle.AvailableSize); err != nil {

			return fmt.Errorf("failed to grow file: %v", err)
		}

		data, err := syscall.Mmap(int(handle.File.Fd()), 0, int(handle.AvailableSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

		if err != nil {

			return fmt.Errorf("mmap failed: %v", err)
		}

		handle.MmapData = data

	}

	lastOffset := handle.Offset

	switch dataType {

	case config.TypeUint64:

		val, ok := row.Value.(uint64)

		if !ok {

			return fmt.Errorf("invalid uint64 value for counter %d", row.CounterId)
		}

		binary.LittleEndian.PutUint64(handle.MmapData[handle.Offset:], val)

		handle.Offset += 8

	case config.TypeFloat64:

		val, ok := row.Value.(float64)

		if !ok {

			return fmt.Errorf("invalid float64 value for counter %d", row.CounterId)
		}

		binary.LittleEndian.PutUint64(handle.MmapData[handle.Offset:], math.Float64bits(val))

		handle.Offset += 8

	case config.TypeString:

		str, ok := row.Value.(string)

		if !ok {

			return fmt.Errorf("invalid string value for counter %d", row.CounterId)
		}

		length := uint32(len(str))

		binary.LittleEndian.PutUint32(handle.MmapData[handle.Offset:], length)

		copy(handle.MmapData[handle.Offset+4:], []byte(str))

		handle.Offset += 4 + int64(length)
	}

	store.indexCfg.Update(row.CounterId, row.ObjectId, lastOffset, row.Timestamp)

	return nil
}

func (store *StorageEngine) Get(counterID uint16, objectID uint32, day time.Time, offsets []int64, dataType config.DataType) ([]interface{}, error) {

	partition := uint8(objectID % uint32(store.fileCfg.PartitionCount))

	dataPath := store.fileCfg.GetPartitionFilePath(counterID, partition, day)

	file, err := os.Open(dataPath)

	if err != nil {

		return nil, err
	}

	defer file.Close()

	fileInfo, err := file.Stat()

	if err != nil {

		return nil, err
	}

	data, err := syscall.Mmap(int(file.Fd()), 0, int(fileInfo.Size()), syscall.PROT_READ, syscall.MAP_SHARED)

	if err != nil {

		return nil, err
	}

	defer syscall.Munmap(data)

	var results []interface{}

	for _, offset := range offsets {

		var value interface{}

		var err error

		switch dataType {

		case config.TypeUint64:

			if offset+8 > int64(len(data)) {

				return nil, fmt.Errorf("offset %d out of bounds (file size: %d)", offset, len(data))
			}

			value = binary.LittleEndian.Uint64(data[offset : offset+8])

		case config.TypeFloat64:

			if offset+8 > int64(len(data)) {

				return nil, fmt.Errorf("offset %d out of bounds (file size: %d)", offset, len(data))
			}

			value = math.Float64frombits(binary.LittleEndian.Uint64(data[offset : offset+8]))

		case config.TypeString:

			if offset+4 > int64(len(data)) {

				return nil, fmt.Errorf("offset %d out of bounds (file size: %d)", offset, len(data))
			}

			length := int(data[offset])

			if offset+4+int64(length) > int64(len(data)) {

				return nil, fmt.Errorf("string data at offset %d exceeds file bounds", offset)
			}

			value = string(data[offset+4 : offset+4+int64(length)])
		}

		if err != nil {

			return nil, err
		}

		results = append(results, value)

	}

	return results, nil
}
