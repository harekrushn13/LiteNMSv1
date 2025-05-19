package writer

import (
	"fmt"
	"go.uber.org/zap"
	. "reportdb/logger"
	. "reportdb/storage"
	. "reportdb/utils"
	"sync"
)

type Writer struct {
	id uint8

	events chan Events

	storePool *StorePool

	waitGroup *sync.WaitGroup

	data []byte // for serializing data
}

func StartWriter(storePool *StorePool) ([]*Writer, error) {

	writers, err := initializeWriters(storePool)

	if err != nil {

		return nil, fmt.Errorf("StartWriter : Error getting writers: %v", err)
	}

	for _, writer := range writers {

		writer.runWriter(GetWorkingDirectory())
	}

	return writers, nil
}

func initializeWriters(storePool *StorePool) ([]*Writer, error) {

	writers := make([]*Writer, GetWriters())

	for i := range writers {

		writers[i] = &Writer{

			id: uint8(i),

			events: make(chan Events, GetEventsBuffer()),

			storePool: storePool,

			waitGroup: &sync.WaitGroup{},

			data: make([]byte, 100),
		}
	}

	return writers, nil
}

func (writer *Writer) runWriter(workingDirectory string) {

	writer.waitGroup.Add(1)

	go func(writer *Writer) {

		defer writer.waitGroup.Done()

		for row := range writer.events {

			store, err := writer.storePool.GetEngine(getPath(workingDirectory, row), true)

			if err != nil {

				Logger.Error("Writer: error getting store",
					zap.Uint8("writer_id", writer.id),
					zap.Uint32("object_id", row.ObjectId),
					zap.Uint16("counter_id", row.CounterId),
					zap.Error(err),
				)

				continue
			}

			lastIndex, err := encodeData(row, &writer.data)

			if err != nil {

				Logger.Error("Writer: failed to encode data",
					zap.Uint8("writer_id", writer.id),
					zap.Uint32("object_id", row.ObjectId),
					zap.Uint16("counter_id", row.CounterId),
					zap.Error(err),
				)

				continue
			}

			err = store.Put(row.ObjectId, writer.data[:lastIndex])

			if err != nil {

				Logger.Error("Writer: failed to write data",
					zap.Uint8("writer_id", writer.id),
					zap.Uint32("object_id", row.ObjectId),
					zap.Uint16("counter_id", row.CounterId),
					zap.Uint8("data_bytes", lastIndex),
					zap.Error(err),
				)

				continue
			}

		}

		return

	}(writer)
}

func ShutdownWriters(writers []*Writer) {

	for _, writer := range writers {

		close(writer.events)

		writer.waitGroup.Wait()
	}
}
