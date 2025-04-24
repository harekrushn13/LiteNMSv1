package reader

import (
	"fmt"
	"log"
	. "reportdb/storage"
	. "reportdb/utils"
	"strconv"
	"sync"
	"time"
)

type Reader struct {
	id uint8

	dayPool chan struct{} // to read multiple day in parallel for query

	queryEvents chan QueryReceive // channel to take query from query distributor

	resultChannel chan Response // channel to send query result

	storePool *StorePool

	waitGroup *sync.WaitGroup // to wait completion of all reader

	dayResultMapping map[string]map[uint32][]DataPoint // keep day-result mapping for caching counter_id-object_id

	results map[uint32][]DataPoint // final result of query

	lock *sync.Mutex
}

type DataPoint struct {
	Timestamp uint32 `json:"timestamp"`

	Value interface{} `json:"value"`
}

func StartReaders(storePool *StorePool, resultChannel chan Response) ([]*Reader, error) {

	readers, err := initializeReaders(storePool, resultChannel)

	if err != nil {

		return readers, fmt.Errorf("startReaders : Error initializing readers: %v", err)
	}

	for _, reader := range readers {

		reader.runReader()
	}

	return readers, nil
}

func initializeReaders(storePool *StorePool, resultChannel chan Response) ([]*Reader, error) {

	readers := make([]*Reader, GetReaders())

	for i := range readers {

		readers[i] = &Reader{

			id: uint8(i),

			dayPool: make(chan struct{}, GetDayWorkers()),

			queryEvents: make(chan QueryReceive, GetQueryBuffer()),

			resultChannel: resultChannel,

			storePool: storePool,

			waitGroup: &sync.WaitGroup{},

			dayResultMapping: make(map[string]map[uint32][]DataPoint),

			results: make(map[uint32][]DataPoint),

			lock: &sync.Mutex{},
		}

		for j := 0; j < GetDayWorkers(); j++ {

			readers[i].dayPool <- struct{}{}
		}
	}

	return readers, nil
}

func (reader *Reader) runReader() {

	reader.waitGroup.Add(1)

	go func() {

		defer reader.waitGroup.Done()

		for query := range reader.queryEvents {

			results, err := reader.FetchData(query.Query)

			if err != nil {

				log.Printf("Error fetching data from reader: %v", err)

				response := Response{

					RequestID: query.RequestID,

					Error: err.Error(),
				}

				reader.resultChannel <- response
			}

			parseResult, err := ParseResult(results, query.Query)

			var response Response

			if err != nil {

				response = Response{

					RequestID: query.RequestID,

					Error: err.Error(),
				}

			} else {

				response = Response{

					RequestID: query.RequestID,

					Data: parseResult,
				}
			}

			reader.resultChannel <- response
		}

	}()
}

func ShutdownReaders(readers []*Reader) {

	for _, reader := range readers {

		close(reader.queryEvents)

		reader.waitGroup.Wait()
	}

}

func (reader *Reader) FetchData(query Query) (map[uint32][]DataPoint, error) {

	dataType, err := GetCounterType(query.CounterID)

	if err != nil {

		return nil, fmt.Errorf("reader.fetchData error : %v", err)
	}

	fromTime := time.Unix(int64(query.From), 0).Truncate(24 * time.Hour).UTC()

	toTime := time.Unix(int64(query.To), 0).Truncate(24 * time.Hour).UTC()

	workingDirectory := GetWorkingDirectory()

	if err != nil {

		return nil, fmt.Errorf("reader.GetWorkingDirectory error : %v", err)
	}

	wg := &sync.WaitGroup{}

	for current := fromTime; !current.After(toTime); current = current.AddDate(0, 0, 1) {

		path := workingDirectory + "/database/" + current.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(query.CounterID))

		for _, ObjectId := range query.ObjectIDs {

			if _, exists := reader.dayResultMapping[path][ObjectId]; exists && !reader.storePool.CheckEngineUsedPut(path) && current.After(fromTime) && current.Before(toTime) {

				continue
			}

			<-reader.dayPool

			wg.Add(1)

			go func(path string) {

				defer func() {

					wg.Done()

					reader.dayPool <- struct{}{}
				}()

				store, err := reader.storePool.GetEngine(path, false)

				if err != nil {

					log.Printf("reader.fetchData error : %v", err)

					return
				}

				dayResult, err := store.Get(ObjectId, query.From, query.To)

				if err != nil {

					log.Printf("store.Get error %s, %s", current.Format("2006/01/02"), err)

					return

				}

				var decodeDayResult []DataPoint

				decodeData(dayResult, dataType, &decodeDayResult)

				reader.lock.Lock()

				objectsMap, exists := reader.dayResultMapping[path]

				if !exists {

					objectsMap = make(map[uint32][]DataPoint)

					reader.dayResultMapping[path] = objectsMap
				}

				reader.dayResultMapping[path][ObjectId] = decodeDayResult

				reader.lock.Unlock()

				return

			}(path)

		}

	}

	wg.Wait()

	reader.lock.Lock()

	//reader.results = reader.results[:0]
	reader.results = make(map[uint32][]DataPoint)

	for current := fromTime; !current.After(toTime); current = current.AddDate(0, 0, 1) {

		path := workingDirectory + "/database/" + current.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(query.CounterID))

		for _, ObjectId := range query.ObjectIDs {

			if data, exists := reader.dayResultMapping[path][ObjectId]; exists {

				if len(data) > 0 {

					//reader.results = append(reader.results, data)

					reader.results[ObjectId] = data
				}

				if !current.After(fromTime) || !current.Before(toTime) {

					delete(reader.dayResultMapping[path], ObjectId)

				}
			}
		}
	}

	reader.lock.Unlock()

	if len(reader.results) == 0 {

		return nil, fmt.Errorf("no data found in time range %d-%d", query.From, query.To)
	}

	return reader.results, nil
}
