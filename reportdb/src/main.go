package main

import (
	"log"
	"reportdb/config"
	"reportdb/src/datastore/reader"
	"reportdb/src/datastore/writer"
	"reportdb/src/polling"
	"reportdb/src/storage/helper"
	"sync"
	"time"
)

func main() {

	var wg sync.WaitGroup

	globalCfg := config.NewGlobalConfig()

	pollerCfg := polling.NewPollerEngine()

	pollCh := pollerCfg.PollData(globalCfg)

	fileCfg := helper.NewFileManager(globalCfg.BaseDir)

	indexCfg := helper.NewIndexManager(globalCfg.BaseDir)

	writePool := writer.NewWriterPool(pollCh, globalCfg.WriterCount)

	writePool.StartWriter(globalCfg.WriterCount, fileCfg, indexCfg, globalCfg.BaseDir, &wg)

	// save Index at specific Interval

	wg.Add(1)

	go func(indexCfg *helper.IndexManager) {

		defer wg.Done()

		t := time.NewTicker(2 * time.Second)

		defer t.Stop()

		stopTimer := time.NewTimer(12 * time.Second)

		defer stopTimer.Stop()

		for {

			select {

			case <-t.C:

				if err := indexCfg.Save(time.Now()); err != nil {

					log.Fatal(err)
				}

			case <-stopTimer.C:

				if err := indexCfg.Save(time.Now()); err != nil {

					log.Fatal(err)
				}

				return
			}
		}
	}(indexCfg)

	// parallel reader

	reader := reader.NewReader(&wg, globalCfg.BaseDir)

	reader.StartReader(globalCfg.ReaderCount, fileCfg, indexCfg)

	wg.Wait()
}
