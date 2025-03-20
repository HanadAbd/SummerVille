package util

import (
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
		ProdDBUser:     os.Getenv("PROD_DB_USER"),
		ProdDBPassword: os.Getenv("PROD_DB_PASSWORD"),
		ProdDBHost:     os.Getenv("PROD_DB_HOST"),
		ProdDBPort:     os.Getenv("PROD_DB_PORT"),
		ProdDBSSLMode:  os.Getenv("PROD_DB_SSLMODE"),
		ProdDBName:     os.Getenv("PROD_DB_NAME"),
	}

	if val := os.Getenv("SERVER_ADDR"); val != "" {
		cfg.ServerAddress = val
	}

	return cfg, nil
}
