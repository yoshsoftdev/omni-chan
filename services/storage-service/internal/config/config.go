package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	sharedconfig "github.com/omnichannel/omni-chan-v2/shared/config"
)

type Config struct {
	App           sharedconfig.Base
	HTTP          sharedconfig.HTTP
	Database      DatabaseConfig
	Storage       StorageConfig
	JWT           JWTConfig
	InternalToken string
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type StorageConfig struct {
	Driver        string
	LocalPath     string
	PublicBaseURL string
	MinIOEndpoint string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket   string
	MinIOUseSSL   bool
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	cfg := &Config{
		App:  sharedconfig.LoadBase(),
		HTTP: sharedconfig.LoadHTTP(8082),
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", ""),
			MaxOpenConns:    getEnvInt("DATABASE_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getEnvInt("DATABASE_MAX_IDLE_CONNS", 2),
			ConnMaxLifetime: getEnvDuration("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Storage: StorageConfig{
			Driver:         getEnv("STORAGE_DRIVER", "local"),
			LocalPath:      getEnv("STORAGE_LOCAL_PATH", "./storage"),
			PublicBaseURL:  getEnv("STORAGE_PUBLIC_BASE_URL", ""),
			MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", ""),
			MinIOSecretKey: getEnv("MINIO_SECRET_KEY", ""),
			MinIOBucket:    getEnv("MINIO_BUCKET", "omni-chan"),
			MinIOUseSSL:    getEnvBool("MINIO_USE_SSL", false),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
			AccessTTL:     getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
		},
		InternalToken: getEnv("INTERNAL_SERVICE_TOKEN", ""),
	}
	return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.Storage.Driver != "local" && c.Storage.Driver != "minio" {
		return fmt.Errorf("STORAGE_DRIVER must be local or minio")
	}
	if c.Storage.Driver == "minio" {
		if c.Storage.MinIOAccessKey == "" || c.Storage.MinIOSecretKey == "" {
			return fmt.Errorf("MINIO_ACCESS_KEY and MINIO_SECRET_KEY are required when STORAGE_DRIVER=minio")
		}
	}
	if c.JWT.AccessSecret == "" || c.JWT.RefreshSecret == "" {
		if c.App.AppEnv == "development" {
			c.JWT.AccessSecret = "dev-access-secret-change-in-production"
			c.JWT.RefreshSecret = "dev-refresh-secret-change-in-production"
		} else {
			return fmt.Errorf("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET are required")
		}
	}
	if c.InternalToken == "" && c.App.AppEnv == "development" {
		c.InternalToken = "dev-internal-token"
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "1" || v == "true" || v == "TRUE"
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
