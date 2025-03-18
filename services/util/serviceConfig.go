package util

import (
	"encoding/json"
	"os"
)

type Connections struct {
	Workspaces map[string]interface{} `json:"workspaces"`
}

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

	Connections Connections
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

	data, err := os.ReadFile("connections.json")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &cfg.Connections); err != nil {
		return nil, err
	}

	if val := os.Getenv("SERVER_ADDR"); val != "" {
		cfg.ServerAddress = val
	}

	return cfg, nil
}
