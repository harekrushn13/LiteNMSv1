package storage

import (
	. "reportdb/config"
	. "reportdb/src/storage/helper"
	. "reportdb/src/utils"
	"sync"
)

type StorageEnginePool struct {
	pool map[string]*StorageEngine // map["./src/storage/database/YYYY/MM/DD/counter_1"]

	mu sync.RWMutex
}

type StorageEngine struct {
	FileCfg *FileManager

	IndexCfg *IndexManager

	PartitionCount uint8

	BaseDir string // ex. ./src/storage/database/YYYY/MM/DD/counter_1
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

func NewStorageEngine(baseDir string) *StorageEngine {

	return &StorageEngine{
		FileCfg: NewFileManager(baseDir),

		IndexCfg: NewIndexManager(baseDir),

		BaseDir: baseDir,

		PartitionCount: 3,
	}
}

func (store *StorageEngine) Put(handle *FileHandle, data []byte) (int64, error) {

	handle.Lock.Lock()

	defer handle.Lock.Unlock()

	offset := handle.Offset

	copy(handle.MmapData[offset:], data)

	handle.Offset += int64(len(data))

	return offset, nil
}

func (store *StorageEngine) Get(data []byte, offset int64, dataType DataType) ([]byte, error) {

	start, end, err := LengthOfData(data, offset, dataType)

	if err != nil {

		return nil, err
	}

	value := data[start:end]

	return value, nil
}
