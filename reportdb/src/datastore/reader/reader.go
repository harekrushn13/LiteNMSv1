package reader

import (
	"fmt"
	"reportdb/src/storage/engine"
	"time"
)

func FetchData(counterID uint16, objectID uint32, from uint32, to uint32) ([]interface{}, error) {

	fromTime := time.Unix(int64(from), 0)

	toTime := time.Unix(int64(to), 0)

	fromTime = time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(), 0, 0, 0, 0, fromTime.Location())

	toTime = time.Date(toTime.Year(), toTime.Month(), toTime.Day(), 0, 0, 0, 0, toTime.Location())

	var results []interface{}

	for current := fromTime; !current.After(toTime); current = current.AddDate(0, 0, 1) {

		dayResults, err, skip := engine.Get(counterID, objectID, from, to, current)

		if skip == true && dayResults == nil {

			continue
		}

		if err != nil {

			return nil, fmt.Errorf("failed to fetch data for %s: %v", current.Format("2006-01-02"), err)
		}

		results = append(results, dayResults...)
	}

	if len(results) == 0 {

		return nil, fmt.Errorf("no data found in time range %d-%d", from, to)
	}

	return results, nil
}
