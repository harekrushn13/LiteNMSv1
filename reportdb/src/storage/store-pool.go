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

	shutdown chan bool
}

func NewStorePool() *StorePool {

	return &StorePool{

		storePool: make(map[string]*StoreEngine),

		poolMutex: &sync.RWMutex{},

		shutdown: make(chan bool, 1),
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

func (storePool *StorePool) SaveEngine() error {

	saveIndexInterval, err := GetSaveIndexInterval()

	if err != nil {

		return fmt.Errorf("storePool.SaveEngine error : %v", err.Error())
	}

	ticker := time.NewTicker(time.Duration(saveIndexInterval) * time.Second)

	go func(storePool *StorePool, ticker *time.Ticker) {

		for {

			select {

			case <-storePool.shutdown:

				ticker.Stop()

				<-storePool.shutdown

				return

			case <-ticker.C:

				storePool.flushAllEngines()
			}
		}

	}(storePool, ticker)

	return nil
}

func (storePool *StorePool) flushAllEngines() {

	currentTime := time.Now().Unix()

	storePool.poolMutex.RLock()

	defer storePool.poolMutex.RUnlock()

	for _, engine := range storePool.storePool {

		if engine.isUsedPut == true && currentTime-engine.lastSave >= 2 {

			engine.lastSave = currentTime

			err := engine.indexManager.Save()

			if err != nil {

				log.Printf("storePool.SaveEngine error: %s\n", err)
			}
		}
	}
}

func (storePool *StorePool) Shutdown() {

	storePool.shutdown <- true

	storePool.poolMutex.Lock()

	currentTime := time.Now().Unix()

	for _, engine := range storePool.storePool {

		if engine.isUsedPut {

			engine.lastSave = currentTime

			if err := engine.indexManager.Save(); err != nil {

				log.Printf("storePool.SaveEngine error: %s\n", err)
			}
		}
	}

	storePool.poolMutex.Unlock()
}
