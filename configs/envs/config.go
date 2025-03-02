package envs

import (
	"fmt"
	"os"
)

type Config struct {
	// Server configuration
	Port string

	// Database configuration
	DB struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
		URL      string
	}

	// Redis configuration
	Redis struct {
		Addr     string
		Password string
	}

	// Logging configuration
	LogLevel string

	// Kafka configuration
	Kafka struct {
		Brokers           []string
		TransactionsTopic string
	}

	// Worker configuration
	Workers struct {
		Count int
	}

	Retry struct {
		MaxRetries int
	}

	// Security configuration
	Security struct {
		EncryptionKey string
	}
}

func Load() *Config {
	cfg := &Config{}

	// Server configuration
	cfg.Port = getEnv("PORT", "8080")

	// Database configuration
	cfg.DB.Host = getEnv("DB_HOST", "localhost")
	cfg.DB.Port = getEnv("DB_PORT", "5432")
	cfg.DB.User = getEnv("DB_USER", "user")
	cfg.DB.Password = getEnv("DB_PASSWORD", "password")
	cfg.DB.Name = getEnv("DB_NAME", "payments")
	cfg.DB.URL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
	// Redis configuration
	cfg.Redis.Addr = getEnv("REDIS_ADDR", "redis:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "password")

	// Logging configuration
	cfg.LogLevel = getEnv("LOG_LEVEL", "info")

	// Kafka configuration
	kafkaBroker := getEnv("KAFKA_BROKER", "kafka:9092")
	cfg.Kafka.Brokers = []string{kafkaBroker}
	cfg.Kafka.TransactionsTopic = getEnv("KAFKA_TRANSACTIONS_TOPIC", "payment-transactions")

	// Worker configuration
	cfg.Workers.Count = 5

	cfg.Retry.MaxRetries = 3

	return cfg
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
