package utils

import (
	"encoding/binary"
	"fmt"
	"math"
	. "reportdb/config"
)

func GetPartition(objectID uint32, partitionCount uint8) uint8 {

	return uint8(objectID % uint32(partitionCount))
}

func EncodeData(row RowData) ([]byte, error) {

	dataType, exists := CounterTypeMapping[row.CounterId]

	if !exists {

		return nil, fmt.Errorf("unknown counter ID: %d", row.CounterId)
	}

	switch dataType {

	case TypeUint64:

		val, ok := row.Value.(uint64)

		if !ok {

			return nil, fmt.Errorf("invalid uint64 value for counter %d", row.CounterId)
		}

		data := make([]byte, 8)

		binary.LittleEndian.PutUint64(data, val)

		return data, nil

	case TypeFloat64:

		val, ok := row.Value.(float64)

		if !ok {

			return nil, fmt.Errorf("invalid float64 value for counter %d", row.CounterId)
		}

		data := make([]byte, 8)

		binary.LittleEndian.PutUint64(data, math.Float64bits(val))

		return data, nil

	case TypeString:

		str, ok := row.Value.(string)

		if !ok {

			return nil, fmt.Errorf("invalid string value for counter %d", row.CounterId)
		}

		data := make([]byte, 4+len(str))

		binary.LittleEndian.PutUint32(data, uint32(len(str)))

		copy(data[4:], []byte(str))

		return data, nil

	default:

		return nil, fmt.Errorf("unsupported data type: %d", dataType)
	}
}

func DecodeData(data []byte, dataType DataType) (interface{}, error) {

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

func LengthOfData(data []byte, offset int64, dataType DataType) (int64, int64, error) {

	switch dataType {

	case TypeUint64, TypeFloat64:

		return offset, offset + 8, nil

	case TypeString:

		return offset + 4, offset + 4 + int64(data[offset]), nil
	}

	return 0, 0, fmt.Errorf("unsupported data type: %d", dataType)

}
