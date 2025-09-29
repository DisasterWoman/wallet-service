package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost         string
	DBPort         int
	DBExternalPort int
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	
	ServerPort     int
	ServerHost     string
	
	LogLevel       string
}

func Load() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	cfg := &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnvAsInt("DB_PORT", 5432),
		DBExternalPort: getEnvAsInt("DB_EXTERNAL_PORT", 5433),
		DBUser:         getEnv("DB_USER", "wallet_user"),
		DBPassword:     getEnv("DB_PASSWORD", "wallet_password"),
		DBName:         getEnv("DB_NAME", "wallet_db"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		
		ServerPort:     getEnvAsInt("SERVER_PORT", 8080),
		ServerHost:     getEnv("SERVER_HOST", "0.0.0.0"),
		
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	log.Printf("Config loaded: DB_HOST=%s, DB_PORT=%d, SERVER_PORT=%d", 
		cfg.DBHost, cfg.DBPort, cfg.ServerPort)

	return cfg, nil
}
func (c *Config) GetDBConnectionString() string {
	if isRunningInDocker() {
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
		)
	}
	
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		"localhost", c.DBExternalPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}

func (c *Config) validate() error {
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST cannot be empty")
	}
	
	if c.DBPort <= 0 || c.DBPort > 65535 {
		return fmt.Errorf("DB_PORT must be between 1 and 65535")
	}
	
	if c.DBExternalPort <= 0 || c.DBExternalPort > 65535 {
		return fmt.Errorf("DB_EXTERNAL_PORT must be between 1 and 65535")
	}
	
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER cannot be empty")
	}
	
	if c.DBPassword == "" {
		return fmt.Errorf("DB_PASSWORD cannot be empty")
	}
	
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME cannot be empty")
	}
	
	if c.ServerPort <= 0 || c.ServerPort > 65535 {
		return fmt.Errorf("SERVER_PORT must be between 1 and 65535")
	}
	
	if c.ServerHost == "" {
		return fmt.Errorf("SERVER_HOST cannot be empty")
	}
	
	return nil
}

func isRunningInDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
		if file, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		if contains(string(file), "docker") {
			return true
		}
	}
		if os.Getenv("DOCKER_CONTAINER") == "true" {
		return true
	}
	
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		log.Printf("Warning: Invalid boolean value for %s: %s, using default: %t", key, value, defaultValue)
	}
	return defaultValue
}