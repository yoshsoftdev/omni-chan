package storage

import (
	"context"
	"io"
	"time"

	"github.com/omnichannel/omni-chan-v2/services/storage-service/internal/config"
)

// MinIOStorageProvider is a placeholder for future MinIO integration.
// Selecting STORAGE_DRIVER=minio wires this provider without changing application code.
type MinIOStorageProvider struct {
	bucket string
}

func NewMinIOStorageProvider(cfg config.StorageConfig) *MinIOStorageProvider {
	return &MinIOStorageProvider{bucket: cfg.MinIOBucket}
}

func (p *MinIOStorageProvider) Driver() string { return DriverMinIO }

func (p *MinIOStorageProvider) Upload(ctx context.Context, input UploadInput) (*ObjectInfo, error) {
	_ = ctx
	_ = input
	return nil, ErrDriverNotImplemented
}

func (p *MinIOStorageProvider) Open(ctx context.Context, key string) (io.ReadCloser, *ObjectInfo, error) {
	_ = ctx
	_ = key
	return nil, nil, ErrDriverNotImplemented
}

func (p *MinIOStorageProvider) Delete(ctx context.Context, key string) error {
	_ = ctx
	_ = key
	return ErrDriverNotImplemented
}

func (p *MinIOStorageProvider) Exists(ctx context.Context, key string) (bool, error) {
	_ = ctx
	_ = key
	return false, ErrDriverNotImplemented
}

func (p *MinIOStorageProvider) Stat(ctx context.Context, key string) (*ObjectInfo, error) {
	_ = ctx
	_ = key
	return nil, ErrDriverNotImplemented
}

func (p *MinIOStorageProvider) PresignGet(ctx context.Context, key string, expiry time.Duration) (*PresignResult, error) {
	_ = ctx
	_ = key
	_ = expiry
	return nil, ErrDriverNotImplemented
}
