package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ConfigType = string

const (
	writers ConfigType = "writers"

	readers ConfigType = "readers"

	partitions ConfigType = "partitions"

	eventsBuffer ConfigType = "eventsBuffer"

	queryWorkers ConfigType = "queryWorkers"

	queryBuffer ConfigType = "queryBuffer"

	dayWorkers ConfigType = "dayWorkers"

	writerEventBuffer ConfigType = "writerEventBuffer"

	fileGrowthSize ConfigType = "fileGrowthSize"
)

var WorkingDirectory ConfigType

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
	SaveIndexInterval Interval = "saveIndexInterval"
)

var intervalMapping = map[Interval]int64{}

type Events struct {
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

	WorkingDirectory = filepath.Dir(currentPath) // ./reportdb

	// Read config.json

	configPath := WorkingDirectory + "/config/config.json"

	configData, err := os.ReadFile(configPath)

	if err != nil {

		return fmt.Errorf("read config.json file error: %s", err)
	}

	if err := json.Unmarshal(configData, &configMapping); err != nil {

		return fmt.Errorf("parse config.json file error: %s", err)
	}

	// Read counter.json

	counterPath := WorkingDirectory + "/config/counter.json"

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

	// Read timer.go

	timerPath := WorkingDirectory + "/config/timer.json"

	timerData, err := os.ReadFile(timerPath)

	if err != nil {

		return fmt.Errorf("read timer.json file error: %s", err)
	}

	if err := json.Unmarshal(timerData, &intervalMapping); err != nil {

		return fmt.Errorf("parse timer.json file error: %s", err)
	}

	return nil
}

func GetWorkingDirectory() (ConfigType, error) {

	if WorkingDirectory == "" {

		return "", fmt.Errorf("InitConfig : working directory is empty")
	}

	return WorkingDirectory, nil
}

func GetWriters() (uint8, error) {

	value, ok := configMapping[writers]

	if !ok {

		return 0, fmt.Errorf("InitConfig : writer not found")
	}

	return uint8(value), nil
}

func GetReaders() (uint8, error) {

	value, ok := configMapping[readers]

	if !ok {

		return 0, fmt.Errorf("InitConfig : reader not found")
	}

	return uint8(value), nil
}

func GetPartitions() (uint8, error) {

	value, ok := configMapping[partitions]

	if !ok {

		return 0, fmt.Errorf("InitConfig : partitions not found")
	}

	return uint8(value), nil
}

func GetEventsBuffer() (uint8, error) {

	value, ok := configMapping[eventsBuffer]

	if !ok {

		return 0, fmt.Errorf("InitConfig : events buffer not found")
	}

	return uint8(value), nil
}

func GetQueryWorkers() (uint8, error) {

	value, ok := configMapping[queryWorkers]

	if !ok {

		return 0, fmt.Errorf("InitConfig : queryWorkers not found")
	}

	return uint8(value), nil
}

func GetQueryBuffer() (uint8, error) {

	value, ok := configMapping[queryBuffer]

	if !ok {

		return 0, fmt.Errorf("InitConfig : queryBuffer not found")
	}

	return uint8(value), nil
}

func GetDayWorkers() (uint8, error) {

	value, ok := configMapping[dayWorkers]

	if !ok {

		return 0, fmt.Errorf("InitConfig : dayWorkers not found")
	}

	return uint8(value), nil
}

func GetWriterEventBuffer() (uint16, error) {

	value, ok := configMapping[writerEventBuffer]

	if !ok {

		return 0, fmt.Errorf("InitConfig : writer task queue buffer not found")
	}

	return uint16(value), nil
}

func GetFileGrowthSize() (int64, error) {

	value, ok := configMapping[fileGrowthSize]

	if !ok {

		return 0, fmt.Errorf("InitConfig : file growth size not found")
	}

	return int64(value), nil
}

func GetCounterType(counterId uint16) (DataType, error) {

	value, ok := counterMapping[counterId]

	if !ok {

		return 0, fmt.Errorf("InitConfig : counter id %d not found", counterId)
	}

	return value, nil
}

func GetSaveIndexInterval() (int64, error) {

	value, ok := intervalMapping[SaveIndexInterval]

	if !ok {

		return 0, fmt.Errorf("InitConfig : saveIndexInterval not found")
	}

	return value, nil
}
