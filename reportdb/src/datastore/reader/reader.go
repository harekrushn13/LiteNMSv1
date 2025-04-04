package reader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reportdb/config"
	"reportdb/src/storage/engine"
	"reportdb/src/storage/helper"
	"strconv"
	"sync"
	"time"
)

type Query struct {
	counterID uint16

	objectID uint32

	from uint32

	to uint32

	fileCfg *helper.FileManager

	indexCfg *helper.IndexManager
}

type Reader struct {
	wg *sync.WaitGroup

	BaseDir string
}

func NewReader(wg *sync.WaitGroup, baseDir string) *Reader {

	return &Reader{

		wg: wg,

		BaseDir: baseDir,
	}
}

func NewQuery(counterID uint16, objectID uint32, from uint32, to uint32, fileCfg *helper.FileManager, indexCfg *helper.IndexManager) *Query {

	return &Query{

		counterID: counterID,

		objectID: objectID,

		from: from,

		to: to,

		fileCfg: fileCfg,

		indexCfg: indexCfg,
	}
}

func (r *Reader) StartReader(readerCount uint8, fileCfg *helper.FileManager, indexCfg *helper.IndexManager) {

	for i := uint8(0); i < readerCount; i++ {

		r.wg.Add(1)

		go runQuery(r.wg, r.BaseDir, fileCfg, indexCfg)
	}
}

func runQuery(wg *sync.WaitGroup, baseDir string, fileCfg *helper.FileManager, indexCfg *helper.IndexManager) {

	defer wg.Done()

	t := time.NewTicker(1 * time.Second)

	defer t.Stop()

	to := time.Now().Unix()

	stopTimer := time.NewTimer(10 * time.Second)

	defer stopTimer.Stop()

	for {
		select {

		case <-t.C:

			query := NewQuery(3, 3, 1743569072, uint32(to)+7, fileCfg, indexCfg)

			v, err := query.fetchData(baseDir)

			if err != nil {

				fmt.Println(err)
			}

			fmt.Printf("%#v\n", v)

		case <-stopTimer.C:

			return
		}
	}

}

func (query *Query) fetchData(baseDir string) ([]interface{}, error) {

	fromTime := time.Unix(int64(query.from), 0)

	toTime := time.Unix(int64(query.to), 0)

	fromTime = time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(), 0, 0, 0, 0, fromTime.Location())

	toTime = time.Date(toTime.Year(), toTime.Month(), toTime.Day(), 0, 0, 0, 0, toTime.Location())

	var results []interface{}

	for current := fromTime; !current.After(toTime); current = current.AddDate(0, 0, 1) {

		dayResults, err, skip := query.fetchDayData(baseDir, current)

		if skip == true && dayResults == nil {

			continue
		}

		if err != nil {

			return nil, fmt.Errorf("failed to fetch data for %s: %v", current.Format("2006-01-02"), err)
		}

		results = append(results, dayResults...)
	}

	if len(results) == 0 {

		return nil, fmt.Errorf("no data found in time range %d-%d", query.from, query.to)
	}

	return results, nil
}

func (query *Query) fetchDayData(baseDir string, day time.Time) ([]interface{}, error, bool) {

	dataType, exists := config.CounterTypeMapping[query.counterID]

	if !exists {

		return nil, fmt.Errorf("unknown counter ID: %d", query.counterID), false
	}

	indexPath := getIndexFilePath(baseDir, query.counterID, day)

	indexData, err := getIndex(query, indexPath, day, time.Now())

	entry, exists := indexData[query.objectID]

	if !exists {

		return nil, fmt.Errorf("no data found for objectID %d", query.objectID), true
	}

	if entry.EndTime < query.from || entry.StartTime > query.to {

		return nil, fmt.Errorf("data for objectID %d is not within the requested time range", query.objectID), true
	}

	validOffsets, err := getValidOffsets(entry, query.from, query.to)

	if err != nil {

		return nil, fmt.Errorf("failed to get valid offsets: %v", err), false
	}

	store := engine.NewStorageEngine(query.fileCfg, query.indexCfg, baseDir)

	dayResults, err := store.Get(query.counterID, query.objectID, day, validOffsets, dataType)

	if err != nil {

		return nil, fmt.Errorf("failed to read day data: %v", err), false
	}

	return dayResults, nil, false
}

func getIndexFilePath(baseDir string, counterID uint16, t time.Time) string {

	year := strconv.Itoa(t.Year())

	month := fmt.Sprintf("%02d", t.Month())

	day := fmt.Sprintf("%02d", t.Day())

	datePath := filepath.Join(baseDir, year, month, day)

	counterDir := filepath.Join(datePath, "counter_"+strconv.Itoa(int(counterID)))

	return filepath.Join(counterDir, "index.json")
}

func getIndex(query *Query, indexPath string, current time.Time, day time.Time) (map[uint32]*helper.IndexEntry, error) {

	var handle *helper.IndexHandle

	cY, cM, cD := current.Date()

	dY, dM, dD := day.Date()

	if cY == dY && cM == dM && cD == dD {

		handle, _ = query.indexCfg.GetIndexHandle(query.counterID)

		if handle != nil {

			handle.Lock.RLock()
		}
	}

	data, err := os.ReadFile(indexPath)

	if handle != nil {

		handle.Lock.RUnlock()
	}

	if err != nil {

		return nil, err
	}

	var indexData map[uint32]*helper.IndexEntry

	if err := json.Unmarshal(data, &indexData); err != nil {

		return nil, err
	}

	return indexData, nil
}

func getValidOffsets(entry *helper.IndexEntry, from uint32, to uint32) ([]int64, error) {

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
