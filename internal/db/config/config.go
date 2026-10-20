package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBHost              string
	DBPort              int
	DBUser              string
	DBPassword          string
	DBName              string
	DBSSLMode           string
	JWTSecret           string
	CloudinaryURL       string
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
	RedisAddr           string
	RedisPassword       string
	RedisDB             int
	RedisTTL            time.Duration
	RefreshTokenTTL     time.Duration
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURL   string
}

func Load() (*Config, error) {

	requiredEnvVars := []string{"DB_HOST", "DB_PASSWORD", "DB_USER", "DB_PORT", "DB_NAME", "JWTSecret", "CLOUDINARY_URL", "CLOUDINARY_CLOUD_NAME", "CLOUDINARY_API_KEY", "CLOUDINARY_API_SECRET", "REDIS_ADDR", "REDIS_DB", "REDIS_TTL", "REFRESH_TOKEN_TTL", "GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET", "GOOGLE_REDIRECT_URL"}
	missingVars := checkMissingENV(requiredEnvVars)

	if len(missingVars) > 0 {
		return nil, fmt.Errorf("Missing or empty environement variables: %v\n", missingVars)
	}

	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	redisDB, redErr := strconv.Atoi(os.Getenv("REDIS_DB"))
	redisTTL, redTTLErr := time.ParseDuration(os.Getenv("REDIS_TTL"))
	refreshTokenTTL, refreshTokenTTLErr := time.ParseDuration(os.Getenv("REFRESH_TOKEN_TTL"))

	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	if redErr != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", redErr)
	}
	if redTTLErr != nil {
		return nil, fmt.Errorf("invalid REDIS_TTL: %w", redTTLErr)
	}
	if refreshTokenTTLErr != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_TTL: %w", refreshTokenTTLErr)
	}

	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = os.Getenv("DB_SSL_MODE")
	}
	if sslMode == "" {
		sslMode = "disable"
	}

	cfg := &Config{
		DBHost:              os.Getenv("DB_HOST"),
		DBPort:              port,
		DBUser:              os.Getenv("DB_USER"),
		DBPassword:          os.Getenv("DB_PASSWORD"),
		DBName:              os.Getenv("DB_NAME"),
		DBSSLMode:           sslMode,
		JWTSecret:           os.Getenv("JWTSecret"),
		CloudinaryURL:       os.Getenv("CLOUDINARY_URL"),
		CloudinaryCloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
		CloudinaryAPIKey:    os.Getenv("CLOUDINARY_API_KEY"),
		CloudinaryAPISecret: os.Getenv("CLOUDINARY_API_SECRET"),
		RedisAddr:           os.Getenv("REDIS_ADDR"),
		RedisPassword:       os.Getenv("REDIS_PASSWORD"),
		RedisDB:             redisDB,
		RedisTTL:            redisTTL,
		RefreshTokenTTL:     refreshTokenTTL,
		GoogleClientID:      os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:  os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:   os.Getenv("GOOGLE_REDIRECT_URL"),
	}
	if cfg.JWTSecret == "" {
		return &Config{}, errors.New("JWTSecret is required")
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
