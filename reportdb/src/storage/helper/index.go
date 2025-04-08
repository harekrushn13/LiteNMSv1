package helper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type IndexEntry struct {
	TimeStamp uint32 `json:"timestamp"`

	Offset int64 `json:"offset"`
}

type IndexManager struct {
	indexHandles map[uint8]*IndexHandle // map[index]

	handleMux *sync.RWMutex

	BaseDir string // ./src/storage/database/YYYY/MM/DD/counter_1

	PartitionCount uint8
}

type IndexHandle struct {
	indexMap map[uint32][]*IndexEntry // IndexMap[objectID][]*IndexEntry

	mapMux *sync.RWMutex
}

func NewIndexManager(baseDir string, partitionCount uint8) *IndexManager {

	return &IndexManager{

		indexHandles: make(map[uint8]*IndexHandle),

		handleMux: &sync.RWMutex{},

		BaseDir: baseDir,

		PartitionCount: partitionCount,
	}
}

func (im *IndexManager) Update(objectID uint32, offset int64, timestamp uint32, indexId uint8) {

	indexHandle := im.getIndexHandle(indexId)

	indexHandle.mapMux.Lock()

	defer indexHandle.mapMux.Unlock()

	entryList, exists := indexHandle.indexMap[objectID]

	if !exists {

		indexHandle.indexMap[objectID] = []*IndexEntry{{TimeStamp: timestamp, Offset: offset}}

	} else {

		indexHandle.indexMap[objectID] = append(entryList, &IndexEntry{TimeStamp: timestamp, Offset: offset})
	}

}

func (im *IndexManager) getIndexHandle(indexId uint8) *IndexHandle {

	im.handleMux.RLock()

	handle, exists := im.indexHandles[indexId]

	if exists {

		im.handleMux.RUnlock()

		return handle
	}

	im.handleMux.RUnlock()

	im.handleMux.Lock()

	defer im.handleMux.Unlock()

	im.indexHandles[indexId] = &IndexHandle{

		indexMap: make(map[uint32][]*IndexEntry),

		mapMux: &sync.RWMutex{},
	}

	return im.indexHandles[indexId]
}

func (im *IndexManager) GetIndexMapEntryList(objectId uint32, indexId uint8) ([]*IndexEntry, error) {

	indexHandle := im.getIndexHandle(indexId)

	if indexHandle == nil {

		return nil, fmt.Errorf("get index map error")
	}

	indexHandle.mapMux.Lock()

	defer indexHandle.mapMux.Unlock()

	if len(indexHandle.indexMap[objectId]) == 0 {

		indexFilePath := im.BaseDir + "/index_" + strconv.Itoa(int(indexId)) + ".json"

		if _, err := os.Stat(indexFilePath); err == nil {

			data, err := os.ReadFile(indexFilePath)

			if err != nil {

				return nil, fmt.Errorf("failed to read index file: %v", err)
			}

			if err := json.Unmarshal(data, &indexHandle.indexMap); err != nil {

				return nil, fmt.Errorf("failed to parse index file: %v", err)
			}
		}
	}

	entryList, exists := indexHandle.indexMap[objectId]

	if !exists {

		return nil, fmt.Errorf("get index map error")
	}

	return entryList, nil
}

func (im *IndexManager) GetValidOffsets(entryList []*IndexEntry, from uint32, to uint32) ([]int64, error) {

	if len(entryList) == 0 {

		return nil, nil
	}

	start := customBinarySearch(entryList, func(ts uint32) bool {

		return ts >= from

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

		ts := entries[mid].TimeStamp

		if condition(ts) {

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

func (im *IndexManager) Save() error {

	im.handleMux.Lock()

	defer im.handleMux.Unlock()

	for index, handle := range im.indexHandles {

		handle.mapMux.Lock()

		indexFilePath := im.BaseDir + "/index_" + strconv.Itoa(int(index)) + ".json"

		if err := os.MkdirAll(filepath.Dir(indexFilePath), 0755); err != nil {

			return err
		}

		data, err := json.MarshalIndent(handle.indexMap, "", "  ")

		if err != nil {

			return err
		}

		if err := os.WriteFile(indexFilePath, data, 0644); err != nil {

			return err
		}

		handle.mapMux.Unlock()
	}

	return nil
}
