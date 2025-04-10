package reader

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

type Query struct {
	counterId uint16

	objectId uint32

	from uint32

	to uint32
}

func NewQuery(counterID uint16, objectID uint32, from uint32, to uint32) *Query {

	return &Query{

		counterId: counterID,

		objectId: objectID,

		from: from,

		to: to,
	}
}

type Reader struct {
	storePool *StorePool

	waitGroup *sync.WaitGroup
}

func StartReader(waitGroup *sync.WaitGroup, storePool *StorePool) {

	numReaders, err := GetReaders()

	if err != nil {

		log.Fatal(err)
	}

	readers := make([]*Reader, numReaders)

	for i := range numReaders {

		readers[i] = &Reader{

			storePool: storePool,

			waitGroup: waitGroup,
		}
	}

	for _, reader := range readers {

		reader.runReader()
	}

}

func (reader *Reader) runReader() {

	reader.waitGroup.Add(1)

	go func() {

		defer reader.waitGroup.Done()

		today := time.Now().Unix()

		ticker := time.NewTicker(time.Second)

		defer ticker.Stop()

		stopIndexTime, err := GetStopIndexSaving()

		if err != nil {

			log.Fatal(err)
		}

		stopTime := time.NewTicker(time.Duration(stopIndexTime) * time.Second)

		defer stopTime.Stop()

		for {
			select {

			case <-ticker.C:
				//query := NewQuery(3, 3, uint32(today), uint32(today)+7, baseDir)

				NewQuery(3, 3, uint32(today), uint32(today)+12).runQuery(reader.storePool)

			case <-stopTime.C:

				return
			}
		}

	}()

}

func (query *Query) runQuery(storePool *StorePool) {

	results, err := query.fetchData(storePool)

	if err != nil {

		//log.Printf("query.runQuery : Error fetching data for query %d: %s", query.objectId, err)
		fmt.Println(err)
	}

	fmt.Printf("%#v\n", results)

}

func (query *Query) fetchData(storePool *StorePool) ([]interface{}, error) {

	dataType, err := GetCounterType(query.counterId)

	if err != nil {

		return nil, fmt.Errorf("reader.fetchData error : %v", err)
	}

	fromTime := time.Unix(int64(query.from), 0).Truncate(24 * time.Hour).UTC()

	toTime := time.Unix(int64(query.to), 0).Truncate(24 * time.Hour).UTC()

	var results []interface{}

	for current := fromTime; !current.After(toTime); current = current.AddDate(0, 0, 1) {

		workingDirectory, err := GetWorkingDirectory()

		if err != nil {

			return nil, fmt.Errorf("reader.fetchData error : %v", err)
		}

		path := workingDirectory + "/database/" + current.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(query.counterId))

		store, error := storePool.GetEngine(path, false)

		if error != nil {

			log.Printf("reader.fetchData error : %v", error)

			continue
		}

		dayResult, err := store.Get(query.objectId, query.from, query.to)

		if err != nil {

			log.Printf("store.Get error %s, %s", current.Format("2006/01/02"), err)

			continue
		}

		for _, data := range dayResult {

			value, err := decodeData(data, dataType)

			if err != nil {

				log.Printf("decodeData error %s %s", current.Format("2006/01/02"), err)

				continue
			}

			results = append(results, value)
		}
	}

	if len(results) == 0 {

		return nil, fmt.Errorf("no data found in time range %d-%d", query.from, query.to)
	}

	return results, nil
}

func decodeData(data []byte, dataType DataType) (interface{}, error) {

	var value interface{}

	switch dataType {

	case TypeUint64:

		value = binary.LittleEndian.Uint64(data)

	case TypeFloat64:

		value = math.Float64frombits(binary.LittleEndian.Uint64(data))

	case TypeString:

		value = string(data)

	default:

		return nil, fmt.Errorf("unsupported data type: %T", dataType)
	}

	return value, nil
}
