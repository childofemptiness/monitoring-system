package config

import "os"

type Config struct {
	DatabaseURL string
}

func LoadConfig() (Config, error) {
	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}, nil
}
