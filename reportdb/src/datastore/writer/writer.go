package writer

import (
	"log"
	. "reportdb/config"
	. "reportdb/src/storage"
	. "reportdb/src/utils"
	"strconv"
	"sync"
	"time"
)

type WriterPool struct {
	writers []*Writer

	pollCh <-chan []RowData

	writerChannels []chan RowData
}

type Writer struct {
	ID uint8

	TaskQueue <-chan RowData

	storePool *StorageEnginePool

	wg *sync.WaitGroup
}

func NewWriterPool(ch <-chan []RowData, writerCount uint8) *WriterPool {

	return &WriterPool{

		writers: make([]*Writer, writerCount),

		pollCh: ch,

		writerChannels: make([]chan RowData, writerCount),
	}
}

func (wp *WriterPool) StartWriter(writerCount uint8, baseDir string, wg *sync.WaitGroup, sp *StorageEnginePool) {

	for i := uint8(0); i < writerCount; i++ {

		wp.writerChannels[i] = make(chan RowData, 50)

		wp.writers[i] = &Writer{

			ID: i,

			TaskQueue: wp.writerChannels[i],

			storePool: sp,

			wg: wg,
		}

		wg.Add(1)

		go wp.writers[i].runWorker(baseDir)
	}

	go func(writerCount uint8) {

		for batch := range wp.pollCh {

			for _, row := range batch {

				writerIdx := uint8((uint32(row.CounterId) + row.ObjectId) % uint32(writerCount))

				wp.writerChannels[writerIdx] <- row
			}
		}

		for _, ch := range wp.writerChannels {

			close(ch)
		}
	}(writerCount)
}

func (w *Writer) runWorker(baseDir string) {

	defer w.wg.Done()

	for row := range w.TaskQueue {

		day := time.Unix(int64(row.Timestamp), 0).Truncate(24 * time.Hour).UTC()

		path := baseDir + day.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(row.CounterId))

		store := w.storePool.GetEngine(path)

		partition := GetPartition(row.ObjectId, store.PartitionCount)

		handle, err := store.FileCfg.GetHandle(partition)

		if err != nil {

			log.Printf("failed to get file handle: %w", err)
		}

		data, err := EncodeData(row)

		if err != nil {

			log.Printf("failed to encode data: %w", err)
		}

		if err := store.FileCfg.EnsureCapacity(handle, int64(len(data))); err != nil {

			log.Printf("failed to ensure capacity: %w", err)
		}

		offset, err := store.Put(handle, data)

		if err != nil {

			log.Printf("failed to write data: %w", err)
		}

		store.IndexCfg.Update(row.ObjectId, offset, row.Timestamp)

		store.IndexCfg.Save()
	}
}
