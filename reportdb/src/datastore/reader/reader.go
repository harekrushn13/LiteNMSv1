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

	queryEvents chan Query // channel to take query from query distributor

	resultChannel chan Response // channel to send query result

	storePool *StorePool

	waitGroup *sync.WaitGroup // to wait completion of all reader

	dayResultMapping map[string]map[uint32][][]byte // keep day-result mapping for caching counter_id-object_id

	results []interface{} // final result of query

	lock *sync.Mutex
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

	numReaders, err := GetReaders()

	if err != nil {

		return nil, fmt.Errorf("initializeReaders : Error getting readres: %v", err)
	}

	dayWorkers, err := GetDayWorkers()

	if err != nil {

		return nil, fmt.Errorf("initializeReaders : Error getting dayWorkers: %v", err)
	}

	queryBuffer, err := GetQueryBuffer()

	if err != nil {

		return nil, fmt.Errorf("initializeReaders : Error getting queryBuffer: %v", err)
	}

	readers := make([]*Reader, numReaders)

	for i := range readers {

		readers[i] = &Reader{

			id: uint8(i),

			dayPool: make(chan struct{}, dayWorkers),

			queryEvents: make(chan Query, queryBuffer),

			resultChannel: resultChannel,

			storePool: storePool,

			waitGroup: &sync.WaitGroup{},

			dayResultMapping: make(map[string]map[uint32][][]byte),

			results: make([]interface{}, 0, 10000),

			lock: &sync.Mutex{},
		}

		for j := uint8(0); j < dayWorkers; j++ {

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

			results, err := reader.FetchData(query)

			var response Response

			if err != nil {

				response = Response{

					RequestID: query.RequestID,

					Error: err.Error(),

					Timestamp: time.Now(),
				}

			} else {

				response = Response{

					RequestID: query.RequestID,

					Data: results,

					Timestamp: time.Now(),
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

func (reader *Reader) FetchData(query Query) ([]interface{}, error) {

	dataType, err := GetCounterType(query.CounterId)

	if err != nil {

		return nil, fmt.Errorf("reader.fetchData error : %v", err)
	}

	fromTime := time.Unix(int64(query.From), 0).Truncate(24 * time.Hour).UTC()

	toTime := time.Unix(int64(query.To), 0).Truncate(24 * time.Hour).UTC()

	workingDirectory, err := GetWorkingDirectory()

	if err != nil {

		return nil, fmt.Errorf("reader.GetWorkingDirectory error : %v", err)
	}

	wg := &sync.WaitGroup{}

	for current := fromTime; !current.After(toTime); current = current.AddDate(0, 0, 1) {

		path := workingDirectory + "/database/" + current.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(query.CounterId))

		if _, exists := reader.dayResultMapping[path][query.ObjectId]; exists && !reader.storePool.CheckEngineUsedPut(path) && current.After(fromTime) && current.Before(toTime) {

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

			dayResult, err := store.Get(query.ObjectId, query.From, query.To)

			if err != nil {

				log.Printf("store.Get error %s, %s", current.Format("2006/01/02"), err)

				return

			}

			reader.lock.Lock()

			objectsMap, exists := reader.dayResultMapping[path]

			if !exists {

				objectsMap = make(map[uint32][][]byte)

				reader.dayResultMapping[path] = objectsMap
			}

			reader.dayResultMapping[path][query.ObjectId] = dayResult

			reader.lock.Unlock()

			return

		}(path)
	}

	wg.Wait()

	reader.lock.Lock()

	reader.results = reader.results[:0]

	for current := fromTime; !current.After(toTime); current = current.AddDate(0, 0, 1) {

		path := workingDirectory + "/database/" + current.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(query.CounterId))

		if data, exists := reader.dayResultMapping[path][query.ObjectId]; exists {

			decodeData(data, dataType, &reader.results)

			if !current.After(fromTime) || !current.Before(toTime) {

				delete(reader.dayResultMapping[path], query.ObjectId)

			}
		}
	}

	reader.lock.Unlock()

	if len(reader.results) == 0 {

		return nil, fmt.Errorf("no data found in time range %d-%d", query.From, query.To)
	}

	return reader.results, nil
}
