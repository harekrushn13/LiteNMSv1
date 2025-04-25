package reader

import (
	"encoding/binary"
	"math"
	. "reportdb/utils"
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
