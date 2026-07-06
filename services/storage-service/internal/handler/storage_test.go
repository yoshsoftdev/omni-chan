package handler

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/internal/config"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/internal/storage"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/pkg/jwt"
)

func TestGetObjectContentStreamsBody(t *testing.T) {
	dir := t.TempDir()
	provider, err := storage.NewLocalStorageProvider(config.StorageConfig{LocalPath: dir}, "http://localhost:8065")
	if err != nil {
		t.Fatalf("NewLocalStorageProvider() error = %v", err)
	}

	teamID := uuid.New()
	objectID := uuid.New()
	payload := []byte("jpeg-binary")
	key, err := storage.BuildObjectKey("attachments", teamID, "photo.jpeg")
	if err != nil {
		t.Fatalf("BuildObjectKey() error = %v", err)
	}
	if _, err := provider.Upload(context.Background(), storage.UploadInput{
		Key:         key,
		Body:        bytes.NewReader(payload),
		Size:        int64(len(payload)),
		ContentType: "image/jpeg",
	}); err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	repo := &memoryRepo{objects: map[uuid.UUID]*storage.StoredObject{
		objectID: {
			ID:            objectID,
			TeamID:        teamID,
			ObjectKey:     key,
			StorageDriver: storage.DriverLocal,
			MimeType:      "image/jpeg",
			SizeBytes:     int64(len(payload)),
		},
	}}
	svc := storage.NewService(provider, repo)
	handler := NewStorageHandler(svc)

	app := fiber.New()
	app.Get("/objects/:id/content", func(c *fiber.Ctx) error {
		c.Locals("auth_claims", &jwt.Claims{TeamID: teamID})
		return handler.GetObjectContent(c)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/objects/"+objectID.String()+"/content", nil)
	resp, err := app.Test(req, int(time.Second))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "image/jpeg" {
		t.Fatalf("Content-Type = %q", got)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if !bytes.Equal(body, payload) {
		t.Fatalf("body = %q", body)
	}
}

type memoryRepo struct {
	objects map[uuid.UUID]*storage.StoredObject
}

func (m *memoryRepo) Create(ctx context.Context, obj *storage.StoredObject) error { return nil }
func (m *memoryRepo) FindByID(ctx context.Context, teamID, id uuid.UUID) (*storage.StoredObject, error) {
	obj, ok := m.objects[id]
	if !ok || obj.TeamID != teamID {
		return nil, storage.ErrNotFound
	}
	return obj, nil
}
func (m *memoryRepo) FindByObjectKey(ctx context.Context, objectKey string) (*storage.StoredObject, error) {
	return nil, storage.ErrNotFound
}
func (m *memoryRepo) SoftDelete(ctx context.Context, teamID, id uuid.UUID) error { return nil }
