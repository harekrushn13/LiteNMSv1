package reader

import (
	"fmt"
	. "reportdb/utils"
	"testing"
)

func TestParseResult_GroupedAvg(t *testing.T) {
	dummyResults := map[uint32][]DataPoint{
		1: {
			{Timestamp: 1, Value: float64(10)},
			{Timestamp: 2, Value: float64(20)},
			{Timestamp: 3, Value: float64(30)},
			{Timestamp: 4, Value: float64(40)},
			{Timestamp: 5, Value: float64(50)},
			{Timestamp: 6, Value: float64(60)},
		},
		2: {
			{Timestamp: 3, Value: float64(34)},
			{Timestamp: 4, Value: float64(44)},
			{Timestamp: 5, Value: float64(54)},
			{Timestamp: 6, Value: float64(64)},
			{Timestamp: 7, Value: float64(74)},
			{Timestamp: 8, Value: float64(84)},
		},
	}

	query := Query{
		From:     0,
		To:       10,
		Interval: "2",
		//GroupByObjects: true,
		Aggregation: "AVG",
	}

	result, err := ParseResult(dummyResults, query)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fmt.Println(result)
}
