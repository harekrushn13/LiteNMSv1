package storage

import (
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"os"
	"path/filepath"
	. "reportdb/utils"
	"strconv"
	"sync"
)

type IndexManager struct {
	indexHandles map[uint8]map[uint32][]*IndexEntry // indexHandles[indexId][key][]*IndexEntry

	lock *sync.RWMutex

	baseDir string // ./database/YYYY/MM/DD/counter_1
}

type IndexEntry struct {
	BlockStart int64 `msgpack:"blockStart"`

	BlockEnd int64 `msgpack:"blockEnd"`

	EntryStart int64 `msgpack:"entryStart"`

	EntryEnd int64 `msgpack:"entryEnd"`
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

		indexFilePath := indexManager.baseDir + "/index_" + strconv.Itoa(int(indexId)) + ".msg"

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

	if err := msgpack.Unmarshal(data, indexMap); err != nil {

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

		indexFilePath := indexManager.baseDir + "/index_" + strconv.Itoa(int(index)) + ".msg"

		if err := os.MkdirAll(filepath.Dir(indexFilePath), 0755); err != nil {

			return err
		}

		data, err := msgpack.Marshal(indexMap)

		if err != nil {

			return err
		}

		if err := os.WriteFile(indexFilePath, data, 0644); err != nil {

			return err
		}

	}

	return nil
}

func (indexManager *IndexManager) GetAllKeys() ([]uint32, error) {

	var allKeys []uint32

	indexManager.lock.Lock()

	defer indexManager.lock.Unlock()

	if len(indexManager.indexHandles) == 0 {

		indexManager.indexHandles = make(map[uint8]map[uint32][]*IndexEntry)

		partitions := GetPartitions()

		for i := 0; i < partitions; i++ {

			indexManager.indexHandles[uint8(i)] = make(map[uint32][]*IndexEntry)
		}
	}

	for indexId, indexMap := range indexManager.indexHandles {

		if len(indexMap) == 0 {

			indexFilePath := indexManager.baseDir + "/index_" + strconv.Itoa(int(indexId)) + ".msg"

			if err := loadIndexFile(indexFilePath, &indexMap); err != nil {

				return nil, fmt.Errorf("error loading index file for indexId %d: %v", indexId, err)
			}

			indexManager.indexHandles[indexId] = indexMap
		}

		for key := range indexMap {

			allKeys = append(allKeys, key)
		}
	}

	return allKeys, nil
}

func (indexManager *IndexManager) Close() {

	indexManager.lock.Lock()

	defer indexManager.lock.Unlock()

	for indexId := range indexManager.indexHandles {

		for key := range indexManager.indexHandles[indexId] {

			indexManager.indexHandles[indexId][key] = nil

			delete(indexManager.indexHandles[indexId], key)

		}

		delete(indexManager.indexHandles, indexId)
	}
}
