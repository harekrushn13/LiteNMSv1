package engine

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reportdb/config"
	"reportdb/src/utils"
	"syscall"
	"time"
)

func Get(counterID uint16, objectID uint32, from uint32, to uint32, day time.Time) ([]interface{}, error, bool) {

	dataType, exists := config.CounterTypeMapping[counterID]

	if !exists {

		return nil, fmt.Errorf("unknown counter ID: %d", counterID), false
	}

	indexPath := utils.GetIndexFilePath(counterID, day)

	indexData, err := getIndex(counterID, indexPath, day, time.Now())

	entry, exists := indexData[objectID]

	if !exists {

		return nil, fmt.Errorf("no data found for objectID %d", objectID), true
	}

	if entry.EndTime < from || entry.StartTime > to {

		return nil, fmt.Errorf("data for objectID %d is not within the requested time range", objectID), true
	}

	validOffsets, err := getValidOffsets(entry, from, to)

	if err != nil {

		return nil, fmt.Errorf("failed to get valid offsets: %v", err), false
	}

	dayResults, err := readDayData(counterID, objectID, day, validOffsets, dataType)

	if err != nil {

		return nil, fmt.Errorf("failed to read day data: %v", err), false
	}

	return dayResults, nil, false
}

func readDayData(counterID uint16, objectID uint32, day time.Time, offsets []int64, dataType config.DataType) ([]interface{}, error) {

	partition := utils.GetPartition(objectID)

	dataPath := utils.GetPartitionFilePath(counterID, partition, day)

	file, err := os.Open(dataPath)

	if err != nil {

		return nil, err
	}

	defer file.Close()

	fileInfo, err := file.Stat()

	if err != nil {

		return nil, err
	}

	data, err := syscall.Mmap(int(file.Fd()), 0, int(fileInfo.Size()), syscall.PROT_READ, syscall.MAP_SHARED)

	if err != nil {

		return nil, err
	}

	defer syscall.Munmap(data)

	var results []interface{}

	for _, offset := range offsets {

		var value interface{}

		var err error

		switch dataType {

		case config.TypeUint64:

			if offset+8 > int64(len(data)) {

				return nil, fmt.Errorf("offset %d out of bounds (file size: %d)", offset, len(data))
			}

			value = binary.LittleEndian.Uint64(data[offset : offset+8])

		case config.TypeFloat64:

			if offset+8 > int64(len(data)) {

				return nil, fmt.Errorf("offset %d out of bounds (file size: %d)", offset, len(data))
			}

			value = math.Float64frombits(binary.LittleEndian.Uint64(data[offset : offset+8]))

		case config.TypeString:

			if offset+4 > int64(len(data)) {

				return nil, fmt.Errorf("offset %d out of bounds (file size: %d)", offset, len(data))
			}

			length := int(data[offset])

			if offset+4+int64(length) > int64(len(data)) {

				return nil, fmt.Errorf("string data at offset %d exceeds file bounds", offset)
			}

			value = string(data[offset+4 : offset+4+int64(length)])
		}

		if err != nil {

			return nil, err
		}

		results = append(results, value)

	}

	return results, nil
}

func getIndex(counterID uint16, indexPath string, current time.Time, day time.Time) (map[uint32]*config.IndexEntry, error) {

	config.InitIndexHandles.Do(utils.InitIndexHandlesFunc)

	var handle *config.IndexHandle

	cY, cM, cD := current.Date()

	dY, dM, dD := day.Date()

	if cY == dY && cM == dM && cD == dD {

		handle = config.IndexHandles[counterID]

		if handle != nil {

			handle.Lock.RLock()
		}
	}

	data, err := os.ReadFile(indexPath)

	if handle != nil {

		handle.Lock.RUnlock()
	}

	if err != nil {

		return nil, err
	}

	var indexData map[uint32]*config.IndexEntry

	if err := json.Unmarshal(data, &indexData); err != nil {

		return nil, err
	}

	return indexData, nil
}

func getValidOffsets(entry *config.IndexEntry, from uint32, to uint32) ([]int64, error) {

	var validOffsets []int64

	length := uint8(len(entry.Offsets))

	var start uint8

	var end uint8

	if from <= entry.StartTime {

		start = uint8(0)

	} else {

		start = uint8(from - entry.StartTime)
	}

	if to >= entry.EndTime {

		end = length

	} else {

		end = length - uint8(entry.EndTime-to)
	}

	validOffsets = append(validOffsets, entry.Offsets[start:end]...)

	return validOffsets, nil
}
