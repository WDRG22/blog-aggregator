package config

import (
	"os"
	"log"
	"strconv"
	"time"
)


type Config struct {
	Port		string
	DatabaseURL	string
	WorkerCount	int
	WorkerInterval	time.Duration
}

// Read configuration from environment variables and return Config struct
func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == ""{ // Default to 8080
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" { // Default to local db value
		dbURL = "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"
		log.Println("WARNING: DATABASE_URL not set, using default local value.")
	}

	workerCountStr := os.Getenv("WORKER_COUNT")
	if workerCountStr == "" { // Default worker count
		workerCountStr = "10"
	}
	workerCount, err := strconv.Atoi(workerCountStr)
	if err != nil {
		return nil, err
	}

	intervalStr := os.Getenv("WORKER_INTERVAL_SECONDS")
	if intervalStr == "" {
		intervalStr = "60" // Default
	}
	intervalSeconds, err := strconv.Atoi(intervalStr)
	if err != nil {
		return nil, err
	}

	// Conver ns to sec
	workerInterval := time.Duration(intervalSeconds) * time.Second

	return &Config{
		Port:		port,
		DatabaseURL:	dbURL,
		WorkerCount:	workerCount,
		WorkerInterval:	workerInterval,
	}, nil
}
