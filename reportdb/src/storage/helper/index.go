package helper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type IndexEntry struct {
	StartTime uint32 `json:"starttime"`

	EndTime uint32 `json:"endtime"`

	Offsets []int64 `json:"offsets"`
}

type IndexManager struct {
	indexMap map[uint32]*IndexEntry // IndexMap[objectID]

	mu *sync.RWMutex

	BaseDir string // ./src/storage/database/YYYY/MM/DD/counter_1
}

func NewIndexManager(baseDir string) *IndexManager {

	return &IndexManager{

		indexMap: make(map[uint32]*IndexEntry),

		mu: &sync.RWMutex{},

		BaseDir: baseDir,
	}
}

func (im *IndexManager) Update(objectID uint32, offset int64, timestamp uint32) {

	im.mu.Lock()

	defer im.mu.Unlock()

	entry, exists := im.indexMap[objectID]

	if exists {

		entry.EndTime = timestamp

		entry.Offsets = append(entry.Offsets, offset)

	} else {

		im.indexMap[objectID] = &IndexEntry{

			StartTime: timestamp,

			EndTime: timestamp,

			Offsets: []int64{offset},
		}
	}
}

func (im *IndexManager) Save() error {

	im.mu.Lock()

	defer im.mu.Unlock()

	indexFilePath := im.BaseDir + "/index.json"

	if err := os.MkdirAll(filepath.Dir(indexFilePath), 0755); err != nil {

		return err
	}

	data, err := json.MarshalIndent(im.indexMap, "", "  ")

	if err != nil {

		return err
	}

	if err := os.WriteFile(indexFilePath, data, 0644); err != nil {

		return err
	}

	return nil
}

func (im *IndexManager) GetIndexMap(objectID uint32) (*IndexEntry, error) {

	im.mu.RLock()

	defer im.mu.RUnlock()

	if len(im.indexMap) == 0 {

		indexFilePath := im.BaseDir + "/index.json"

		if _, err := os.Stat(indexFilePath); err == nil {

			data, err := os.ReadFile(indexFilePath)

			if err != nil {

				return nil, fmt.Errorf("failed to read index file: %v", err)
			}

			if err := json.Unmarshal(data, &im.indexMap); err != nil {

				return nil, fmt.Errorf("failed to parse index file: %v", err)
			}
		}
	}

	entry, exists := im.indexMap[objectID]

	if !exists {

		return nil, fmt.Errorf("no index map found for objectID %d", objectID)
	}

	return entry, nil
}

func (im *IndexManager) GetValidOffsets(entry *IndexEntry, from uint32, to uint32) ([]int64, error) {

	var validOffsets []int64

	length := uint8(len(entry.Offsets))

	var start uint8

	var end uint8

	if from <= entry.StartTime {

		start = uint8(0)

	} else {

		start = uint8(from - entry.StartTime)
	}

	if to >= entry.EndTime {

		end = length

	} else {

		end = length - uint8(entry.EndTime-to)
	}

	validOffsets = append(validOffsets, entry.Offsets[start:end]...)

	return validOffsets, nil
}
