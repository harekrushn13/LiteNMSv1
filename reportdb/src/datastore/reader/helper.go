package reader

import (
	"context"
	"encoding/binary"
	"fmt"
	"go.uber.org/zap"
	"math"
	. "reportdb/logger"
	. "reportdb/storage"
	. "reportdb/utils"
	"strconv"
	"sync"
	"time"
)

func (reader *Reader) FetchData(query Query) (map[uint32][]DataPoint, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)

	defer cancel()

	fromTime, toTime := getTimeBounds(query.From, query.To)

	dataType, err := GetCounterType(query.CounterID)

	if err != nil {

		return nil, fmt.Errorf("reader.fetchData error : %v", err)
	}

	workingDirectory := GetWorkingDirectory()

	wg := &sync.WaitGroup{}

	for current := fromTime; !current.After(toTime); current = current.AddDate(0, 0, 1) {

		path, store, err := reader.getStorePathAndEngine(current, query.CounterID, workingDirectory)

		if err != nil {

			continue
		}

		if len(query.ObjectIDs) == 0 {

			objects, err := store.GetKeys()

			if err != nil {

				return nil, fmt.Errorf("GetKeys : %v", err)
			}

			reader.fetchForObjectIDs(ctx, wg, path, store, query, dataType, objects)

		} else {

			reader.fetchForObjectIDs(ctx, wg, path, store, query, dataType, nil)
		}

	}

	if err := reader.waitWithContext(ctx, wg); err != nil {

		return nil, err
	}

	reader.mergeResults(query, fromTime, toTime, workingDirectory)

	if len(reader.results) == 0 {

		return nil, fmt.Errorf("no data found in time range %d-%d", query.From, query.To)
	}

	return reader.results, nil
}

func getTimeBounds(from, to uint32) (time.Time, time.Time) {

	return time.Unix(int64(from), 0).Truncate(24 * time.Hour).UTC(),
		time.Unix(int64(to), 0).Truncate(24 * time.Hour).UTC()
}

func (reader *Reader) getStorePathAndEngine(current time.Time, counterID uint16, base string) (string, *StoreEngine, error) {

	path := base + "/database/" + current.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(counterID))

	store, err := reader.storePool.GetEngine(path, false)

	if err != nil {

		Logger.Error("GetEngine failed", zap.Error(err), zap.Uint16("counter_id", counterID))

		return "", nil, err
	}

	return path, store, nil
}

func (reader *Reader) fetchForObjectIDs(ctx context.Context, wg *sync.WaitGroup, path string, store *StoreEngine, query Query, dataType DataType, objects []uint32) {

	var objectIds []uint32

	if len(objects) > 0 {

		objectIds = objects

	} else {

		objectIds = query.ObjectIDs
	}

	for _, objectID := range objectIds {

		reader.lock.RLock()

		skip := reader.dayResultMapping[path][objectID] != nil &&
			!reader.storePool.CheckEngineUsedPut(path)

		reader.lock.RUnlock()

		if skip {

			continue
		}

		select {

		case <-ctx.Done():

			return

		case <-reader.dayPool:

			wg.Add(1)

			go func(objectID uint32) {

				defer func() {

					wg.Done()

					reader.dayPool <- struct{}{}
				}()

				result, err := store.Get(objectID, query.From, query.To)

				if err != nil {

					Logger.Error("Get failed", zap.Error(err), zap.Uint32("object_id", objectID))

					return
				}

				var dp []DataPoint

				decodeData(result, dataType, &dp)

				reader.lock.Lock()

				if _, ok := reader.dayResultMapping[path]; !ok {

					reader.dayResultMapping[path] = make(map[uint32][]DataPoint)
				}

				reader.dayResultMapping[path][objectID] = dp

				reader.lock.Unlock()

			}(objectID)
		}
	}
}

func (reader *Reader) waitWithContext(ctx context.Context, wg *sync.WaitGroup) error {

	done := make(chan struct{})

	go func() {

		wg.Wait()

		close(done)
	}()

	select {

	case <-ctx.Done():

		return fmt.Errorf("query timeout")

	case <-done:

		return nil
	}
}

func (reader *Reader) mergeResults(query Query, fromTime, toTime time.Time, base string) {

	reader.lock.Lock()

	defer reader.lock.Unlock()

	reader.results = make(map[uint32][]DataPoint)

	for current := fromTime; !current.After(toTime); current = current.AddDate(0, 0, 1) {

		path := base + "/database/" + current.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(query.CounterID))

		var objectIDs []uint32

		if len(query.ObjectIDs) == 0 {

			for id := range reader.dayResultMapping[path] {

				objectIDs = append(objectIDs, id)
			}

		} else {

			objectIDs = query.ObjectIDs
		}

		for _, id := range objectIDs {

			if data, ok := reader.dayResultMapping[path][id]; ok && len(data) > 0 {

				reader.results[id] = data
			}

			if !current.After(fromTime) || !current.Before(toTime) {

				delete(reader.dayResultMapping[path], id)
			}
		}
	}
}

func decodeData(data [][]byte, dataType DataType, result *[]DataPoint) {

	for _, row := range data {

		switch dataType {

		case TypeUint64:

			*result = append(*result, DataPoint{

				Timestamp: binary.LittleEndian.Uint32(row[:4]),

				Value: binary.LittleEndian.Uint64(row[4:]),
			})

		case TypeFloat64:

			*result = append(*result, DataPoint{

				Timestamp: binary.LittleEndian.Uint32(row[:4]),

				Value: math.Float64frombits(binary.LittleEndian.Uint64(row[4:])),
			})

		case TypeString:

			*result = append(*result, DataPoint{

				Timestamp: binary.LittleEndian.Uint32(row[:4]),

				Value: string(row[4:]),
			})

		}
	}
}
