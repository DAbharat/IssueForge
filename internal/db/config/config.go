package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
}

func Load() (*Config, error) {

	requiredEnvVars := []string{"DB_HOST", "DB_PASSWORD", "DB_USER", "DB_PORT", "DB_NAME", "SECRETKEY"}
	missingVars := checkMissingENV(requiredEnvVars)

	if len(missingVars) > 0 {
		return nil, fmt.Errorf("Missing or empty environement variables: %v\n", missingVars)
	}

	port, err := strconv.Atoi(os.Getenv("DB_PORT"))

	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     port,
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		JWTSecret:  os.Getenv("SECRETKEY"),
	}
	if cfg.JWTSecret == "" {
		return &Config{}, errors.New("JWT_SECRET is required")
	}

	return cfg, nil
}

func checkMissingENV(keys []string) []string {
	var missing []string

	for _, key := range keys {
		if val, exists := os.LookupEnv(key); !exists || val == "" {
			missing = append(missing, key)
		}
	}

	return missing
}
