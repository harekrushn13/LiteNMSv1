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

func (store *StoreEngine) Put(key uint32, data []byte) error {

	store.isUsedPut = true

	fileId, err := getPartitionId(key)

	if err != nil {

		return err
	}

	entryList, err := store.indexManager.GetIndexMapEntryList(key, fileId, store.isUsedPut)

	if err != nil {

		return fmt.Errorf("GetIndexMapEntryList error: %v", err)
	}

	handle, err := store.fileManager.GetHandle(fileId)

	if err != nil {

		return fmt.Errorf("fileManager.GetHandle(%d): %v", fileId, err)
	}

	entryList, err = store.fileManager.CheckCapacity(handle, entryList, int64(len(data)))

	if err != nil {

		return fmt.Errorf("fileManager.CheckCapacity(%d): %v", fileId, err)
	}

	handle.lock.Lock()

	defer handle.lock.Unlock()

	offset := entryList[len(entryList)-1].EntryEnd

	copy(handle.mappedBuffer[offset:], data)

	entryList[len(entryList)-1].EntryEnd += int64(len(data))

	store.indexManager.Update(key, fileId, entryList)

	return nil
}

func (store *StoreEngine) Get(key uint32, from uint32, to uint32) ([][]byte, error) {

	fileId, err := getPartitionId(key)

	if err != nil {

		return nil, err
	}

	entryList, err := store.indexManager.GetIndexMapEntryList(key, fileId, store.isUsedPut)

	if err != nil {

		return nil, fmt.Errorf("store.indexManager.GetIndexMapEntryList error: %v", err)
	}

	handle, err := store.fileManager.GetHandle(fileId)

	if err != nil {

		return nil, fmt.Errorf("failed to get handle for partition %d: %v", fileId, err)
	}

	handle.lock.RLock()

	defer handle.lock.RUnlock()

	var dayResult [][]byte

	for _, entry := range entryList {

		start := entry.EntryStart

		end := entry.EntryEnd

		for start < end {

			length := binary.LittleEndian.Uint32(handle.mappedBuffer[start : start+4])

			timestamp := binary.LittleEndian.Uint32(handle.mappedBuffer[start+4 : start+8])

			if timestamp >= from && timestamp <= to {

				dayResult = append(dayResult, handle.mappedBuffer[start+8:start+8+int64(length)])
			}

			start = start + 8 + int64(length)
		}
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
