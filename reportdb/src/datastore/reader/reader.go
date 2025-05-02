package reader

import (
	"fmt"
	"go.uber.org/zap"
	. "reportdb/logger"
	. "reportdb/storage"
	. "reportdb/utils"
	"sync"
)

type Reader struct {
	id uint8

	dayPool chan struct{} // to read multiple day in parallel for query

	queryEvents chan QueryReceive // channel to take query from query distributor

	resultChannel chan Response // channel to send query result

	storePool *StorePool

	waitGroup *sync.WaitGroup // to wait completion of all reader

	dayResultMapping map[string]map[uint32][]DataPoint // keep day-result mapping for caching counter_id-object_id

	results map[uint32][]DataPoint // result of query

	lock *sync.RWMutex
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

			lock: &sync.RWMutex{},
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

				Logger.Error("Error fetching data from reader", zap.Error(err))

				response := Response{

					RequestID: query.RequestID,

					Error: err.Error(),
				}

				reader.resultChannel <- response

				continue
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
