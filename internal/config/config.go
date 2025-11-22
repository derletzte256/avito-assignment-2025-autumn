package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Host     string `mapstructure:"DATABASE_HOST"`
	Port     int    `mapstructure:"DATABASE_PORT"`
	Username string `mapstructure:"DATABASE_USER"`
	Password string `mapstructure:"DATABASE_PASSWORD"`
	Name     string `mapstructure:"DATABASE_NAME"`
}

type HTTPConfig struct {
	Address      string        `mapstructure:"HTTP_ADDRESS"`
	ReadTimeout  time.Duration `mapstructure:"HTTP_READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"HTTP_WRITE_TIMEOUT"`
	IdleTimeout  time.Duration `mapstructure:"HTTP_IDLE_TIMEOUT"`
}

type Config struct {
	Database DatabaseConfig
	HTTP     HTTPConfig
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()

	cfg := &Config{
		Database: DatabaseConfig{
			Host:     v.GetString("DATABASE_HOST"),
			Port:     v.GetInt("DATABASE_PORT"),
			Username: v.GetString("DATABASE_USER"),
			Password: v.GetString("DATABASE_PASSWORD"),
			Name:     v.GetString("DATABASE_NAME"),
		},
		HTTP: HTTPConfig{
			Address:      v.GetString("HTTP_ADDRESS"),
			ReadTimeout:  v.GetDuration("HTTP_READ_TIMEOUT"),
			WriteTimeout: v.GetDuration("HTTP_WRITE_TIMEOUT"),
			IdleTimeout:  v.GetDuration("HTTP_IDLE_TIMEOUT"),
		},
	}

	if cfg.Database.Host == "" {
		return nil, fmt.Errorf("DATABASE_HOST is required")
	}
	if cfg.Database.Port == 0 {
		return nil, fmt.Errorf("DATABASE_PORT is required")
	}
	if cfg.Database.Username == "" {
		return nil, fmt.Errorf("DATABASE_USER is required")
	}
	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("DATABASE_PASSWORD is required")
	}
	if cfg.Database.Name == "" {
		return nil, fmt.Errorf("DATABASE_NAME is required")
	}
	if cfg.HTTP.Address == "" {
		return nil, fmt.Errorf("HTTP_ADDRESS is required")
	}
	if cfg.HTTP.ReadTimeout == 0 {
		return nil, fmt.Errorf("HTTP_READ_TIMEOUT is required")
	}
	if cfg.HTTP.WriteTimeout == 0 {
		return nil, fmt.Errorf("HTTP_WRITE_TIMEOUT is required")
	}
	if cfg.HTTP.IdleTimeout == 0 {
		return nil, fmt.Errorf("HTTP_IDLE_TIMEOUT is required")
	}

	return cfg, nil
}
