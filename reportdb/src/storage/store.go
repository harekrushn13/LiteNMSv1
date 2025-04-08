package storage

import (
	"encoding/binary"
	"fmt"
	. "reportdb/src/storage/helper"
	. "reportdb/src/utils"
)

type StorageEngine struct {
	fileCfg *FileManager

	indexCfg *IndexManager

	PartitionCount uint8

	BaseDir string // ex. ./src/storage/database/YYYY/MM/DD/counter_1

	isUsedPut bool

	lastSave int64
}

func NewStorageEngine(baseDir string) *StorageEngine {

	return &StorageEngine{
		fileCfg: NewFileManager(baseDir),

		indexCfg: NewIndexManager(baseDir, 3),

		BaseDir: baseDir,

		PartitionCount: 3,
	}
}

func (store *StorageEngine) Put(objectId uint32, timestamp uint32, data []byte) error {

	store.isUsedPut = true

	fileId := GetPartition(objectId, store.PartitionCount)

	handle, err := store.fileCfg.GetHandle(fileId)

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

	store.indexCfg.Update(objectId, offset, timestamp, fileId)

	return nil
}

func (store *StorageEngine) Get(objectId uint32, from uint32, to uint32) ([][]byte, error) {

	fileId := GetPartition(objectId, store.PartitionCount)

	entryList, err := store.indexCfg.GetIndexMapEntryList(objectId, fileId)

	if err != nil {

		return nil, fmt.Errorf("get index map error: %v", err)
	}

	validOffsets, err := store.indexCfg.GetValidOffsets(entryList, from, to)

	if err != nil {

		return nil, fmt.Errorf("failed to get valid offsets: %v", err)
	}

	handle, err := store.fileCfg.GetHandle(fileId)

	if err != nil {

		return nil, fmt.Errorf("failed to get handle for partition %d: %v", fileId, err)
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
