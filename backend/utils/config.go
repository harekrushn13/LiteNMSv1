package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	DeviceBuffer int `json:"deviceBuffer"`

	DataBuffer int `json:"dataBuffer"`

	QueryBuffer int `json:"queryBuffer"`
}

var config Configuration

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

	return nil
}

func GetDeviceBuffer() int {

	return config.DeviceBuffer
}

func GetDataBuffer() int {

	return config.DataBuffer
}

func GetQueryBuffer() int {

	return config.QueryBuffer
}
