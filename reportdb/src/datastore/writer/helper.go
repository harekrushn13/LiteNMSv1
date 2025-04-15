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

func encodeData(row Events) ([]byte, error) {

	dataType, err := GetCounterType(row.CounterId)

	if err != nil {

		return nil, fmt.Errorf("encodeData : Error getting counter type: %v", err)
	}

	switch dataType {

	case TypeUint64:

		val, ok := row.Value.(float64)

		if !ok {

			return nil, fmt.Errorf("encodeData : invalid uint64 value for counter %d", row.CounterId)
		}

		newvalue := uint64(val)

		data := make([]byte, 12)

		binary.LittleEndian.PutUint32(data, 8)

		binary.LittleEndian.PutUint64(data[4:], newvalue)

		return data, nil

	case TypeFloat64:

		val, ok := row.Value.(float64)

		if !ok {

			return nil, fmt.Errorf("encodeData : invalid float64 value for counter %d", row.CounterId)
		}

		data := make([]byte, 12)

		binary.LittleEndian.PutUint32(data, 8)

		binary.LittleEndian.PutUint64(data[4:], math.Float64bits(val))

		return data, nil

	case TypeString:

		str, ok := row.Value.(string)

		if !ok {

			return nil, fmt.Errorf("encodeData : invalid string value for counter %d", row.CounterId)
		}

		data := make([]byte, 4+len(str))

		binary.LittleEndian.PutUint32(data, uint32(len(str)))

		copy(data[4:], str)

		return data, nil

	default:

		return nil, fmt.Errorf("encodeData : unsupported data type: %d", dataType)
	}
}
