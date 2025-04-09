package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

type ConfigType = string

const (
	Writers ConfigType = "writers"

	Readers ConfigType = "readers"

	Objects ConfigType = "objects"

	Counters ConfigType = "counters"

	Partitions ConfigType = "partitions"

	WriterTaskQueueBuffer ConfigType = "writerTaskQueueBuffer"

	FileGrowthSize ConfigType = "fileGrowthSize"
)

var ProjectPath ConfigType

var configMapping = map[string]int{}

type DataType uint8

const (
	TypeUint64 DataType = iota + 1

	TypeFloat64

	TypeString
)

var counterMapping = map[uint16]DataType{}

type Interval string

const (
	PollingInterval Interval = "pollingInterval"

	BatchInterval Interval = "batchInterval"

	StopTime Interval = "stopTime"
)

var pollerMapping = map[Interval]int64{}

type RowData struct {
	ObjectId uint32

	CounterId uint16

	Timestamp uint32

	Value interface{}
}

func InitConfig() error {

	currentPath, err := os.Getwd()

	if err != nil {

		return fmt.Errorf("get current path err: %v", err)
	}

	ProjectPath = currentPath // ./reportdb

	configPath := currentPath + "/config/config.json"

	configData, err := os.ReadFile(configPath)

	if err != nil {

		return fmt.Errorf("read config.json file error: %s", err)
	}

	if err := json.Unmarshal(configData, &configMapping); err != nil {

		return fmt.Errorf("parse config.json file error: %s", err)
	}

	counterPath := currentPath + "/config/counter.json"

	counterData, err := os.ReadFile(counterPath)

	if err != nil {

		return fmt.Errorf("read counter.json file error: %s", err)
	}

	tempCounterMapping := map[uint16]struct {
		Name string `json:"name"`

		Type string `json:"type"`
	}{}

	if err := json.Unmarshal(counterData, &tempCounterMapping); err != nil {

		return fmt.Errorf("parse counter.json file error: %s", err)
	}

	for key, value := range tempCounterMapping {

		switch value.Type {

		case "uint64":

			counterMapping[key] = TypeUint64

		case "float64":

			counterMapping[key] = TypeFloat64

		case "string":
			counterMapping[key] = TypeString

		default:

			return fmt.Errorf("unknown counter type %s", value.Type)
		}

	}

	tempCounterMapping = nil

	pollerPath := currentPath + "/config/poller.json"

	pollerData, err := os.ReadFile(pollerPath)

	if err != nil {

		return fmt.Errorf("read poller.json file error: %s", err)
	}

	if err := json.Unmarshal(pollerData, &pollerMapping); err != nil {

		return fmt.Errorf("parse poller.json file error: %s", err)
	}

	return nil
}

func GetProjectPath() ConfigType {

	return ProjectPath
}

func GetWriters() uint8 {

	value, _ := configMapping[Writers]

	return uint8(value)
}

func GetReaders() uint8 {

	value, _ := configMapping[Readers]

	return uint8(value)
}

func GetObjects() uint32 {

	value, _ := configMapping[Objects]

	return uint32(value)
}

func GetCounters() uint16 {

	value, _ := configMapping[Counters]

	return uint16(value)
}

func GetPartitions() uint8 {

	value, _ := configMapping[Partitions]

	return uint8(value)
}

func GetWriterTaskQueueBuffer() uint16 {

	value, _ := configMapping[WriterTaskQueueBuffer]

	return uint16(value)
}

func GetFileGrowthSize() int64 {

	value, _ := configMapping[FileGrowthSize]

	return int64(value)
}

func GetCounterType(counterId uint16) DataType {

	value, _ := counterMapping[counterId]

	return value
}

func GetPollingInterval() int64 {

	value, _ := pollerMapping[PollingInterval]

	return value
}

func GetBatchInterval() int64 {

	value, _ := pollerMapping[BatchInterval]

	return value
}

func GetStopTime() int64 {

	value, _ := pollerMapping[StopTime]

	return value
}
