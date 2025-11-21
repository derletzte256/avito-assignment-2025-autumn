package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

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

const envFile = ".env"

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(envFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load %s: %w", envFile, err)
	}

	dbHost, err := getStringEnv("DATABASE_HOST")
	if err != nil {
		return nil, err
	}

	dbPort, err := getIntEnv("DATABASE_PORT")
	if err != nil {
		return nil, err
	}

	dbUser, err := getStringEnv("DATABASE_USER")
	if err != nil {
		return nil, err
	}

	dbPassword, err := getStringEnv("DATABASE_PASSWORD")
	if err != nil {
		return nil, err
	}

	dbName, err := getStringEnv("DATABASE_NAME")
	if err != nil {
		return nil, err
	}

	httpAddress, err := getStringEnv("HTTP_ADDRESS")
	if err != nil {
		return nil, err
	}

	readTimeout, err := getDurationEnv("HTTP_READ_TIMEOUT")
	if err != nil {
		return nil, err
	}

	writeTimeout, err := getDurationEnv("HTTP_WRITE_TIMEOUT")
	if err != nil {
		return nil, err
	}

	idleTimeout, err := getDurationEnv("HTTP_IDLE_TIMEOUT")
	if err != nil {
		return nil, err
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     dbHost,
			Port:     dbPort,
			Username: dbUser,
			Password: dbPassword,
			Name:     dbName,
		},
		HTTP: HTTPConfig{
			Address:      httpAddress,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
	}, nil
}

func getStringEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return "", fmt.Errorf("environment variable %s is not set", key)
	}
	return value, nil
}

func getIntEnv(key string) (int, error) {
	value, err := getStringEnv(key)
	if err != nil {
		return 0, err
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s=%q as int: %w", key, value, err)
	}

	return parsed, nil
}

func getDurationEnv(key string) (time.Duration, error) {
	value, err := getStringEnv(key)
	if err != nil {
		return 0, err
	}

	dur, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s=%q as duration: %w", key, value, err)
	}

	return dur, nil
}
