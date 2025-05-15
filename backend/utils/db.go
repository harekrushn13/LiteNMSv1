package utils

import "os"

type DBConfig struct {
	Host string

	Port string

	User string

	Password string

	DBName string

	SSLMode string
}

func LoadDBConfig() DBConfig {

	return DBConfig{

		Host: getEnv("DB_HOST", GetDBHost()),

		Port: getEnv("DB_PORT", GetDBPort()),

		User: getEnv("DB_USER", GetDBUser()),

		Password: getEnv("DB_PASSWORD", GetDBPassword()),

		DBName: getEnv("DB_NAME", GetDBName()),

		SSLMode: getEnv("DB_SSLMODE", GetDBSSLMode()),
	}
}

func getEnv(key, defaultValue string) string {

	if value, exists := os.LookupEnv(key); exists {

		return value

	}

	return defaultValue
}
