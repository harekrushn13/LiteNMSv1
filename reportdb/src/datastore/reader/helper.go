package reader

import (
	"encoding/binary"
	"math"
	. "reportdb/utils"
)

func decodeData(data [][]byte, dataType DataType, result *[]interface{}) {

	for _, row := range data {

		switch dataType {

		case TypeUint64:

			*result = append(*result, binary.LittleEndian.Uint64(row))

		case TypeFloat64:

			*result = append(*result, math.Float64frombits(binary.LittleEndian.Uint64(row)))

		case TypeString:

			*result = append(*result, string(row))

		}
	}
}
