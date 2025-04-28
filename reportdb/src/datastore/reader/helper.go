package reader

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	. "reportdb/utils"
	"strconv"
	"time"
)

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

func getObjectIDs(fromTime time.Time, toTime time.Time) []uint32 {

	counters := GetAllCounterTypes()

	workingDirectory := GetWorkingDirectory()

	indexMap := make(map[uint32]interface{})

	for current := fromTime; !current.After(toTime); current = current.AddDate(0, 0, 1) {

		for counterID := range counters {

			for indexID := range counters {

				path := workingDirectory + "/database/" + current.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(counterID)) + "/index_" + strconv.Itoa(int(indexID)-1) + ".json"

				data, err := os.ReadFile(path)

				if err != nil {

					continue
				}

				if err := json.Unmarshal(data, &indexMap); err != nil {

					continue
				}
			}

		}
	}

	objectIDs := make([]uint32, 0, len(indexMap))

	for objectID := range indexMap {

		objectIDs = append(objectIDs, objectID)
	}

	return objectIDs
}
