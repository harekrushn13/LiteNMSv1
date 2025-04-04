package writer

import (
	"log"
	"reportdb/config"
	"reportdb/src/storage/engine"
	"reportdb/src/storage/helper"
	"sync"
)

type WriterPool struct {
	workers []*Writer

	pollCh <-chan []config.RowData

	workerChannels []chan config.RowData
}

type Writer struct {
	ID uint8

	TaskQueue <-chan config.RowData

	store *engine.StorageEngine

	wg *sync.WaitGroup
}

func NewWriterPool(ch <-chan []config.RowData, writerCount uint8) *WriterPool {

	return &WriterPool{

		workers: make([]*Writer, writerCount),

		pollCh: ch,

		workerChannels: make([]chan config.RowData, writerCount),
	}
}

func (wp *WriterPool) StartWriter(writerCount uint8, fileCfg *helper.FileManager, indexCfg *helper.IndexManager, baseDir string, wg *sync.WaitGroup) {

	for i := uint8(0); i < writerCount; i++ {

		wp.workerChannels[i] = make(chan config.RowData, 50)

		wp.workers[i] = &Writer{

			ID: i,

			TaskQueue: wp.workerChannels[i],

			store: engine.NewStorageEngine(fileCfg, indexCfg, baseDir),

			wg: wg,
		}

		wg.Add(1)

		go wp.workers[i].runWorker()
	}

	go func(writerCount uint8) {

		for batch := range wp.pollCh {

			for _, row := range batch {

				writerIdx := uint8((uint32(row.CounterId) + row.ObjectId) % uint32(writerCount))

				wp.workerChannels[writerIdx] <- row
			}
		}

		for _, ch := range wp.workerChannels {

			close(ch)
		}
	}(writerCount)
}

func (w *Writer) runWorker() {

	defer w.wg.Done()

	for row := range w.TaskQueue {

		if err := w.store.Save(row); err != nil {

			log.Printf("Error writing row: %v", err)
		}
	}
}
