package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

type Interval string

const (
	PollingInterval Interval = "pollingInterval"

	BatchInterval Interval = "batchInterval"
)

var intervalMapping = map[Interval]int64{}

type DataType uint8

const (
	TypeUint64 DataType = iota + 1

	TypeFloat64

	TypeString
)

type ConfigType = string

var counterMapping = map[uint16]DataType{}

const (
	Objects ConfigType = "objects"

	Counters ConfigType = "counters"
)

var configMapping = map[string]int{}

func InitConfig() error {

	currentPath, err := os.Getwd()

	if err != nil {

		return fmt.Errorf("get current path err: %v", err)
	}

	timerPath := currentPath + "/config/timer.json"

	timerData, err := os.ReadFile(timerPath)

	if err != nil {

		return fmt.Errorf("read timer.json file error: %s", err)
	}

	if err := json.Unmarshal(timerData, &intervalMapping); err != nil {

		return fmt.Errorf("parse timer.json file error: %s", err)
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

	configPath := currentPath + "/config/config.json"

	configData, err := os.ReadFile(configPath)

	if err != nil {

		return fmt.Errorf("read config.json file error: %s", err)
	}

	if err := json.Unmarshal(configData, &configMapping); err != nil {

		return fmt.Errorf("parse config.json file error: %s", err)
	}

	return nil
}

func GetPollingInterval() int64 {

	value, _ := intervalMapping[PollingInterval]

	return value
}

func GetBatchInterval() int64 {

	value, _ := intervalMapping[BatchInterval]

	return value
}

func GetCounterType(counterId uint16) DataType {

	value, _ := counterMapping[counterId]

	return value
}

func GetObjects() uint32 {

	value, _ := configMapping[Objects]

	return uint32(value)
}

func GetCounters() uint16 {

	value, _ := configMapping[Counters]

	return uint16(value)
}
