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

	objectPool chan struct{} // to read multiple key data in parallel for a query

	queryEvents chan QueryReceive // channel to take query from query distributor

	resultChannel chan Response // channel to send query result

	storePool *StorePool

	waitGroup *sync.WaitGroup // to wait completion of all readers

	objectsMapping map[string][]uint32 // useful when no objectId given in a query

	results map[uint32][]DataPoint // result of query

	ParserBuffer
}

type ParserBuffer struct {
	dataValues []interface{}

	getDataValues []interface{}

	dataPoints []DataPoint

	allDataPoints []DataPoint

	grid map[uint32]interface{} // map[objectID]->value

	bucketed map[uint32][]DataPoint // map[objectID]->[]DataPoint

	bucketMap map[uint32][]interface{} // map[timestamp]->[values]
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

			objectPool: make(chan struct{}, GetObjectWorkers()),

			queryEvents: make(chan QueryReceive, GetQueryBuffer()),

			resultChannel: resultChannel,

			storePool: storePool,

			waitGroup: &sync.WaitGroup{},

			objectsMapping: make(map[string][]uint32),

			results: make(map[uint32][]DataPoint),

			ParserBuffer: ParserBuffer{

				dataValues: make([]interface{}, 0, 100),

				getDataValues: make([]interface{}, 0, 100),

				dataPoints: make([]DataPoint, 0, 100),

				allDataPoints: make([]DataPoint, 0, 100),

				grid: make(map[uint32]interface{}),

				bucketed: make(map[uint32][]DataPoint),

				bucketMap: make(map[uint32][]interface{}),
			},
		}

		for j := 0; j < GetObjectWorkers(); j++ {

			readers[i].objectPool <- struct{}{}
		}

	}

	return readers, nil
}

func (reader *Reader) runReader() {

	reader.waitGroup.Add(1)

	go func() {

		defer reader.waitGroup.Done()

		for query := range reader.queryEvents {

			err := reader.FetchData(query.Query)

			if err != nil {

				Logger.Error("Error fetching data from reader", zap.Error(err))

				response := Response{

					RequestID: query.RequestID,

					Error: err.Error(),
				}

				reader.resultChannel <- response

				continue
			}

			parseResult, err := reader.ParseResult(query.Query)

			var response Response

			if err != nil {

				Logger.Error("Error fetching data from reader", zap.Error(err))

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
