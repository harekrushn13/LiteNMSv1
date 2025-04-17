package storage

import (
	"encoding/binary"
	"fmt"
	. "reportdb/utils"
)

type StoreEngine struct {
	fileManager *FileManager

	indexManager *IndexManager

	baseDir string // ex. /reportdb/database/YYYY/MM/DD/counter_1

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

func (store *StoreEngine) Put(key uint32, timestamp uint32, data []byte) error {

	store.isUsedPut = true

	fileId, err := getPartitionId(key)

	if err != nil {

		return err
	}

	handle, err := store.fileManager.GetHandle(fileId)

	if err != nil {

		return fmt.Errorf("fileManager.GetHandle(%d): %v", fileId, err)
	}

	if err := store.fileManager.CheckCapacity(handle, int64(len(data))); err != nil {

		return fmt.Errorf("fileManager.CheckCapacity(%d): %v", fileId, err)
	}

	handle.lock.Lock()

	defer handle.lock.Unlock()

	offset := handle.Offset

	copy(handle.mappedBuffer[offset:], data)

	handle.Offset += int64(len(data))

	binary.LittleEndian.PutUint64(handle.mappedBuffer[:8], uint64(handle.Offset))

	store.indexManager.Update(key, offset, timestamp, fileId)

	return nil
}

func (store *StoreEngine) Get(key uint32, from uint32, to uint32) ([][]byte, error) {

	fileId, err := getPartitionId(key)

	if err != nil {

		return nil, err
	}

	entryList, err := store.indexManager.GetIndexMapEntryList(key, fileId)

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

	handle.lock.RLock()

	defer handle.lock.RUnlock()

	var dayResult [][]byte

	for _, offset := range validOffsets {

		if offset+4 > int64(len(handle.mappedBuffer)) {

			return nil, fmt.Errorf("offset %d out of bounds for length prefix", offset)
		}

		length := binary.LittleEndian.Uint32(handle.mappedBuffer[offset : offset+4])

		if offset+4+int64(length) > int64(len(handle.mappedBuffer)) {

			return nil, fmt.Errorf("record at offset %d extends beyond file bounds", offset)
		}

		dayResult = append(dayResult, handle.mappedBuffer[offset+4:offset+4+int64(length)])
	}

	return dayResult, nil
}

func getPartitionId(key uint32) (uint8, error) {

	partitions, err := GetPartitions()

	if err != nil {

		return 0, fmt.Errorf("GetPartitions error: %v", err)
	}

	index := uint8(key % uint32(partitions))

	if index < 0 || index >= partitions {

		return 0, fmt.Errorf("invalid partition index: %d", index)
	}

	return index, nil
}
