// Package config loads database settings from .env.
package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds the database settings shared by every shard.
type Config struct {
	Host     string
	DBName   string
	Username string
	Password string
}

// Load reads .env from the working directory and returns the settings.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil {
		return Config{}, err
	}
	return Config{
		Host:     os.Getenv("DB_HOST"),
		DBName:   os.Getenv("DB_NAME"),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
	}, nil
}
