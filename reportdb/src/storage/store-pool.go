package storage

import (
	"sync"
	"time"
)

type StorageEnginePool struct {
	pool map[string]*StorageEngine // map["./src/storage/database/YYYY/MM/DD/counter_1"]

	mu sync.RWMutex
}

func NewStorageEnginePool() *StorageEnginePool {

	return &StorageEnginePool{

		pool: make(map[string]*StorageEngine),
	}
}

func (p *StorageEnginePool) GetEngine(baseDir string) *StorageEngine {

	p.mu.RLock()

	if engine, exists := p.pool[baseDir]; exists {

		p.mu.RUnlock()

		return engine
	}

	p.mu.RUnlock()

	p.mu.Lock()

	defer p.mu.Unlock()

	engine := NewStorageEngine(baseDir)

	p.pool[baseDir] = engine

	return engine
}

func (p *StorageEnginePool) SaveEngine(wg *sync.WaitGroup) {

	defer wg.Done()

	ticker := time.NewTicker(2500 * time.Millisecond)

	defer ticker.Stop()

	stopTime := time.NewTicker(12 * time.Second)

	defer stopTime.Stop()

	for {
		select {

		case <-ticker.C:

			currentTime := time.Now().Unix()

			for _, engine := range p.pool {

				if engine.isUsedPut && engine.lastAccess > engine.lastSave && currentTime-engine.lastSave >= 2 {

					engine.indexCfg.Save()
				}
			}

		case <-stopTime.C:

			for _, engine := range p.pool {

				if engine.isUsedPut {

					engine.indexCfg.Save()
				}
			}

			return
		}
	}
}
