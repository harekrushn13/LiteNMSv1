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

		newvalue := uint64(val)

		binary.LittleEndian.PutUint32(*data, 8)

		binary.LittleEndian.PutUint64((*data)[4:], newvalue)

		return 12, nil

	case TypeFloat64:

		val, ok := row.Value.(float64)

		if !ok {

			return 0, fmt.Errorf("encodeData : invalid float64 value for counter %d", row.CounterId)
		}

		binary.LittleEndian.PutUint32(*data, 8)

		binary.LittleEndian.PutUint64((*data)[4:], math.Float64bits(val))

		return 12, nil

	case TypeString:

		str, ok := row.Value.(string)

		if !ok {

			return 0, fmt.Errorf("encodeData : invalid string value for counter %d", row.CounterId)
		}

		binary.LittleEndian.PutUint32(*data, uint32(len(str)))

		copy((*data)[4:], str)

		return uint8(4 + len(str)), nil

	default:

		return 0, fmt.Errorf("encodeData : unsupported data type: %d", dataType)
	}
}
