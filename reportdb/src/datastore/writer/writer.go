package writer

import (
	"fmt"
	"log"
	. "reportdb/storage"
	. "reportdb/utils"
	"sync"
)

type Writer struct {
	id uint8

	events chan Events

	storePool *StorePool

	waitGroup *sync.WaitGroup
}

func initializeWriters(storePool *StorePool) ([]*Writer, error) {

	numWriters, err := GetWriters()

	if err != nil {

		return nil, fmt.Errorf("initializeWriters : Error getting writers: %v", err)
	}

	eventsBuffer, err := GetWriterEventBuffer()

	if err != nil {

		return nil, fmt.Errorf("initializeWriters : Error getting writers: %v", err)
	}

	writers := make([]*Writer, numWriters)

	for i := range writers {

		writers[i] = &Writer{

			id: uint8(i),

			events: make(chan Events, eventsBuffer),

			storePool: storePool,

			waitGroup: &sync.WaitGroup{},
		}
	}

	return writers, nil
}

func StartWriter(storePool *StorePool) ([]*Writer, error) {

	writers, err := initializeWriters(storePool)

	if err != nil {

		return nil, fmt.Errorf("StartWriter : Error getting writers: %v", err)
	}

	workingDirectory, err := GetWorkingDirectory()

	if err != nil {

		return nil, fmt.Errorf("writer.runWriter : Error getting working directory: %v", err)

	}

	for _, writer := range writers {

		writer.runWriter(workingDirectory)
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

				log.Printf("writer.runWriter : Error getting store: %v", err)

				continue
			}

			data, err := encodeData(row)

			if err != nil {

				log.Printf("writer.runWriter : failed to encode data: %s", err)

				continue
			}

			err = store.Put(row.ObjectId, row.Timestamp, data)

			if err != nil {

				log.Printf("writer.runWriter : failed to write data: %s", err)

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
