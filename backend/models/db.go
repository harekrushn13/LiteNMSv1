package models

import (
	"backend/config"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func InitDB() error {

	dbConfig := config.LoadDBConfig()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode)

	var err error

	DB, err = sqlx.Connect("postgres", connStr)

	if err != nil {

		return fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := DB.Ping(); err != nil {

		return fmt.Errorf("failed to ping database: %v", err)
	}

	return nil
}
