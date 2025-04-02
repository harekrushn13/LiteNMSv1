package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reportdb/config"
	"sync"
	"time"
)

func InitIndexHandlesFunc() {

	for counterID := uint16(1); counterID <= config.CounterCount; counterID++ {

		config.IndexHandles[counterID] = &config.IndexHandle{

			Lock: &sync.RWMutex{},
		}
	}
}

func UpdateIndex(counterID uint16, objectID uint32, offset int64, timestamp uint32) {

	//config.InitIndexHandles.Do(InitIndexHandlesFunc)
	//
	//currentDay := time.Now()
	//
	//if currentDay.Day() != config.LastDayChecked.Day() {
	//
	//	err := RotateDailyIndex(config.LastDayChecked)
	//
	//	if err != nil {
	//
	//		log.Printf("Error rotating index: %v", err)
	//
	//	} else {
	//
	//		config.LastDayChecked = currentDay
	//	}
	//}

	config.Mu.Lock()

	if _, exists := config.IndexMap[counterID]; !exists {

		config.IndexMap[counterID] = make(map[uint32]*config.IndexEntry)
	}

	entry, exists := config.IndexMap[counterID][objectID]

	if exists {

		entry.EndTime = timestamp

		entry.Offsets = append(entry.Offsets, offset)

	} else {

		config.IndexMap[counterID][objectID] = &config.IndexEntry{

			StartTime: timestamp,

			EndTime: timestamp,

			Offsets: []int64{offset},
		}
	}

	config.Mu.Unlock()
}

func SaveIndex(day time.Time) error {

	for counterID, counterIndex := range config.IndexMap {

		indexFilePath := GetIndexFilePath(counterID, day)

		counterDir := filepath.Dir(indexFilePath)

		data, err := json.MarshalIndent(counterIndex, "", "  ")

		if err != nil {

			return err
		}

		if err := os.MkdirAll(counterDir, 0755); err != nil {

			return err
		}

		handle := config.IndexHandles[counterID]

		handle.Lock.Lock()

		if err := os.WriteFile(indexFilePath, data, 0644); err != nil {

			handle.Lock.Unlock()

			return err
		}

		handle.Lock.Unlock()
	}

	return nil
}

func RotateDailyIndex(day time.Time) error {

	//if err := SaveIndex(day); err != nil {
	//	return err
	//}

	//for counterID := range config.IndexMap {

	//config.IndexMap[counterID] = make(map[uint32]*config.IndexEntry)

	//config.IndexHandles[counterID] = &config.IndexHandle{
	//
	//	Lock: &sync.RWMutex{},
	//}
	//}

	//InitIndexHandlesFunc()
	//
	//InitFiles()

	//config.IndexMap = make(map[uint16]map[uint32]*config.IndexEntry)

	//config.IndexHandles = make(map[uint16]*config.IndexHandle)

	return nil
}
