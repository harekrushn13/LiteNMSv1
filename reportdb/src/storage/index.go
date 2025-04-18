package storage

import (
	"encoding/json"
	"fmt"
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
	BlockStart int64 `json:"blockStart"`

	BlockEnd int64 `json:"blockEnd"`

	EntryStart int64 `json:"entryStart"`

	EntryEnd int64 `json:"entryEnd"`
}

func NewIndexManager(baseDir string) *IndexManager {

	return &IndexManager{

		indexHandles: make(map[uint8]map[uint32][]*IndexEntry),

		lock: &sync.RWMutex{},

		baseDir: baseDir,
	}
}

func (indexManager *IndexManager) GetIndexMapEntryList(key uint32, indexId uint8, isUsedPut bool) ([]*IndexEntry, error) {

	indexManager.lock.RLock()

	indexMap, exists := indexManager.indexHandles[indexId]

	if exists {

		if entryList, exists := indexMap[key]; exists {

			indexManager.lock.RUnlock()

			return entryList, nil
		}
	}

	indexManager.lock.RUnlock()

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

		if isUsedPut {

			if err := os.MkdirAll(filepath.Dir(indexFilePath), 0755); err != nil {

				return nil, fmt.Errorf("error creating index directory: %v", err)
			}
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

func (indexManager *IndexManager) Update(key uint32, indexId uint8, entryList []*IndexEntry) {

	indexManager.lock.Lock()

	defer indexManager.lock.Unlock()

	indexManager.indexHandles[indexId][key] = entryList

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
