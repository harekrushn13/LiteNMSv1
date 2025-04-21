package main

import (
	. "backend/models"
	. "backend/routes"
	"log"
)

func main() {

	if err := InitDB(); err != nil {

		log.Fatalf("Failed to initialize database: %v", err)
	}

	defer DB.Close()

	router := SetupRoutes(DB)

	log.Println("Starting server on :8080")

	if err := router.Run(":8080"); err != nil {

		log.Fatalf("Failed to start server: %v", err)
	}
}
