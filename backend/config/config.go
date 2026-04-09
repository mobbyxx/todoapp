package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Server     ServerConfig
	FCM        FCMConfig
	LogLevel   string
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	URL string
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	URL string
}

// JWTConfig holds JWT settings
type JWTConfig struct {
	Secret         string
	SecretPrevious string
	RotationWindow int
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Domain string
	Port   int
}

// FCMConfig holds Firebase Cloud Messaging settings
type FCMConfig struct {
	APIKey string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			URL: getEnv("DB_URL", "postgres://user:password@localhost:5432/todo?sslmode=disable"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379"),
		},
		JWT: JWTConfig{
			Secret:         getEnv("JWT_SECRET", "your-secret-key"),
			SecretPrevious: getEnv("JWT_SECRET_PREVIOUS", ""),
			RotationWindow: getEnvAsInt("JWT_ROTATION_WINDOW", 86400),
		},
		Server: ServerConfig{
			Domain: getEnv("API_DOMAIN", "localhost"),
			Port:   getEnvAsInt("PORT", 8080),
		},
		FCM: FCMConfig{
			APIKey: getEnv("FCM_API_KEY", ""),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
