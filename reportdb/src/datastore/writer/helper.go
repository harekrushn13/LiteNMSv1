package writer

import (
	"encoding/binary"
	"fmt"
	"math"
	. "reportdb/utils"
	"strconv"
	"time"
)

func getPath(workingDirectory string, row Events) string {

	day := time.Unix(int64(row.Timestamp), 0).Truncate(24 * time.Hour).UTC()

	return workingDirectory + "/database/" + day.Format("2006/01/02") + "/counter_" + strconv.Itoa(int(row.CounterId))
}

func encodeData(row Events, data *[]byte) (uint8, error) {

	dataType, err := GetCounterType(row.CounterId)

	if err != nil {

		return 0, fmt.Errorf("encodeData : Error getting counter type: %v", err)
	}

	switch dataType {

	case TypeUint64:

		val, ok := row.Value.(float64)

		if !ok {

			return 0, fmt.Errorf("encodeData : invalid uint64 value for counter %d", row.CounterId)
		}

		binary.LittleEndian.PutUint32(*data, 8)

		binary.LittleEndian.PutUint32((*data)[4:], row.Timestamp)

		binary.LittleEndian.PutUint64((*data)[8:], uint64(val))

		return 16, nil

	case TypeFloat64:

		val, ok := row.Value.(float64)

		if !ok {

			return 0, fmt.Errorf("encodeData : invalid float64 value for counter %d", row.CounterId)
		}

		binary.LittleEndian.PutUint32(*data, 8)

		binary.LittleEndian.PutUint32((*data)[4:], row.Timestamp)

		binary.LittleEndian.PutUint64((*data)[8:], math.Float64bits(val))

		return 16, nil

	case TypeString:

		str, ok := row.Value.(string)

		if !ok {

			return 0, fmt.Errorf("encodeData : invalid string value for counter %d", row.CounterId)
		}

		binary.LittleEndian.PutUint32(*data, uint32(len(str)))

		binary.LittleEndian.PutUint32((*data)[4:], row.Timestamp)

		copy((*data)[8:], str)

		return uint8(8 + len(str)), nil

	default:

		return 0, fmt.Errorf("encodeData : unsupported data type: %d", dataType)
	}
}
