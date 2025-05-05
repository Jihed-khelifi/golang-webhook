package config

import (
	"fmt"
	"os"
	"strconv"
)

func GetStringEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}
	return value, nil
}

func GetIntEnv(key string) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, fmt.Errorf("required environment variable %s is not set", key)
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("environment variable %s is not a valid integer: %v", key, err)
	}
	return intValue, nil
}

func IsDevelopment() bool {
	env := os.Getenv("ENV")
	return env == "development" || env == "dev"
}
