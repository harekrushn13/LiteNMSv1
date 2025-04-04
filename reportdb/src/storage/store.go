package storage

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

// Storing

func (store *StorageEngine) Save(row config.RowData) error {

	handle, err := store.getFileHandle(row)

	if err != nil {

		return fmt.Errorf("failed to get file handle: %w", err)
	}

	data, err := store.encodeData(row)

	if err != nil {

		return fmt.Errorf("failed to encode data: %w", err)
	}

	if err := store.ensureCapacity(handle, int64(len(data))); err != nil {

		return fmt.Errorf("failed to ensure capacity: %w", err)
	}

	offset, err := store.writeData(handle, data)

	if err != nil {

		return fmt.Errorf("failed to write data: %w", err)
	}

	store.UpdateIndex(row, offset)

	return nil
}

func (store *StorageEngine) getFileHandle(row config.RowData) (*helper.FileHandle, error) {

	partition := store.fileCfg.GetPartition(row.ObjectId)

	handle, err := store.fileCfg.GetHandle(row.CounterId, partition)

	if err != nil {

		return nil, fmt.Errorf("failed to get handle for counter %d partition %d: %w", row.CounterId, partition, err)
	}

	return handle, nil
}

func (store *StorageEngine) ensureCapacity(handle *helper.FileHandle, requiredSize int64) error {

	handle.Lock.Lock()

	defer handle.Lock.Unlock()

	if handle.Offset+requiredSize <= handle.AvailableSize {

		return nil
	}

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

	return nil
}

func (store *StorageEngine) encodeData(row config.RowData) ([]byte, error) {

	dataType, exists := config.CounterTypeMapping[row.CounterId]

	if !exists {

		return nil, fmt.Errorf("unknown counter ID: %d", row.CounterId)
	}

	switch dataType {

	case config.TypeUint64:

		val, ok := row.Value.(uint64)

		if !ok {

			return nil, fmt.Errorf("invalid uint64 value for counter %d", row.CounterId)
		}

		data := make([]byte, 8)

		binary.LittleEndian.PutUint64(data, val)

		return data, nil

	case config.TypeFloat64:

		val, ok := row.Value.(float64)

		if !ok {

			return nil, fmt.Errorf("invalid float64 value for counter %d", row.CounterId)
		}

		data := make([]byte, 8)

		binary.LittleEndian.PutUint64(data, math.Float64bits(val))

		return data, nil

	case config.TypeString:

		str, ok := row.Value.(string)

		if !ok {

			return nil, fmt.Errorf("invalid string value for counter %d", row.CounterId)
		}

		data := make([]byte, 4+len(str))

		binary.LittleEndian.PutUint32(data, uint32(len(str)))

		copy(data[4:], []byte(str))

		return data, nil

	default:

		return nil, fmt.Errorf("unsupported data type: %d", dataType)
	}
}

func (store *StorageEngine) writeData(handle *helper.FileHandle, data []byte) (int64, error) {

	handle.Lock.Lock()

	defer handle.Lock.Unlock()

	offset := handle.Offset

	copy(handle.MmapData[offset:], data)

	handle.Offset += int64(len(data))

	return offset, nil
}

func (store *StorageEngine) UpdateIndex(row config.RowData, offset int64) {

	store.indexCfg.Update(row.CounterId, row.ObjectId, offset, row.Timestamp)
}

// Fetching

func (store *StorageEngine) Get(counterID uint16, objectID uint32, day time.Time, offsets []int64, dataType config.DataType) ([]interface{}, error) {

	partition := store.fileCfg.GetPartition(objectID)

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
