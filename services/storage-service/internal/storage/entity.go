package storage

import (
	"io"
	"time"

	"github.com/google/uuid"
)

const (
	DriverLocal = "local"
	DriverMinIO = "minio"
)

// StoredObject is persisted metadata for a binary object. Never store absolute filesystem paths.
type StoredObject struct {
	ID            uuid.UUID `json:"id"`
	TeamID        uuid.UUID `json:"team_id"`
	ObjectKey     string    `json:"object_key"`
	StorageDriver string    `json:"storage_driver"`
	MimeType      string    `json:"mime_type"`
	SizeBytes     int64     `json:"size_bytes"`
	Hash          string    `json:"hash"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ObjectInfo describes an object in the active storage provider.
type ObjectInfo struct {
	Key         string
	Size        int64
	ContentType string
	Hash        string
	ModTime     time.Time
}

// UploadInput is provider-neutral upload payload.
type UploadInput struct {
	Key         string
	Body        io.Reader
	Size        int64
	ContentType string
}

// UploadRequest is used by the storage service when saving a new object.
type UploadRequest struct {
	TeamID      uuid.UUID
	Namespace   string
	Filename    string
	Body        io.Reader
	Size        int64
	ContentType string
}

// PresignResult contains a time-limited URL for reading an object.
type PresignResult struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}
