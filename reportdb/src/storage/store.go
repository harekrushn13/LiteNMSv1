package storage

import (
	"encoding/binary"
	"fmt"
	. "reportdb/src/storage/helper"
	. "reportdb/src/utils"
	"time"
)

type StorageEngine struct {
	fileCfg *FileManager

	indexCfg *IndexManager

	PartitionCount uint8

	BaseDir string // ex. ./src/storage/database/YYYY/MM/DD/counter_1

	isUsedPut bool

	lastAccess int64

	lastSave int64
}

func NewStorageEngine(baseDir string) *StorageEngine {

	return &StorageEngine{
		fileCfg: NewFileManager(baseDir),

		indexCfg: NewIndexManager(baseDir),

		BaseDir: baseDir,

		PartitionCount: 3,
	}
}

func (store *StorageEngine) Put(objectId uint32, timestamp uint32, data []byte) error {

	store.isUsedPut = true

	store.lastAccess = time.Now().Unix()

	partition := GetPartition(objectId, store.PartitionCount)

	handle, err := store.fileCfg.GetHandle(partition)

	if err != nil {

		return err
	}

	if err := store.fileCfg.EnsureCapacity(handle, int64(len(data))); err != nil {

		return err
	}

	handle.Lock.Lock()

	defer handle.Lock.Unlock()

	offset := handle.Offset

	copy(handle.MmapData[offset:], data)

	handle.Offset += int64(len(data))

	store.indexCfg.Update(objectId, offset, timestamp)

	//err = store.indexCfg.Save()
	//
	//if err != nil {
	//
	//	return err
	//}

	return nil
}

func (store *StorageEngine) Get(objectId uint32, from uint32, to uint32) ([][]byte, error) {

	store.lastAccess = time.Now().Unix()

	entry, err := store.indexCfg.GetIndexMap(objectId)

	if err != nil {

		return nil, fmt.Errorf("get index map error: %v", err)
	}

	if entry.EndTime < from || entry.StartTime > to {

		return nil, fmt.Errorf("data for objectID %d is not within the requested time range", objectId)
	}

	validOffsets, err := store.indexCfg.GetValidOffsets(entry, from, to)

	if err != nil {

		return nil, fmt.Errorf("failed to get valid offsets: %v", err)
	}

	partition := GetPartition(objectId, store.PartitionCount)

	handle, err := store.fileCfg.GetHandle(partition)

	if err != nil {

		return nil, fmt.Errorf("failed to get handle for partition %d: %v", partition, err)
	}

	handle.Lock.RLock()

	defer handle.Lock.RUnlock()

	var dayResults [][]byte

	for _, offset := range validOffsets {

		if offset+4 > int64(len(handle.MmapData)) {

			return nil, fmt.Errorf("offset %d out of bounds for length prefix", offset)
		}

		length := binary.LittleEndian.Uint32(handle.MmapData[offset : offset+4])

		if offset+4+int64(length) > int64(len(handle.MmapData)) {

			return nil, fmt.Errorf("record at offset %d extends beyond file bounds", offset)
		}

		record := make([]byte, length)

		copy(record, handle.MmapData[offset+4:offset+4+int64(length)])

		dayResults = append(dayResults, record)
	}

	return dayResults, nil
}
