package writer

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	. "reportdb/storage"
	. "reportdb/utils"
	"strconv"
	"time"
)

type Writer struct {
	id uint8

	events chan Events

	storePool *StorePool
}

func initializeWriters(storePool *StorePool) ([]*Writer, error) {

	numWriters, err := GetWriters()

	if err != nil {

		return nil, fmt.Errorf("initializeWriters : Error getting writers: %v", err)
	}

	eventsBuffer, err := GetEventsBuffer()

	if err != nil {

		return nil, fmt.Errorf("initializeWriters : Error getting writers: %v", err)
	}

	writers := make([]*Writer, numWriters)

	for i := range writers {

		writers[i] = &Writer{

			id: uint8(i),

			events: make(chan Events, eventsBuffer),

			storePool: storePool,
		}
	}

	return writers, nil
}

func StartWriter(storePool *StorePool) ([]*Writer, error) {

	writers, err := initializeWriters(storePool)

	if err != nil {

		return nil, fmt.Errorf("StartWriter : Error getting writers: %v", err)
	}

	for _, writer := range writers {

		writer.runWriter()
	}

	return writers, nil
}

func ShutdownWriters(writers []*Writer) {

	for _, writer := range writers {

		close(writer.events)
	}
}

func (writer *Writer) runWriter() {

	go func(writer *Writer) {

		for row := range writer.events {

			day := time.Unix(int64(row.Timestamp), 0).Truncate(24 * time.Hour).UTC()

			workingDirectory, err := GetWorkingDirectory()

			if err != nil {

				log.Printf("writer.runWriter : Error getting working directory: %v", err)

				return
			}

			path := workingDirectory + "/database/" + day.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(row.CounterId))

			store, err := writer.storePool.GetEngine(path, true)

			if err != nil {

				log.Printf("writer.runWriter : Error getting store: %v", err)

				return
			}

			data, err := encodeData(row)

			if err != nil {

				log.Printf("writer.runWriter : failed to encode data: %s", err)

				return
			}

			err = store.Put(row.ObjectId, row.Timestamp, data)

			if err != nil {

				log.Printf("writer.runWriter : failed to write data: %s", err)

				return
			}

		}

		return

	}(writer)
}

func encodeData(row Events) ([]byte, error) {

	dataType, err := GetCounterType(row.CounterId)

	if err != nil {

		return nil, fmt.Errorf("encodeData : Error getting counter type: %v", err)
	}

	switch dataType {

	case TypeUint64:

		val, ok := row.Value.(float64)

		if !ok {

			return nil, fmt.Errorf("encodeData : invalid uint64 value for counter %d", row.CounterId)
		}

		newvalue := uint64(val)

		data := make([]byte, 12)

		binary.LittleEndian.PutUint32(data, 8)

		binary.LittleEndian.PutUint64(data[4:], newvalue)

		return data, nil

	case TypeFloat64:

		val, ok := row.Value.(float64)

		if !ok {

			return nil, fmt.Errorf("encodeData : invalid float64 value for counter %d", row.CounterId)
		}

		data := make([]byte, 12)

		binary.LittleEndian.PutUint32(data, 8)

		binary.LittleEndian.PutUint64(data[4:], math.Float64bits(val))

		return data, nil

	case TypeString:

		str, ok := row.Value.(string)

		if !ok {

			return nil, fmt.Errorf("encodeData : invalid string value for counter %d", row.CounterId)
		}

		data := make([]byte, 4+len(str))

		binary.LittleEndian.PutUint32(data, uint32(len(str)))

		copy(data[4:], str)

		return data, nil

	default:

		return nil, fmt.Errorf("encodeData : unsupported data type: %d", dataType)
	}
}
