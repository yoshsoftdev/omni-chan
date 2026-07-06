package storage

import (
	"context"
	"fmt"

	"github.com/omnichannel/omni-chan-v2/services/storage-service/internal/config"
)

func NewProvider(ctx context.Context, cfg config.StorageConfig, publicBaseURL string) (StorageProvider, error) {
	_ = ctx
	switch cfg.Driver {
	case DriverLocal:
		return NewLocalStorageProvider(cfg, publicBaseURL)
	case DriverMinIO:
		return NewMinIOStorageProvider(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Driver)
	}
}

func New(ctx context.Context, cfg config.StorageConfig, publicBaseURL string, repo Repository) (*Service, error) {
	provider, err := NewProvider(ctx, cfg, publicBaseURL)
	if err != nil {
		return nil, err
	}
	return NewService(provider, repo), nil
}
