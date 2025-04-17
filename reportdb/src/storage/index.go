package storage

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
	indexHandles map[uint8]map[uint32][]*IndexEntry // indexHandles[indexId][key][]*IndexEntry

	lock *sync.RWMutex

	baseDir string // ./database/YYYY/MM/DD/counter_1
}

type IndexEntry struct {
	TimeStamp uint32 `json:"timestamp"`

	Offset int64 `json:"offset"`
}

func NewIndexManager(baseDir string) *IndexManager {

	return &IndexManager{

		indexHandles: make(map[uint8]map[uint32][]*IndexEntry),

		lock: &sync.RWMutex{},

		baseDir: baseDir,
	}
}

func (indexManager *IndexManager) Update(key uint32, offset int64, timestamp uint32, indexId uint8) {

	indexManager.lock.Lock()

	defer indexManager.lock.Unlock()

	indexMap, exists := indexManager.indexHandles[indexId]

	if !exists {

		indexMap = make(map[uint32][]*IndexEntry)

		indexManager.indexHandles[indexId] = indexMap

		indexFilePath := indexManager.baseDir + "/index_" + strconv.Itoa(int(indexId)) + ".json"

		if err := loadIndexFile(indexFilePath, &indexMap); err != nil {

			log.Printf("indexManager.loadIndexFile error: %v", err)

			return
		}

		if err := os.MkdirAll(filepath.Dir(indexFilePath), 0755); err != nil {

			log.Printf("Error creating index directory: %v", err)

			return
		}

	}

	indexMap[key] = append(indexMap[key], &IndexEntry{TimeStamp: timestamp, Offset: offset})
}

// This is function only used by Get

func (indexManager *IndexManager) GetIndexMapEntryList(key uint32, indexId uint8) ([]*IndexEntry, error) {

	indexManager.lock.RLock()

	indexMap, exists := indexManager.indexHandles[indexId]

	indexManager.lock.RUnlock()

	if exists {

		if entryList, exists := indexMap[key]; exists {

			return entryList, nil
		}
	}

	indexManager.lock.Lock()

	defer indexManager.lock.Unlock()

	indexMap, exists = indexManager.indexHandles[indexId]

	if !exists {

		indexMap = make(map[uint32][]*IndexEntry)

		indexManager.indexHandles[indexId] = indexMap

		indexFilePath := indexManager.baseDir + "/index_" + strconv.Itoa(int(indexId)) + ".json"

		if err := loadIndexFile(indexFilePath, &indexMap); err != nil {

			return nil, fmt.Errorf("indexManager.loadIndexFile error: %v", err)
		}

	}

	return indexMap[key], nil
}

func loadIndexFile(indexFilePath string, indexMap *map[uint32][]*IndexEntry) error {

	if _, err := os.Stat(indexFilePath); err != nil {

		return nil
	}

	data, err := os.ReadFile(indexFilePath)

	if err != nil {

		return fmt.Errorf("error reading index file: %v", err)
	}

	if err := json.Unmarshal(data, indexMap); err != nil {

		return fmt.Errorf("error parsing index map: %v", err)
	}

	return nil
}

func (indexManager *IndexManager) Save() error {

	indexManager.lock.Lock()

	defer indexManager.lock.Unlock()

	for index, indexMap := range indexManager.indexHandles {

		indexFilePath := indexManager.baseDir + "/index_" + strconv.Itoa(int(index)) + ".json"

		if err := os.MkdirAll(filepath.Dir(indexFilePath), 0755); err != nil {

			return err
		}

		data, err := json.MarshalIndent(indexMap, "", "  ")

		if err != nil {

			return err
		}

		if err := os.WriteFile(indexFilePath, data, 0644); err != nil {

			return err
		}

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
