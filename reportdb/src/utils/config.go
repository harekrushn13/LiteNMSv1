package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

type Config struct {
	Writers int `json:"writers"`

	Readers int `json:"readers"`

	Partitions int `json:"partitions"`

	DataBuffer int `json:"dataBuffer"`

	ResponseBuffer int `json:"responseBuffer"`

	EventsBuffer int `json:"eventsBuffer"`

	QueryBuffer int `json:"queryBuffer"`

	ObjectWorkers int `json:"objectWorkers"`

	FileGrowthSize int `json:"fileGrowthSize"`

	SaveIndexInterval int `json:"saveIndexInterval"`

	QueryTimeout int `json:"queryTimeout"`
}

type DataType uint8

const (
	TypeUint64 DataType = iota + 1

	TypeFloat64

	TypeString
)

type CounterConfig struct {
	Name string `json:"name"`

	Type string `json:"type"`
}

var (
	appConfig Config

	counterTypes = map[uint16]DataType{}

	workingDir string
)

func InitConfig() error {

	currentPath, err := os.Getwd()

	if err != nil {

		return fmt.Errorf("get current path err: %v", err)
	}

	workingDir = filepath.Dir(currentPath) // ./reportdb

	configPath := workingDir + "/config/config.json"

	configData, err := os.ReadFile(configPath)

	if err != nil {

		return fmt.Errorf("read timer.json file error: %s", err)
	}

	if err := json.Unmarshal(configData, &appConfig); err != nil {

		return fmt.Errorf("parse timer.json file error: %s", err)
	}

	counterPath := workingDir + "/config/counter.json"

	counterData, err := os.ReadFile(counterPath)

	if err != nil {

		return fmt.Errorf("read counter.json file error: %s", err)
	}

	tempCounterMapping := map[uint16]CounterConfig{}

	if err := json.Unmarshal(counterData, &tempCounterMapping); err != nil {

		return fmt.Errorf("parse counter.json file error: %s", err)
	}

	for key, value := range tempCounterMapping {

		switch value.Type {

		case "uint64":

			counterTypes[key] = TypeUint64

		case "float64":

			counterTypes[key] = TypeFloat64

		case "string":
			counterTypes[key] = TypeString

		default:

			return fmt.Errorf("unknown counter type %s", value.Type)
		}

	}

	return nil
}

func GetWorkingDirectory() string {

	return workingDir
}

func GetWriters() int {

	return appConfig.Writers
}

func GetReaders() int {

	return appConfig.Readers
}

func GetPartitions() int {

	return appConfig.Partitions
}

func GetDataBuffer() int {

	return appConfig.DataBuffer
}

func GetResponseBuffer() int {

	return appConfig.ResponseBuffer
}

func GetEventsBuffer() int {

	return appConfig.EventsBuffer
}

func GetQueryBuffer() int {

	return appConfig.QueryBuffer
}

func GetObjectWorkers() int {

	return appConfig.ObjectWorkers
}

func GetFileGrowthSize() int {

	return appConfig.FileGrowthSize
}

func GetSaveIndexInterval() int {

	return appConfig.SaveIndexInterval
}

func GetCounterType(counterId uint16) (DataType, error) {

	dataType, ok := counterTypes[counterId]

	if !ok {

		return 0, fmt.Errorf("counter ID %d not found", counterId)
	}

	return dataType, nil
}

func GetAllCounterTypes() map[uint16]DataType {

	return counterTypes
}

func GetQueryTimeout() int {

	return appConfig.QueryTimeout
}

func SysTotalMemory() uint64 {

	in := &syscall.Sysinfo_t{}

	err := syscall.Sysinfo(in)

	if err != nil {

		return 0
	}

	// If this is a 32-bit system, then these fields are
	// uint32 instead of uint64.
	// So we always convert to uint64 to match signature.

	return uint64(in.Totalram) * uint64(in.Unit)
}
