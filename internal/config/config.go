package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTP HTTPConfig
	DB   DBConfig
}

type HTTPConfig struct {
	Port string
}

type DBConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
	SSLMode  string
}

func Load() (*Config, error) {
	// Load .env file (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env file not found, using environment variables: %v", err)
	}

	cfg := &Config{
		HTTP: HTTPConfig{
			Port: getenv("APP_HTTP_PORT", "8080"),
		},
		DB: DBConfig{
			Host:     getenv("POSTGRES_HOST", "localhost"),
			Port:     getenv("POSTGRES_PORT", "5432"),
			Database: getenv("POSTGRES_DB", "players"),
			User:     getenv("POSTGRES_USER", "players"),
			Password: getenv("POSTGRES_PASSWORD", "players"),
			SSLMode:  getenv("POSTGRES_SSLMODE", "disable"),
		},
	}

	return cfg, nil
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
