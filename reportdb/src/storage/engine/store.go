package engine

import (
	"encoding/binary"
	"fmt"
	"math"
	"reportdb/config"
	utils2 "reportdb/src/utils"
	"syscall"
)

func Save(row config.RowData) error {

	partition := utils2.GetPartition(row.ObjectId)

	handle, err := utils2.GetDataFile(row.CounterId, partition)

	if err != nil {

		return err
	}

	handle.Lock.Lock()

	defer handle.Lock.Unlock()

	dataType, exists := config.CounterTypeMapping[row.CounterId]

	if !exists {

		return fmt.Errorf("unknown counter ID: %d", row.CounterId)
	}

	var requiredSize int64

	switch dataType {

	case config.TypeUint64, config.TypeFloat64:

		requiredSize = 8

	case config.TypeString:

		str, ok := row.Value.(string)

		if !ok {

			return fmt.Errorf("invalid string value for counter %d", row.CounterId)
		}

		requiredSize = 4 + int64(len(str))
	}

	if handle.Offset+requiredSize > handle.AvailableSize {

		if handle.MmapData != nil {

			if err := syscall.Munmap(handle.MmapData); err != nil {

				return fmt.Errorf("munmap failed: %v", err)
			}
		}

		handle.AvailableSize = handle.Offset + config.FileGrowth // just 64 byte adding

		if err := handle.File.Truncate(handle.AvailableSize); err != nil {

			return fmt.Errorf("failed to grow file: %v", err)
		}

		data, err := syscall.Mmap(int(handle.File.Fd()), 0, int(handle.AvailableSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

		if err != nil {

			return fmt.Errorf("mmap failed: %v", err)
		}

		handle.MmapData = data

	}

	lastOffset := handle.Offset

	switch dataType {

	case config.TypeUint64:

		val, ok := row.Value.(uint64)

		if !ok {

			return fmt.Errorf("invalid uint64 value for counter %d", row.CounterId)
		}

		binary.LittleEndian.PutUint64(handle.MmapData[handle.Offset:], val)

		handle.Offset += 8

	case config.TypeFloat64:

		val, ok := row.Value.(float64)

		if !ok {

			return fmt.Errorf("invalid float64 value for counter %d", row.CounterId)
		}

		binary.LittleEndian.PutUint64(handle.MmapData[handle.Offset:], math.Float64bits(val))

		handle.Offset += 8

	case config.TypeString:

		str, ok := row.Value.(string)

		if !ok {

			return fmt.Errorf("invalid string value for counter %d", row.CounterId)
		}

		length := uint32(len(str))

		binary.LittleEndian.PutUint32(handle.MmapData[handle.Offset:], length)

		copy(handle.MmapData[handle.Offset+4:], []byte(str))

		handle.Offset += 4 + int64(length)
	}

	utils2.UpdateIndex(row.CounterId, row.ObjectId, lastOffset, row.Timestamp)

	return nil
}
