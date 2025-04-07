package main

import (
	. "reportdb/config"
	. "reportdb/src/datastore/reader"
	. "reportdb/src/datastore/writer"
	. "reportdb/src/polling"
	. "reportdb/src/storage"
	"sync"
)

func main() {

	var wg sync.WaitGroup

	globalCfg := NewGlobalConfig()

	pollerCfg := NewPollerEngine()

	pollCh := pollerCfg.PollData(globalCfg)

	storagePool := NewStorageEnginePool()

	go storagePool.SaveEngine(&wg)

	writePool := NewWriterPool(pollCh, globalCfg.WriterCount)

	writePool.StartWriter(globalCfg.WriterCount, globalCfg.BaseDir, &wg, storagePool)

	readerPool := NewReaderPool(1)

	readerPool.StartReader(1, globalCfg.BaseDir, &wg, storagePool)

	wg.Wait()
}
