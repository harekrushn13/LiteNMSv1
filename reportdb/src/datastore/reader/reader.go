package reader

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

func FetchData(query Query, storePool *StorePool) ([]interface{}, error) {

	dataType, err := GetCounterType(query.CounterId)

	if err != nil {

		return nil, fmt.Errorf("reader.fetchData error : %v", err)
	}

	fromTime := time.Unix(int64(query.From), 0).Truncate(24 * time.Hour).UTC()

	toTime := time.Unix(int64(query.To), 0).Truncate(24 * time.Hour).UTC()

	var results []interface{}

	for current := fromTime; !current.After(toTime); current = current.AddDate(0, 0, 1) {

		workingDirectory, err := GetWorkingDirectory()

		if err != nil {

			return nil, fmt.Errorf("reader.fetchData error : %v", err)
		}

		path := workingDirectory + "/database/" + current.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(query.CounterId))

		store, error := storePool.GetEngine(path, false)

		if error != nil {

			//log.Printf("reader.fetchData error : %v", error)

			continue
		}

		dayResult, err := store.Get(query.ObjectId, query.From, query.To)

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

		return nil, fmt.Errorf("no data found in time range %d-%d", query.From, query.To)
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
