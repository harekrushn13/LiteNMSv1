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

	lock *sync.RWMutex

	shutdown chan bool
}

func NewStorePool() *StorePool {

	return &StorePool{

		storePool: make(map[string]*StoreEngine),

		lock: &sync.RWMutex{},

		shutdown: make(chan bool, 1),
	}
}

func (storePool *StorePool) GetEngine(path string, isForPut bool) (*StoreEngine, error) {

	// Reading From storePool

	storePool.lock.RLock()

	if engine, exists := storePool.storePool[path]; exists {

		storePool.lock.RUnlock()

		return engine, nil
	}

	storePool.lock.RUnlock()

	// Updating Into storePool

	storePool.lock.Lock()

	defer storePool.lock.Unlock()

	if engine, exists := storePool.storePool[path]; exists {

		return engine, nil
	}

	if !isForPut && !storePool.engineAvailable(path) {

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

	ticker := time.NewTicker(time.Duration(GetSaveIndexInterval()) * time.Second)

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

	storePool.lock.RLock()

	defer storePool.lock.RUnlock()

	for _, engine := range storePool.storePool {

		if engine.isUsedPut == true && currentTime-engine.lastSave >= int64(GetSaveIndexInterval()) {

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

	storePool.lock.Lock()

	currentTime := time.Now().Unix()

	for _, engine := range storePool.storePool {

		if engine.isUsedPut {

			engine.lastSave = currentTime

			if err := engine.indexManager.Save(); err != nil {

				log.Printf("storePool.SaveEngine error: %s\n", err)
			}
		}

		engine.fileManager.Close()
	}

	storePool.lock.Unlock()
}

func (storePool *StorePool) CheckEngineUsedPut(day string) bool {

	storePool.lock.RLock()

	defer storePool.lock.RUnlock()

	if engine, exists := storePool.storePool[day]; exists && engine.isUsedPut == true {

		return true
	}

	return false
}
