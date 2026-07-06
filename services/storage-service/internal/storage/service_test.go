package storage

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/internal/config"
)

type memoryRepo struct {
	objects map[uuid.UUID]*StoredObject
	byKey   map[string]*StoredObject
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		objects: make(map[uuid.UUID]*StoredObject),
		byKey:   make(map[string]*StoredObject),
	}
}

func (m *memoryRepo) Create(ctx context.Context, obj *StoredObject) error {
	_ = ctx
	copyObj := *obj
	m.objects[obj.ID] = &copyObj
	m.byKey[obj.ObjectKey] = &copyObj
	return nil
}

func (m *memoryRepo) FindByID(ctx context.Context, teamID, id uuid.UUID) (*StoredObject, error) {
	_ = ctx
	obj, ok := m.objects[id]
	if !ok || obj.TeamID != teamID {
		return nil, ErrNotFound
	}
	copyObj := *obj
	return &copyObj, nil
}

func (m *memoryRepo) FindByObjectKey(ctx context.Context, objectKey string) (*StoredObject, error) {
	_ = ctx
	obj, ok := m.byKey[objectKey]
	if !ok {
		return nil, ErrNotFound
	}
	copyObj := *obj
	return &copyObj, nil
}

func (m *memoryRepo) SoftDelete(ctx context.Context, teamID, id uuid.UUID) error {
	_ = ctx
	obj, ok := m.objects[id]
	if !ok || obj.TeamID != teamID {
		return ErrNotFound
	}
	delete(m.byKey, obj.ObjectKey)
	delete(m.objects, id)
	return nil
}

func TestService_UploadPersistsMetadata(t *testing.T) {
	dir := t.TempDir()
	provider, err := NewLocalStorageProvider(configWithPath(dir), "http://localhost:8080")
	if err != nil {
		t.Fatalf("NewLocalStorageProvider() error = %v", err)
	}
	repo := newMemoryRepo()
	svc := NewService(provider, repo)

	teamID := uuid.New()
	body := []byte("payload")
	obj, err := svc.Upload(context.Background(), UploadRequest{
		TeamID:      teamID,
		Namespace:   "attachments",
		Filename:    "note.pdf",
		Body:        bytes.NewReader(body),
		Size:        int64(len(body)),
		ContentType: "application/pdf",
	})
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if obj.StorageDriver != DriverLocal {
		t.Fatalf("driver = %s", obj.StorageDriver)
	}
	if obj.ObjectKey == "" || obj.Hash == "" {
		t.Fatal("expected metadata fields")
	}

	stored, err := repo.FindByID(context.Background(), teamID, obj.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if stored.ObjectKey != obj.ObjectKey {
		t.Fatalf("object key mismatch")
	}
}

func configWithPath(path string) config.StorageConfig {
	return config.StorageConfig{LocalPath: path}
}
