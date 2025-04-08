package reader

import (
	"fmt"
	. "reportdb/config"
	. "reportdb/src/storage"
	. "reportdb/src/utils"
	"strconv"
	"sync"
	"time"
)

type Query struct {
	counterID uint16

	objectID uint32

	from uint32

	to uint32

	baseDir string // ./src/storage/database/
}

type ReaderPool struct {
	readers []*Reader
}

type Reader struct {
	storePool *StorageEnginePool

	wg *sync.WaitGroup
}

func NewReaderPool(readerCount uint8) *ReaderPool {

	return &ReaderPool{

		readers: make([]*Reader, readerCount),
	}
}

func NewQuery(counterID uint16, objectID uint32, from uint32, to uint32, baseDir string) *Query {

	return &Query{

		counterID: counterID,

		objectID: objectID,

		from: from,

		to: to,

		baseDir: baseDir,
	}
}

func (rp *ReaderPool) StartReader(readerCount uint8, baseDir string, wg *sync.WaitGroup, sp *StorageEnginePool) {

	for i := uint8(0); i < readerCount; i++ {

		rp.readers[i] = &Reader{

			storePool: sp,

			wg: wg,
		}

		wg.Add(1)

		go rp.readers[i].runReader(baseDir)
	}
}

func (r *Reader) runReader(baseDir string) {

	defer r.wg.Done()

	to := time.Now().Unix()

	ticker := time.NewTicker(time.Second)

	defer ticker.Stop()

	stopTime := time.NewTicker(12 * time.Second)

	defer stopTime.Stop()

	for {
		select {

		case <-ticker.C:
			//query := NewQuery(3, 3, uint32(to), uint32(to)+7, baseDir)
			query := NewQuery(2, 3, 1744016025, uint32(to)+10, baseDir)

			query.runQuery(r.storePool)

		case <-stopTime.C:

			return
		}
	}

}

func (q *Query) runQuery(sp *StorageEnginePool) {

	v, err := q.fetchData(sp)

	if err != nil {

		fmt.Println(err)
	}

	fmt.Printf("%#v\n", v)

}

func (q *Query) fetchData(sp *StorageEnginePool) ([]interface{}, error) {

	dataType, exists := CounterTypeMapping[q.counterID]

	if !exists {

		return nil, fmt.Errorf("unknown counter ID: %d", q.counterID)
	}

	fromTime := time.Unix(int64(q.from), 0).Truncate(24 * time.Hour).UTC()

	toTime := time.Unix(int64(q.to), 0).Truncate(24 * time.Hour).UTC()

	var results []interface{}

	for current := fromTime; !current.After(toTime); current = current.AddDate(0, 0, 1) {

		path := q.baseDir + current.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(q.counterID))

		store := sp.GetEngine(path)

		data, err := store.Get(q.objectID, q.from, q.to)

		if err != nil {

			continue
		}

		for _, val := range data {

			value, err := DecodeData(val, dataType)

			if err != nil {

				continue
			}

			results = append(results, value)
		}
	}

	if len(results) == 0 {

		return nil, fmt.Errorf("no data found in time range %d-%d", q.from, q.to)
	}

	return results, nil
}
