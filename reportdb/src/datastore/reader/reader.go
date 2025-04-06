package reader

import (
	"fmt"
	. "reportdb/config"
	. "reportdb/src/storage"
	. "reportdb/src/storage/helper"
	. "reportdb/src/utils"
	"strconv"
	"sync"
	"syscall"
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
	readres []*Reader
}

type Reader struct {
	storePool *StorageEnginePool

	wg *sync.WaitGroup
}

func NewReaderPool(readerCount uint8) *ReaderPool {

	return &ReaderPool{

		readres: make([]*Reader, readerCount),
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

		rp.readres[i] = &Reader{

			storePool: sp,

			wg: wg,
		}

		wg.Add(1)

		go rp.readres[i].runReader(baseDir)
	}
}

func (r *Reader) runReader(baseDir string) {

	defer r.wg.Done()

	to := time.Now().Unix()

	ticker := time.NewTicker(time.Second)

	defer ticker.Stop()

	stopTime := time.NewTicker(10 * time.Second)

	defer stopTime.Stop()

	for {
		select {

		case <-ticker.C:
			//query := NewQuery(3, 3, uint32(to), uint32(to)+7, baseDir)
			query := NewQuery(3, 3, 1743767934, uint32(to)+7, baseDir)

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

		dayResults, err, skip := q.fetchDayData(store, dataType)

		if skip == true && dayResults == nil {

			continue
		}

		if err != nil {

			return nil, fmt.Errorf("failed to fetch data for %s: %v", current.Format("2006-01-02"), err)
		}

		results = append(results, dayResults...)
	}

	if len(results) == 0 {

		return nil, fmt.Errorf("no data found in time range %d-%d", q.from, q.to)
	}

	return results, nil
}

func (q *Query) fetchDayData(store *StorageEngine, dataType DataType) ([]interface{}, error, bool) {

	entry, err := store.IndexCfg.GetIndexMap(q.objectID)

	if err != nil {

		return nil, err, true
	}

	if entry.EndTime < q.from || entry.StartTime > q.to {

		return nil, fmt.Errorf("data for objectID %d is not within the requested time range", q.objectID), true
	}

	validOffsets, err := getValidOffsets(entry, q.from, q.to)

	if err != nil {

		return nil, fmt.Errorf("failed to get valid offsets: %v", err), false
	}

	partition := GetPartition(q.objectID, store.PartitionCount)

	handle, err := store.FileCfg.GetHandle(partition)

	if err != nil {

		return nil, err, false
	}

	data, err := syscall.Mmap(int(handle.File.Fd()), 0, int(handle.Offset), syscall.PROT_READ, syscall.MAP_SHARED)

	var dayResults []interface{}

	for _, offset := range validOffsets {

		value, err := store.Get(data, offset, dataType)

		if err != nil {

			continue
		}

		decode, err := DecodeData(value, dataType)

		if err != nil {

			continue
		}

		dayResults = append(dayResults, decode)

	}

	if err != nil {

		return nil, fmt.Errorf("failed to read day data: %v", err), false
	}

	return dayResults, nil, false
}

func getValidOffsets(entry *IndexEntry, from uint32, to uint32) ([]int64, error) {

	var validOffsets []int64

	length := uint8(len(entry.Offsets))

	var start uint8

	var end uint8

	if from <= entry.StartTime {

		start = uint8(0)

	} else {

		start = uint8(from - entry.StartTime)
	}

	if to >= entry.EndTime {

		end = length

	} else {

		end = length - uint8(entry.EndTime-to)
	}

	validOffsets = append(validOffsets, entry.Offsets[start:end]...)

	return validOffsets, nil
}
