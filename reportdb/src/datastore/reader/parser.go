package reader

import (
	"fmt"
	"math"
	. "reportdb/utils"
	"sort"
	"strconv"
)

func ParseResult(results map[uint32][]DataPoint, query Query) (interface{}, error) {

	if len(results) == 0 {

		return nil, fmt.Errorf("no data available for processing")
	}

	var parsedResult interface{}

	if query.Interval != "" {

		intervalSec, err := strconv.ParseUint(query.Interval, 10, 32)

		if err != nil {

			return nil, fmt.Errorf("invalid interval: %v", err)
		}

		results = bucketData(results, uint32(intervalSec), query.From, query.To)
	}

	if query.GroupByObjects == true {

		return results, nil
	}

	parsedResult = mergeAllObjects(results)

	return parsedResult, nil
}

func bucketData(data map[uint32][]DataPoint, interval uint32, from uint32, to uint32) map[uint32][]DataPoint {

	bucketed := make(map[uint32][]DataPoint)

	for objID, points := range data {

		bucketed[objID] = createBuckets(points, interval, from, to)
	}

	return bucketed
}

func createBuckets(points []DataPoint, interval uint32, from uint32, to uint32) []DataPoint {

	if interval == 0 || from > to {

		return points
	}

	sort.Slice(points, func(i, j int) bool {

		return points[i].Timestamp < points[j].Timestamp
	})

	bucketMap := make(map[uint32][]interface{})

	for _, point := range points {

		if point.Timestamp < from || point.Timestamp > to {

			continue
		}

		bucket := point.Timestamp - (point.Timestamp % interval)

		bucketMap[bucket] = append(bucketMap[bucket], point.Value)
	}

	var bucketed []DataPoint

	for time := from - (from % interval); time <= to; time += interval {

		values := bucketMap[time]

		var aggregated interface{}

		if len(values) > 0 {

			aggregated = aggregateValues(values, "AVG")

		} else {

			aggregated = 0
		}

		bucketed = append(bucketed, DataPoint{

			Timestamp: time,

			Value: aggregated,
		})

	}

	return bucketed
}

func aggregateValues(values []interface{}, aggType string) interface{} {

	if len(values) == 0 {

		return nil
	}

	switch aggType {

	case "AVG":

		return calculateAverage(values)

	case "MIN":

		return calculateMin(values)

	case "MAX":

		return calculateMax(values)

	case "COUNT":

		return len(values)

	case "SUM":

		return calculateSum(values)
	}

	if len(values) == 1 {

		return values[0]
	}

	return values
}

func processGrouped(data map[uint32][]DataPoint, aggregation string) (map[uint32]interface{}, error) {

	result := make(map[uint32]interface{})

	for objID, points := range data {

		if aggregation != "" {

			values := extractValues(points)

			result[objID] = aggregateValues(values, aggregation)

		} else {

			result[objID] = points
		}
	}

	return result, nil
}

func mergeAllObjects(data map[uint32][]DataPoint) []DataPoint {

	var allPoints []DataPoint

	for _, points := range data {

		allPoints = append(allPoints, points...)
	}

	sort.Slice(allPoints, func(i, j int) bool {

		return allPoints[i].Timestamp < allPoints[j].Timestamp
	})

	var merged []DataPoint

	var currentTime uint32

	var valuesAtSameTime []interface{}

	for _, point := range allPoints {

		if point.Timestamp != currentTime && len(valuesAtSameTime) > 0 {

			merged = append(merged, DataPoint{

				Timestamp: currentTime,

				Value: aggregateValues(valuesAtSameTime, "AVG"),
			})

			valuesAtSameTime = nil
		}

		currentTime = point.Timestamp

		valuesAtSameTime = append(valuesAtSameTime, point.Value)

	}

	if len(valuesAtSameTime) > 0 {

		merged = append(merged, DataPoint{

			Timestamp: currentTime,

			Value: aggregateValues(valuesAtSameTime, "AVG"),
		})

	}

	return merged
}

func extractValues(points []DataPoint) []interface{} {

	var values []interface{}

	for _, point := range points {

		switch v := point.Value.(type) {

		case []interface{}:

			values = append(values, v...)

		default:

			values = append(values, v)
		}
	}

	return values
}

func isNumeric(v interface{}) bool {

	switch v.(type) {

	case uint64, float64:

		return true

	default:

		return false
	}
}

func calculateAverage(values []interface{}) float64 {

	sum := 0.0

	count := 0

	for _, v := range values {

		if f, ok := convertToFloat64(v); ok {

			sum += f

			count++
		}
	}

	if count == 0 {

		return 0
	}

	return sum / float64(count)
}

func calculateMin(values []interface{}) float64 {

	min := math.MaxFloat64

	for _, v := range values {

		if f, ok := convertToFloat64(v); ok && f < min {

			min = f
		}
	}

	if min == math.MaxFloat64 {

		return 0
	}

	return min
}

func calculateMax(values []interface{}) float64 {
	max := -math.MaxFloat64

	for _, v := range values {

		if f, ok := convertToFloat64(v); ok && f > max {

			max = f
		}
	}

	if max == -math.MaxFloat64 {

		return 0
	}
	return max
}

func calculateSum(values []interface{}) float64 {
	sum := 0.0

	for _, v := range values {

		if f, ok := convertToFloat64(v); ok {

			sum += f

		}
	}

	return sum
}

func convertToFloat64(v interface{}) (float64, bool) {

	switch n := v.(type) {

	case float64:

		return n, true

	case uint64:

		return float64(n), true

	default:
		return 0, false
	}
}
