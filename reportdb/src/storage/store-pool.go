package storage

import (
	"log"
	"sync"
	"time"
)

type StorePool struct {
	storePool map[string]*StoreEngine // map["./database/YYYY/MM/DD/counter_1"]

	poolMutex sync.RWMutex
}

func NewStorePool() *StorePool {

	return &StorePool{

		storePool: make(map[string]*StoreEngine),
	}
}

func (storePool *StorePool) GetEngine(path string) *StoreEngine {

	storePool.poolMutex.RLock()

	if engine, exists := storePool.storePool[path]; exists {

		storePool.poolMutex.RUnlock()

		return engine
	}

	storePool.poolMutex.RUnlock()

	storePool.poolMutex.Lock()

	defer storePool.poolMutex.Unlock()

	engine := NewStorageEngine(path)

	storePool.storePool[path] = engine

	return engine
}

func (storePool *StorePool) SaveEngine(waitGroup *sync.WaitGroup) {

	waitGroup.Add(1)

	go func() {

		defer waitGroup.Done()

		ticker := time.NewTicker(2 * time.Millisecond)

		defer ticker.Stop()

		stopTime := time.NewTicker(60 * time.Second)

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

}
