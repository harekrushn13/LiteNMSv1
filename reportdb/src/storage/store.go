package storage

import (
	"encoding/binary"
	"fmt"
	. "reportdb/storage/helper"
	. "reportdb/utils"
)

type StoreEngine struct {
	fileManager *FileManager

	indexManager *IndexManager

	baseDir string // ex. ./src/storage/database/YYYY/MM/DD/counter_1

	isUsedPut bool

	lastSave int64
}

func NewStorageEngine(baseDir string) *StoreEngine {

	return &StoreEngine{
		fileManager: NewFileManager(baseDir),

		indexManager: NewIndexManager(baseDir),

		baseDir: baseDir,
	}
}

func (store *StoreEngine) Put(objectId uint32, timestamp uint32, data []byte) error {

	store.isUsedPut = true

	fileId := getPartitionId(objectId)

	handle, err := store.fileManager.GetHandle(fileId)

	if err != nil {

		return fmt.Errorf("fileManager.GetHandle(%d): %v", fileId, err)
	}

	if err := store.fileManager.CheckCapacity(handle, int64(len(data))); err != nil {

		return fmt.Errorf("fileManager.CheckCapacity(%d): %v", fileId, err)
	}

	handle.Lock.Lock()

	defer handle.Lock.Unlock()

	offset := handle.Offset

	copy(handle.MmapData[offset:], data)

	handle.Offset += int64(len(data))

	store.indexManager.Update(objectId, offset, timestamp, fileId)

	return nil
}

func (store *StoreEngine) Get(objectId uint32, from uint32, to uint32) ([][]byte, error) {

	fileId := getPartitionId(objectId)

	entryList, err := store.indexManager.GetIndexMapEntryList(objectId, fileId)

	if err != nil {

		return nil, fmt.Errorf("store.indexManager.GetIndexMapEntryList error: %v", err)
	}

	validOffsets, err := store.indexManager.GetValidOffsets(entryList, from, to)

	if err != nil {

		return nil, fmt.Errorf("failed to get valid offsets: %v", err)
	}

	handle, err := store.fileManager.GetHandle(fileId)

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

func getPartitionId(objectID uint32) uint8 {

	return uint8(objectID % uint32(GetPartitions()))
}
