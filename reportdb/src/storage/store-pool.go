package storage

import (
	"fmt"
	"log"
	"os"
	. "reportdb/utils"
	"sync"
	"time"
)

type StorePool struct {
	storePool map[string]*StoreEngine // map["./database/YYYY/MM/DD/counter_1"]

	poolMutex *sync.RWMutex
}

func NewStorePool() *StorePool {

	return &StorePool{

		storePool: make(map[string]*StoreEngine),

		poolMutex: &sync.RWMutex{},
	}
}

func (storePool *StorePool) GetEngine(path string, isForPut bool) (*StoreEngine, error) {

	// Reading From storePool

	storePool.poolMutex.RLock()

	if engine, exists := storePool.storePool[path]; exists {

		storePool.poolMutex.RUnlock()

		return engine, nil
	}

	storePool.poolMutex.RUnlock()

	// Updating Into storePool

	storePool.poolMutex.Lock()

	defer storePool.poolMutex.Unlock()

	if engine, exists := storePool.storePool[path]; exists {

		return engine, nil
	}

	isEngineAvailable := storePool.engineAvailable(path)

	if !isForPut && !isEngineAvailable {

		return nil, fmt.Errorf("engine %s is not available", path)
	}

	engine := NewStorageEngine(path)

	storePool.storePool[path] = engine

	return engine, nil
}

func (storePool *StorePool) engineAvailable(path string) bool {

	if info, err := os.Stat(path); err != nil || !info.IsDir() {

		return false
	}

	return true
}

func (storePool *StorePool) SaveEngine(waitGroup *sync.WaitGroup) error {

	waitGroup.Add(1)

	go func() {

		defer waitGroup.Done()

		saveIndexInterval, err := GetSaveIndexInterval()

		if err != nil {

			log.Fatal("storePool.SaveEngine error : ", err.Error())
		}

		stopIndexInterval, err := GetStopIndexSaving()

		if err != nil {

			log.Fatal("storePool.SaveEngine error : ", err.Error())
		}

		ticker := time.NewTicker(time.Duration(saveIndexInterval) * time.Second)

		defer ticker.Stop()

		stopTime := time.NewTicker(time.Duration(stopIndexInterval) * time.Second)

		defer stopTime.Stop()

		for {

			select {

			case <-ticker.C:

				currentTime := time.Now().Unix()

				for _, engine := range storePool.storePool {

					if engine.isUsedPut == true && currentTime-engine.lastSave >= 2 {

						engine.lastSave = currentTime

						err := engine.indexManager.Save()

						if err != nil {

							log.Printf("storePool.SaveEngine error: %s\n", err)
						}
					}
				}

			case <-stopTime.C:

				for _, engine := range storePool.storePool {

					if engine.isUsedPut == true {

						engine.lastSave = time.Now().Unix()

						err := engine.indexManager.Save()

						if err != nil {

							log.Printf("storePool.SaveEngine error: %s\n", err)
						}
					}
				}

				return
			}
		}

	}()

	return nil
}
