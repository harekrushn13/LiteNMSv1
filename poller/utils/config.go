package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

type Interval string

const (
	BatchInterval Interval = "batchInterval"
)

var intervalMapping = map[Interval]int64{}

type DataType uint8

const (
	TypeUint64 DataType = iota + 1

	TypeFloat64

	TypeString
)

type CounterConfig struct {
	Name string `json:"name"`

	Type string `json:"type"`

	Polling int64 `json:"polling"`
}

var counterConfigs = map[uint16]CounterConfig{}

var counterMapping = map[uint16]DataType{}

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

	if err := json.Unmarshal(counterData, &counterConfigs); err != nil {

		return fmt.Errorf("parse counter.json file error: %s", err)
	}

	for key, value := range counterConfigs {

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

	return nil
}

func GetBatchInterval() int64 {

	value, _ := intervalMapping[BatchInterval]

	return value
}

func GetCounterType(counterId uint16) DataType {

	value, _ := counterMapping[counterId]

	return value
}

func GetCounterName(counterId uint16) string {

	if config, exists := counterConfigs[counterId]; exists {

		return config.Name
	}

	return ""
}

func GetCounterPollingInterval(counterId uint16) int64 {

	if config, exists := counterConfigs[counterId]; exists {

		return config.Polling
	}

	return 0
}
