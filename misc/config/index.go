package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type AppConfig struct {
	Port          int
	Endpoint      string
	BatchSize     int
	BatchInterval time.Duration
}

type Env struct {
	once sync.Once
	App  *AppConfig
}

var (
	envInstance = &Env{}
)

func GetEnv() *Env {
	return envInstance
}

func loadDotEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		os.Setenv(key, val)
	}
	return scanner.Err()
}

func (e *Env) Setup() error {
	var setupErr error
	e.once.Do(func() {
		if err := loadDotEnv(".env"); err != nil {
			setupErr = fmt.Errorf("failed to load .env file: %w", err)
			return
		}
		port, err := GetIntEnv("PORT")
		if err != nil {
			setupErr = err
			return
		}
		endpoint, err := GetStringEnv("ENDPOINT")
		if err != nil {
			setupErr = err
			return
		}
		batchSize, err := GetIntEnv("BATCH_SIZE")
		if err != nil {
			setupErr = err
			return
		}
		batchInterval, err := GetIntEnv("BATCH_INTERVAL")
		if err != nil {
			setupErr = err
			return
		}

		e.App = &AppConfig{
			Port:          port,
			Endpoint:      endpoint,
			BatchSize:     batchSize,
			BatchInterval: time.Duration(batchInterval) * time.Second,
		}
	})

	return setupErr
}
