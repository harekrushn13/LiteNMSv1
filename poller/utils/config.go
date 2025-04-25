package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	DeviceBuffer int `json:"deviceBuffer"`

	DataBuffer int `json:"dataBuffer"`

	Workers int `json:"workers"`

	EventBuffer int `json:"eventBuffer"`

	BatchInterval int `json:"batchInterval"`
}

var config Configuration

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

	configPath := currentPath + "/config/config.json"

	configData, err := os.ReadFile(configPath)

	if err != nil {

		return fmt.Errorf("read config.json file error: %s", err)
	}

	if err := json.Unmarshal(configData, &config); err != nil {

		return fmt.Errorf("parse config.json file error: %s", err)
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

func GetDeviceBuffer() int {

	return config.DeviceBuffer
}

func GetDataBuffer() int {

	return config.DataBuffer
}

func GetWorkerCount() int {

	return config.Workers
}

func GetEventBuffer() int {

	return config.EventBuffer
}

func GetBatchInterval() int {

	return config.BatchInterval
}

func GetCounterType(counterId uint16) DataType {

	return counterMapping[counterId]
}

func GetCounterPollingInterval(counterId uint16) int64 {

	if config, exists := counterConfigs[counterId]; exists {

		return config.Polling
	}

	return 0
}

func GetAllCounters() map[uint16]CounterConfig {

	return counterConfigs
}
