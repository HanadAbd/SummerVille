package services

import (
	"os"
	"time"
)

type Config struct {
	Environment      string
	ServerAddress    string
	RefreshInterval  time.Duration
	KafkaBroker      string
	KafkaTopic       string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string
	MSSQLServer      string
	MSSQLDatabase    string
}

func LoadConfig() *Config {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	cfg := &Config{
		Environment:      env,
		ServerAddress:    "localhost:8080",
		KafkaBroker:      os.Getenv("KAFKA_BROKER"),
		KafkaTopic:       os.Getenv("KAFKA_TOPIC"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),
		MSSQLServer:      os.Getenv("MSSQL_SERVER"),
		MSSQLDatabase:    os.Getenv("MSSQL_DATABASE"),
	}
	if val := os.Getenv("SERVER_ADDR"); val != "" {
		cfg.ServerAddress = val
	}

	return cfg
}
