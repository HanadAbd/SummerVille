package util

import (
	"fmt"
	"os"
)

type Config struct {
	Environment   string
	ServerAddress string
	Location      string

	ProdDBUser     string
	ProdDBPassword string
	ProdDBHost     string
	ProdDBPort     string
	ProdDBSSLMode  string
	ProdDBName     string

	Port string
}

func LoadConfig() (*Config, error) {
	env, found := os.LookupEnv("APP_ENV")
	if !found {
		env = "dev"
	}

	cfg := &Config{
		Environment:    env,
		Location:       os.Getenv("APP_LOCATAION"),
		ServerAddress:  os.Getenv("SERVER_ADDR"),
		ProdDBUser:     os.Getenv("DB_USER"),
		ProdDBPassword: os.Getenv("DB_PASSWORD"),
		ProdDBHost:     os.Getenv("DB_HOST"),
		ProdDBPort:     os.Getenv("DB_PORT"),
		ProdDBSSLMode:  os.Getenv("DB_SSLMODE"),
		ProdDBName:     os.Getenv("DB_NAME"),
		Port:           os.Getenv("PORT"),
	}

	if env == "dev" {
		cfg.ServerAddress = fmt.Sprintf("localhost:%s", cfg.Port)
	} else {
		cfg.ServerAddress = ":" + cfg.Port
	}

	return cfg, nil
}

func GetEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
