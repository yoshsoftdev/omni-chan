package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Base struct {
	AppName string
	AppEnv  string
}

type HTTP struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type NATS struct {
	URL              string
	StreamName       string
	InternalToken    string
	MaxPublishRetry  int
	PublishRetryWait time.Duration
}

func (h HTTP) Addr() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

func LoadBase() Base {
	_ = godotenv.Load()
	return Base{
		AppName: getEnv("APP_NAME", "omni-chan"),
		AppEnv:  getEnv("APP_ENV", "development"),
	}
}

func LoadHTTP(defaultPort int) HTTP {
	_ = godotenv.Load()
	return HTTP{
		Host:         getEnv("HTTP_HOST", "0.0.0.0"),
		Port:         getEnvInt("HTTP_PORT", defaultPort),
		ReadTimeout:  getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout: getEnvDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
	}
}

func LoadNATS() NATS {
	_ = godotenv.Load()
	return NATS{
		URL:              getEnv("NATS_URL", "nats://localhost:4222"),
		StreamName:       getEnv("NATS_STREAM_NAME", "WEBHOOK_EVENTS"),
		InternalToken:    getEnv("INTERNAL_SERVICE_TOKEN", ""),
		MaxPublishRetry:  getEnvInt("NATS_PUBLISH_MAX_RETRIES", 3),
		PublishRetryWait: getEnvDuration("NATS_PUBLISH_RETRY_WAIT", 200*time.Millisecond),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}
