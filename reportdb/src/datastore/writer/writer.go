package writer

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	. "reportdb/storage"
	. "reportdb/utils"
	"strconv"
	"sync"
	"time"
)

type Writer struct {
	id uint8

	taskQueue chan RowData

	storePool *StorePool

	waitGroup *sync.WaitGroup
}

func StartWriter(pollChannel <-chan []RowData, waitGroup *sync.WaitGroup, storePool *StorePool) {

	writers := make([]*Writer, GetWriters())

	for i := range writers {

		writers[i] = &Writer{

			id: uint8(i),

			taskQueue: make(chan RowData, GetWriterTaskQueueBuffer()),

			storePool: storePool,

			waitGroup: waitGroup,
		}
	}

	for _, writer := range writers {

		writer.runWriter()
	}

	waitGroup.Add(1)

	go func(writers []*Writer) {

		defer waitGroup.Done()

		for batch := range pollChannel {

			for _, row := range batch {

				index := uint8((uint32(row.CounterId) + row.ObjectId) % uint32(GetWriters()))

				writers[index].taskQueue <- row
			}
		}

		for _, writer := range writers {

			close(writer.taskQueue)
		}

	}(writers)

}

func (writer *Writer) runWriter() {

	writer.waitGroup.Add(1)

	go func() {

		defer writer.waitGroup.Done()

		for row := range writer.taskQueue {

			day := time.Unix(int64(row.Timestamp), 0).Truncate(24 * time.Hour).UTC()

			path := GetProjectPath() + "/database/" + day.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(row.CounterId))

			store := writer.storePool.GetEngine(path)

			data, err := encodeData(row)

			if err != nil {

				log.Printf("failed to encode data: %s", err)
			}

			err = store.Put(row.ObjectId, row.Timestamp, data)

			if err != nil {

				log.Printf("failed to write data: %s", err)
			}

		}

		return
	}()
}

func encodeData(row RowData) ([]byte, error) {

	dataType := GetCounterType(row.CounterId)

	switch dataType {

	case TypeUint64:

		val, ok := row.Value.(uint64)

		if !ok {

			return nil, fmt.Errorf("invalid uint64 value for counter %d", row.CounterId)
		}

		data := make([]byte, 4+8)

		binary.LittleEndian.PutUint32(data, 8)

		binary.LittleEndian.PutUint64(data[4:], val)

		return data, nil

	case TypeFloat64:

		val, ok := row.Value.(float64)

		if !ok {

			return nil, fmt.Errorf("invalid float64 value for counter %d", row.CounterId)
		}

		data := make([]byte, 4+8)

		binary.LittleEndian.PutUint32(data, 8)

		binary.LittleEndian.PutUint64(data[4:], math.Float64bits(val))

		return data, nil

	case TypeString:

		str, ok := row.Value.(string)

		if !ok {

			return nil, fmt.Errorf("invalid string value for counter %d", row.CounterId)
		}

		data := make([]byte, 4+len(str))

		binary.LittleEndian.PutUint32(data, uint32(len(str)))

		copy(data[4:], str)

		return data, nil

	default:

		return nil, fmt.Errorf("unsupported data type: %d", dataType)
	}
}
