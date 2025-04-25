package reader

import (
	"fmt"
	. "reportdb/utils"
	"testing"
)

func TestParseResult_GroupedAvg(t *testing.T) {

	dummyResults := map[uint32][]DataPoint{
		1: {
			{Timestamp: 1, Value: uint64(10)},
			{Timestamp: 2, Value: uint64(20)},
			{Timestamp: 3, Value: uint64(30)},
			{Timestamp: 4, Value: uint64(40)},
			{Timestamp: 5, Value: uint64(50)},
			{Timestamp: 6, Value: uint64(60)},
		},
		2: {
			{Timestamp: 3, Value: uint64(34)},
			{Timestamp: 4, Value: uint64(44)},
			{Timestamp: 5, Value: uint64(54)},
			{Timestamp: 6, Value: uint64(64)},
			{Timestamp: 7, Value: uint64(74)},
			{Timestamp: 8, Value: uint64(84)},
		},
	}

	//dummyResults := map[uint32][]DataPoint{
	//	1: {
	//		{Timestamp: 1, Value: "abcd"},
	//		{Timestamp: 2, Value: "abcd"},
	//		{Timestamp: 3, Value: "abcd"},
	//		{Timestamp: 4, Value: "abcd"},
	//		{Timestamp: 5, Value: "abcd"},
	//		{Timestamp: 6, Value: "abcd"},
	//	},
	//	2: {
	//		{Timestamp: 3, Value: "abcd"},
	//		{Timestamp: 4, Value: "abcd"},
	//		{Timestamp: 5, Value: "abcd"},
	//		{Timestamp: 6, Value: "abcd"},
	//		{Timestamp: 7, Value: "abcd"},
	//		{Timestamp: 8, Value: "abcd"},
	//	},
	//}

	query := Query{
		//CounterID: 3,
		//ObjectIDs: []uint32{2},
		From: 0,
		To:   10,
		//Interval:  "2",
		//GroupByObjects: true,
		Aggregation: "AVG",
	}

	result, err := ParseResult(dummyResults, query)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fmt.Println(result)
}
