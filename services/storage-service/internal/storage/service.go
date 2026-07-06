package storage

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service is the only entry point application modules should use for object storage.
type Service struct {
	provider StorageProvider
	repo     Repository
}

func NewService(provider StorageProvider, repo Repository) *Service {
	return &Service{provider: provider, repo: repo}
}

func (s *Service) Driver() string {
	return s.provider.Driver()
}

func (s *Service) Upload(ctx context.Context, req UploadRequest) (*StoredObject, error) {
	mimeType := strings.TrimSpace(req.ContentType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if err := ValidateUpload(mimeType, req.Size); err != nil {
		return nil, err
	}

	namespace := strings.Trim(strings.TrimSpace(req.Namespace), "/")
	if namespace == "" {
		namespace = "attachments"
	}
	key, err := BuildObjectKey(namespace, req.TeamID, req.Filename)
	if err != nil {
		return nil, err
	}

	info, err := s.provider.Upload(ctx, UploadInput{
		Key:         key,
		Body:        req.Body,
		Size:        req.Size,
		ContentType: strings.TrimSpace(req.ContentType),
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	obj := &StoredObject{
		ID:            uuid.New(),
		TeamID:        req.TeamID,
		ObjectKey:     info.Key,
		StorageDriver: s.provider.Driver(),
		MimeType:      info.ContentType,
		SizeBytes:     info.Size,
		Hash:          info.Hash,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if obj.MimeType == "" {
		obj.MimeType = "application/octet-stream"
	}
	if s.repo != nil {
		if err := s.repo.Create(ctx, obj); err != nil {
			_ = s.provider.Delete(ctx, key)
			return nil, err
		}
	}
	return obj, nil
}

func (s *Service) Open(ctx context.Context, objectKey string) (io.ReadCloser, *ObjectInfo, error) {
	return s.provider.Open(ctx, objectKey)
}

func (s *Service) OpenForTeam(ctx context.Context, teamID uuid.UUID, objectKey string) (io.ReadCloser, *ObjectInfo, error) {
	if s.repo != nil {
		meta, err := s.repo.FindByObjectKey(ctx, objectKey)
		if err != nil {
			return nil, nil, err
		}
		if meta.TeamID != teamID {
			return nil, nil, ErrNotFound
		}
	}
	return s.provider.Open(ctx, objectKey)
}

func (s *Service) OpenForTeamByID(ctx context.Context, teamID, objectID uuid.UUID) (*StoredObject, io.ReadCloser, *ObjectInfo, error) {
	if s.repo == nil {
		return nil, nil, nil, ErrNotFound
	}
	meta, err := s.repo.FindByID(ctx, teamID, objectID)
	if err != nil {
		return nil, nil, nil, err
	}
	reader, info, err := s.provider.Open(ctx, meta.ObjectKey)
	if err != nil {
		return nil, nil, nil, err
	}
	if info != nil && info.ContentType == "" {
		info.ContentType = meta.MimeType
	}
	return meta, reader, info, nil
}

func (s *Service) Delete(ctx context.Context, teamID uuid.UUID, objectID uuid.UUID) error {
	if s.repo == nil {
		return ErrNotFound
	}
	meta, err := s.repo.FindByID(ctx, teamID, objectID)
	if err != nil {
		return err
	}
	if err := s.provider.Delete(ctx, meta.ObjectKey); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, teamID, objectID)
}

func (s *Service) Exists(ctx context.Context, objectKey string) (bool, error) {
	return s.provider.Exists(ctx, objectKey)
}

func (s *Service) Stat(ctx context.Context, objectKey string) (*ObjectInfo, error) {
	return s.provider.Stat(ctx, objectKey)
}

func (s *Service) PresignGet(ctx context.Context, objectKey string, expiry time.Duration) (*PresignResult, error) {
	return s.provider.PresignGet(ctx, objectKey, expiry)
}

func (s *Service) GetByID(ctx context.Context, teamID, objectID uuid.UUID) (*StoredObject, error) {
	if s.repo == nil {
		return nil, ErrNotFound
	}
	return s.repo.FindByID(ctx, teamID, objectID)
}

func (s *Service) GetByObjectKey(ctx context.Context, objectKey string) (*StoredObject, error) {
	if s.repo == nil {
		return nil, ErrNotFound
	}
	return s.repo.FindByObjectKey(ctx, objectKey)
}
