package reader

import (
	"fmt"
	"math"
	. "reportdb/utils"
	"sort"
)

func (reader *Reader) ParseResult(query Query) (interface{}, error) {

	if len(reader.results) == 0 {

		return nil, fmt.Errorf("no data available for processing")
	}

	dataType, err := GetCounterType(query.CounterID)

	if err != nil {

		return nil, fmt.Errorf("reader.fetchData error : %v", err)
	}

	if dataType == TypeString {

		return reader.results, nil
	}

	if query.Interval == 0 {

		if query.GroupByObjects {

			return reader.GridQuery(query)
		}

		return reader.GaugeQuery(query)
	}

	return reader.HistogramQuery(query)
}

func (reader *Reader) GaugeQuery(query Query) (interface{}, error) {

	reader.dataValues = reader.dataValues[:0]

	for _, points := range reader.results {

		reader.dataValues = append(reader.dataValues, aggregateValues(reader.getValues(points), query.Aggregation))

	}

	return aggregateValues(reader.dataValues, query.Aggregation), nil
}

func (reader *Reader) HistogramQuery(query Query) (interface{}, error) {

	bucketed := reader.bucketData(uint32(query.Interval), query.From, query.To, query.Aggregation)

	if query.GroupByObjects {

		return bucketed, nil
	}

	return reader.mergeAllObjects(query.Aggregation), nil
}

func (reader *Reader) GridQuery(query Query) (interface{}, error) {

	for k := range reader.grid {

		delete(reader.grid, k)
	}

	for objID, points := range reader.results {

		reader.grid[objID] = aggregateValues(reader.getValues(points), query.Aggregation)
	}

	return reader.grid, nil
}

func (reader *Reader) bucketData(interval uint32, from uint32, to uint32, aggregation string) map[uint32][]DataPoint {

	for k := range reader.bucketed {

		delete(reader.bucketed, k)
	}

	for objID, points := range reader.results {

		reader.bucketed[objID] = reader.createBuckets(points, interval, from, to, aggregation)
	}

	return reader.bucketed
}

func (reader *Reader) createBuckets(points []DataPoint, interval uint32, from uint32, to uint32, aggregation string) []DataPoint {

	sort.Slice(points, func(i, j int) bool {

		return points[i].Timestamp < points[j].Timestamp
	})

	for k := range reader.bucketMap { // map[timestamp]->[points]

		delete(reader.bucketMap, k)
	}

	for _, point := range points {

		if point.Timestamp < from || point.Timestamp > to {

			continue
		}

		bucket := point.Timestamp - (point.Timestamp % interval)

		reader.bucketMap[bucket] = append(reader.bucketMap[bucket], point.Value)
	}

	var bucketed = make([]DataPoint, 0, (to-from)/interval)

	for time := from - (from % interval); time <= to; time += interval {

		values := reader.bucketMap[time]

		var aggregated interface{}

		if len(values) > 0 {

			aggregated = aggregateValues(values, aggregation)

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

func (reader *Reader) mergeAllObjects(aggregation string) []DataPoint {

	reader.allDataPoints = reader.allDataPoints[:0]

	for _, points := range reader.bucketed {

		reader.allDataPoints = append(reader.allDataPoints, points...)
	}

	sort.Slice(reader.allDataPoints, func(i, j int) bool {

		return reader.allDataPoints[i].Timestamp < reader.allDataPoints[j].Timestamp
	})

	var currentTime uint32

	reader.dataPoints = reader.dataPoints[:0]

	reader.dataValues = reader.dataValues[:0]

	for _, point := range reader.allDataPoints {

		if point.Timestamp != currentTime && len(reader.dataValues) > 0 {

			reader.dataPoints = append(reader.dataPoints, DataPoint{

				Timestamp: currentTime,

				Value: aggregateValues(reader.dataValues, aggregation),
			})

			reader.dataValues = reader.dataValues[:0]
		}

		currentTime = point.Timestamp

		reader.dataValues = append(reader.dataValues, point.Value)

	}

	if len(reader.dataValues) > 0 {

		reader.dataPoints = append(reader.dataPoints, DataPoint{

			Timestamp: currentTime,

			Value: aggregateValues(reader.dataValues, aggregation),
		})

	}

	return reader.dataPoints
}

func (reader *Reader) getValues(points []DataPoint) []interface{} {

	reader.getDataValues = reader.getDataValues[:0]

	for _, point := range points {

		reader.getDataValues = append(reader.getDataValues, point.Value)

	}

	return reader.getDataValues
}

func aggregateValues(values []interface{}, aggType string) interface{} {

	if len(values) == 0 {

		return nil
	}

	switch aggType {

	case "AVG":

		return getAverage(values)

	case "MIN":

		return getMin(values)

	case "MAX":

		return getMax(values)

	case "SUM":

		return getSum(values)
	}

	if len(values) == 1 {

		return values[0]
	}

	return values
}

func getAverage(values []interface{}) float64 {

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

func getMin(values []interface{}) float64 {

	minVal := math.MaxFloat64

	for _, v := range values {

		if f, ok := convertToFloat64(v); ok && f < minVal {

			minVal = f
		}
	}

	if minVal == math.MaxFloat64 {

		return 0
	}

	return minVal
}

func getMax(values []interface{}) float64 {

	maxVal := -math.MaxFloat64

	for _, v := range values {

		if f, ok := convertToFloat64(v); ok && f > maxVal {

			maxVal = f
		}
	}

	if maxVal == -math.MaxFloat64 {

		return 0
	}
	return maxVal
}

func getSum(values []interface{}) float64 {

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
