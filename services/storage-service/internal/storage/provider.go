package storage

import (
	"context"
	"io"
	"time"
)

// StorageProvider abstracts object storage backends (local filesystem, MinIO, etc.).
type StorageProvider interface {
	Driver() string
	Upload(ctx context.Context, input UploadInput) (*ObjectInfo, error)
	Open(ctx context.Context, key string) (io.ReadCloser, *ObjectInfo, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Stat(ctx context.Context, key string) (*ObjectInfo, error)
	PresignGet(ctx context.Context, key string, expiry time.Duration) (*PresignResult, error)
}
