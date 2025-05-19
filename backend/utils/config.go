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

	DbHost string `json:"dbHost"`

	DbPort string `json:"dbPort"`

	DbUser string `json:"dbUser"`

	DbPassword string `json:"dbPassword"`

	DbName string `json:"dbName"`

	DbSSLMode string `json:"dbSSLMode"`
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

func GetDBHost() string {

	return config.DbHost
}

func GetDBPort() string {

	return config.DbPort
}

func GetDBUser() string {

	return config.DbUser
}

func GetDBPassword() string {

	return config.DbPassword
}

func GetDBName() string {

	return config.DbName
}

func GetDBSSLMode() string {

	return config.DbSSLMode
}
