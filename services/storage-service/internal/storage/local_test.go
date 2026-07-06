package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/omnichannel/omni-chan-v2/services/storage-service/internal/config"
)

func TestLocalStorageProvider_UploadOpenDelete(t *testing.T) {
	dir := t.TempDir()
	provider, err := NewLocalStorageProvider(config.StorageConfig{LocalPath: dir}, "http://localhost:8080")
	if err != nil {
		t.Fatalf("NewLocalStorageProvider() error = %v", err)
	}

	ctx := context.Background()
	body := []byte("hello storage")
	info, err := provider.Upload(ctx, UploadInput{
		Key:         "attachments/team/2026/06/test.txt",
		Body:        bytes.NewReader(body),
		Size:        int64(len(body)),
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if info.Size != int64(len(body)) {
		t.Fatalf("size = %d", info.Size)
	}
	if info.Hash == "" {
		t.Fatal("expected hash")
	}

	exists, err := provider.Exists(ctx, info.Key)
	if err != nil || !exists {
		t.Fatalf("Exists() = %v, %v", exists, err)
	}

	reader, stat, err := provider.Open(ctx, info.Key)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer reader.Close()
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("content mismatch")
	}
	if stat.Size != int64(len(body)) {
		t.Fatalf("stat size = %d", stat.Size)
	}

	presign, err := provider.PresignGet(ctx, info.Key, 0)
	if err != nil {
		t.Fatalf("PresignGet() error = %v", err)
	}
	if presign.URL == "" {
		t.Fatal("expected presign url")
	}

	if err := provider.Delete(ctx, info.Key); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	target := filepath.Join(dir, filepath.FromSlash(info.Key))
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("file should be deleted: %v", err)
	}
}

func TestLocalStorageProvider_RejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	provider, err := NewLocalStorageProvider(config.StorageConfig{LocalPath: dir}, "http://localhost:8080")
	if err != nil {
		t.Fatalf("NewLocalStorageProvider() error = %v", err)
	}
	_, err = provider.Upload(context.Background(), UploadInput{
		Key:  "../outside.txt",
		Body: bytes.NewReader([]byte("x")),
	})
	if err == nil {
		t.Fatal("expected path traversal to be rejected")
	}
}

func TestMinIOStorageProvider_IsStub(t *testing.T) {
	provider := NewMinIOStorageProvider(config.StorageConfig{MinIOBucket: "omni-chan"})
	if provider.Driver() != DriverMinIO {
		t.Fatalf("driver = %s", provider.Driver())
	}
	_, err := provider.Upload(context.Background(), UploadInput{Key: "a/b"})
	if err != ErrDriverNotImplemented {
		t.Fatalf("Upload() error = %v", err)
	}
}
