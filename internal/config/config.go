package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort       string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	RedisPoolSize int
	RedisTimeout  time.Duration
	RateLimitRPM  int
}

func Load() (Config, error) {
	_ = godotenv.Load()

	redisDB, err := readInt("REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}

	poolSize, err := readInt("REDIS_POOL_SIZE", 10)
	if err != nil {
		return Config{}, err
	}

	redisTimeout, err := readDuration("REDIS_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	rateLimitRPM, err := readInt("RATE_LIMIT_RPM", 120)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppPort:       readString("APP_PORT", "8080"),
		RedisHost:     readString("REDIS_HOST", "localhost"),
		RedisPort:     readString("REDIS_PORT", "6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       redisDB,
		RedisPoolSize: poolSize,
		RedisTimeout:  redisTimeout,
		RateLimitRPM:  rateLimitRPM,
	}

	if cfg.AppPort == "" {
		return Config{}, fmt.Errorf("APP_PORT cannot be empty")
	}

	if cfg.RedisHost == "" {
		return Config{}, fmt.Errorf("REDIS_HOST cannot be empty")
	}

	if cfg.RedisPort == "" {
		return Config{}, fmt.Errorf("REDIS_PORT cannot be empty")
	}

	if cfg.RedisPoolSize <= 0 {
		return Config{}, fmt.Errorf("REDIS_POOL_SIZE must be greater than 0")
	}

	if cfg.RedisTimeout <= 0 {
		return Config{}, fmt.Errorf("REDIS_TIMEOUT must be greater than 0")
	}

	if cfg.RateLimitRPM <= 0 {
		return Config{}, fmt.Errorf("RATE_LIMIT_RPM must be greater than 0")
	}

	return cfg, nil
}

func (c Config) RedisAddress() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func readString(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func readInt(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid integer for %s: %w", key, err)
	}

	return value, nil
}

func readDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}

	return value, nil
}
