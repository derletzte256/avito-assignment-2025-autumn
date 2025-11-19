package config

import "time"

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Name     string
}

type HTTPConfig struct {
	Address      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type Config struct {
	Database DatabaseConfig
	HTTP     HTTPConfig
}

func LoadConfig() (*Config, error) {
	return &Config{
		Database: DatabaseConfig{
			Host:     "db",
			Port:     5432,
			Username: "avito",
			Password: "password",
			Name:     "reviews",
		},
		HTTP: HTTPConfig{
			Address:      ":8080",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}, nil
}
