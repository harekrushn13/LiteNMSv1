package helper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type IndexEntry struct {
	StartTime uint32 `json:"starttime"`

	EndTime uint32 `json:"endtime"`

	Offsets []int64 `json:"offsets"`
}

type IndexHandle struct {
	Lock *sync.RWMutex
}

type IndexManager struct {
	indexMap map[uint16]map[uint32]*IndexEntry // IndexMap[counterID][objectID]

	indexHandles map[uint16]*IndexHandle // IndexHandles[counterID]

	mu sync.RWMutex

	BaseDir string
}

func NewIndexManager(baseDir string) *IndexManager {

	return &IndexManager{

		indexMap: make(map[uint16]map[uint32]*IndexEntry),

		indexHandles: make(map[uint16]*IndexHandle),

		mu: sync.RWMutex{},

		BaseDir: baseDir,
	}
}

func (im *IndexManager) Update(counterID uint16, objectID uint32, offset int64, timestamp uint32) {

	im.mu.Lock()

	defer im.mu.Unlock()

	if _, exists := im.indexMap[counterID]; !exists {

		im.indexMap[counterID] = make(map[uint32]*IndexEntry)

		im.indexHandles[counterID] = &IndexHandle{

			Lock: &sync.RWMutex{},
		}
	}

	entry, exists := im.indexMap[counterID][objectID]

	if exists {

		entry.EndTime = timestamp

		entry.Offsets = append(entry.Offsets, offset)

	} else {

		im.indexMap[counterID][objectID] = &IndexEntry{

			StartTime: timestamp,

			EndTime: timestamp,

			Offsets: []int64{offset},
		}
	}
}

func (im *IndexManager) Save(day time.Time) error {

	im.mu.Lock()

	defer im.mu.Unlock()

	for counterID, counterIndex := range im.indexMap {

		indexFilePath := im.GetIndexFilePath(counterID, day)

		if err := os.MkdirAll(filepath.Dir(indexFilePath), 0755); err != nil {

			return err
		}

		data, err := json.MarshalIndent(counterIndex, "", "  ")

		if err != nil {

			return err
		}

		handle := im.indexHandles[counterID]

		handle.Lock.Lock()

		if err := os.WriteFile(indexFilePath, data, 0644); err != nil {

			handle.Lock.Unlock()

			return err
		}

		handle.Lock.Unlock()
	}

	return nil
}

func (im *IndexManager) GetIndexMap(counterID uint16, objectID uint32) (*IndexEntry, error) {

	im.mu.RLock()

	defer im.mu.RUnlock()

	data := im.indexMap[counterID][objectID]

	if data == nil {

		return nil, fmt.Errorf("no index map found for counter %d", counterID)
	}

	return data, nil
}

func (im *IndexManager) GetIndexHandle(counterID uint16) (*IndexHandle, error) {

	if handle, exits := im.indexHandles[counterID]; exits {

		return handle, nil
	}

	return nil, fmt.Errorf("no index handle for counter %d", counterID)
}

func (im *IndexManager) GetIndexFilePath(counterID uint16, t time.Time) string {

	year := strconv.Itoa(t.Year())

	month := fmt.Sprintf("%02d", t.Month())

	day := fmt.Sprintf("%02d", t.Day())

	datePath := filepath.Join(im.BaseDir, year, month, day)

	counterDir := filepath.Join(datePath, "counter_"+strconv.Itoa(int(counterID)))

	return filepath.Join(counterDir, "index.json")
}
