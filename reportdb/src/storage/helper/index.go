package helper

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type IndexManager struct {
	indexHandles map[uint8]*IndexHandle // indexHandles[indexId]

	handleMutex *sync.RWMutex

	baseDir string // ./database/YYYY/MM/DD/counter_1
}

type IndexHandle struct {
	indexMapping map[uint32][]*IndexEntry // IndexMap[key][]*IndexEntry

	indexMutex *sync.RWMutex
}

type IndexEntry struct {
	TimeStamp uint32 `json:"timestamp"`

	Offset int64 `json:"offset"`
}

func NewIndexManager(baseDir string) *IndexManager {

	return &IndexManager{

		indexHandles: make(map[uint8]*IndexHandle),

		handleMutex: &sync.RWMutex{},

		baseDir: baseDir,
	}
}

func (indexManager *IndexManager) Update(key uint32, offset int64, timestamp uint32, indexId uint8) {

	indexHandle := indexManager.getIndexHandle(indexId)

	indexHandle.indexMutex.Lock()

	defer indexHandle.indexMutex.Unlock()

	entryList, exists := indexHandle.indexMapping[key]

	if exists {

		indexHandle.indexMapping[key] = append(entryList, &IndexEntry{TimeStamp: timestamp, Offset: offset})

		return
	}

	if len(indexHandle.indexMapping) == 0 {

		indexFilePath := indexManager.baseDir + "/index_" + strconv.Itoa(int(indexId)) + ".json"

		err := loadIndexFile(indexFilePath, indexHandle, indexId)

		if err != nil {

			log.Printf("indexManager.loadIndexFile error: %v", err)

			return
		}

		if err := os.MkdirAll(filepath.Dir(indexFilePath), 0755); err != nil {

			log.Printf("Error creating index directory: %v", err)

			return
		}

	}

	indexHandle.indexMapping[key] = append(indexHandle.indexMapping[key], &IndexEntry{TimeStamp: timestamp, Offset: offset})

}

func (indexManager *IndexManager) GetIndexMapEntryList(objectId uint32, indexId uint8) ([]*IndexEntry, error) {

	indexHandle := indexManager.getIndexHandle(indexId)

	indexHandle.indexMutex.RLock()

	entryList, exists := indexHandle.indexMapping[objectId]

	indexHandle.indexMutex.RUnlock()

	if exists {

		return entryList, nil
	}

	indexHandle.indexMutex.Lock()

	defer indexHandle.indexMutex.Unlock()

	if entryList, exists = indexHandle.indexMapping[objectId]; exists {

		return entryList, nil
	}

	if len(indexHandle.indexMapping) == 0 {

		indexFilePath := indexManager.baseDir + "/index_" + strconv.Itoa(int(indexId)) + ".json"

		err := loadIndexFile(indexFilePath, indexHandle, indexId)

		if err != nil {

			return nil, fmt.Errorf("indexManager.loadIndexFile error: %v", err)
		}
	}

	return indexHandle.indexMapping[objectId], nil
}

func (indexManager *IndexManager) getIndexHandle(indexId uint8) *IndexHandle {

	indexManager.handleMutex.RLock()

	handle, exists := indexManager.indexHandles[indexId]

	indexManager.handleMutex.RUnlock()

	if exists {

		return handle
	}

	indexManager.handleMutex.Lock()

	defer indexManager.handleMutex.Unlock()

	if handle, exists = indexManager.indexHandles[indexId]; exists {

		return handle
	}

	indexManager.indexHandles[indexId] = &IndexHandle{

		indexMapping: make(map[uint32][]*IndexEntry),

		indexMutex: &sync.RWMutex{},
	}

	return indexManager.indexHandles[indexId]
}

func loadIndexFile(indexFilePath string, handle *IndexHandle, indexId uint8) error {

	if _, err := os.Stat(indexFilePath); err != nil {

		return nil
	}

	data, err := os.ReadFile(indexFilePath)

	if err != nil {

		return fmt.Errorf("error reading index file: %v", err)
	}

	if err := json.Unmarshal(data, &handle.indexMapping); err != nil {

		return fmt.Errorf("error parsing index map: %v", err)
	}

	return nil
}

func (indexManager *IndexManager) Save() error {

	indexManager.handleMutex.Lock()

	defer indexManager.handleMutex.Unlock()

	for index, handle := range indexManager.indexHandles {

		handle.indexMutex.Lock()

		indexFilePath := indexManager.baseDir + "/index_" + strconv.Itoa(int(index)) + ".json"

		if err := os.MkdirAll(filepath.Dir(indexFilePath), 0755); err != nil {

			handle.indexMutex.Unlock()

			return err
		}

		data, err := json.MarshalIndent(handle.indexMapping, "", "  ")

		if err != nil {

			handle.indexMutex.Unlock()

			return err
		}

		if err := os.WriteFile(indexFilePath, data, 0644); err != nil {

			handle.indexMutex.Unlock()

			return err
		}

		handle.indexMutex.Unlock()
	}

	return nil
}

func (indexManager *IndexManager) GetValidOffsets(entryList []*IndexEntry, from uint32, to uint32) ([]int64, error) {

	if len(entryList) == 0 {

		return nil, nil
	}

	start := customBinarySearch(entryList, func(timestamp uint32) bool {

		return timestamp >= from

	}, true)

	if start == -1 {

		return nil, nil
	}

	end := customBinarySearch(entryList, func(ts uint32) bool {

		return ts <= to

	}, false)

	if end == -1 || start > end {

		return nil, nil
	}

	validOffsets := make([]int64, 0, end-start+1)

	for i := start; i <= end; i++ {

		validOffsets = append(validOffsets, entryList[i].Offset)
	}

	return validOffsets, nil

}

func customBinarySearch(entries []*IndexEntry, condition func(uint32) bool, searchFirst bool) int {

	low := 0

	high := len(entries) - 1

	result := -1

	for low <= high {

		mid := (low + high) / 2

		timestamp := entries[mid].TimeStamp

		if condition(timestamp) {

			result = mid

			if searchFirst {

				high = mid - 1

			} else {

				low = mid + 1
			}

		} else {

			if searchFirst {

				low = mid + 1

			} else {

				high = mid - 1
			}
		}
	}

	return result
}
