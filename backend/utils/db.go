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

		Host: getEnv("DB_HOST", "localhost"),

		Port: getEnv("DB_PORT", "5432"),

		User: getEnv("DB_USER", "postgres"),

		Password: getEnv("DB_PASSWORD", "postgres"),

		DBName: getEnv("DB_NAME", "LiteNMS"),

		SSLMode: getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, defaultValue string) string {

	if value, exists := os.LookupEnv(key); exists {

		return value

	}

	return defaultValue
}
