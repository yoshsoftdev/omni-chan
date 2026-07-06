package storage

import (
	"context"

	"github.com/google/uuid"
	storagedb "github.com/omnichannel/omni-chan-v2/services/storage-service/pkg/db"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/pkg/repository"
)

type Repository interface {
	Create(ctx context.Context, obj *StoredObject) error
	FindByID(ctx context.Context, teamID, id uuid.UUID) (*StoredObject, error)
	FindByObjectKey(ctx context.Context, objectKey string) (*StoredObject, error)
	SoftDelete(ctx context.Context, teamID, id uuid.UUID) error
}

type PostgresRepository struct {
	db *storagedb.DB
}

func NewRepository(db *storagedb.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, obj *StoredObject) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO stored_objects (id, team_id, object_key, storage_driver, mime_type, size_bytes, hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, obj.ID, obj.TeamID, obj.ObjectKey, obj.StorageDriver, obj.MimeType, obj.SizeBytes, obj.Hash, obj.CreatedAt, obj.UpdatedAt)
	return repository.MapError(err)
}

func (r *PostgresRepository) FindByID(ctx context.Context, teamID, id uuid.UUID) (*StoredObject, error) {
	query := `
		SELECT id, team_id, object_key, storage_driver, mime_type, size_bytes, hash, created_at, updated_at
		FROM stored_objects
		WHERE id = $1 AND team_id = $2 AND deleted_at IS NULL
	`
	var obj StoredObject
	err := r.db.Pool.QueryRow(ctx, query, id, teamID).Scan(
		&obj.ID, &obj.TeamID, &obj.ObjectKey, &obj.StorageDriver, &obj.MimeType, &obj.SizeBytes, &obj.Hash, &obj.CreatedAt, &obj.UpdatedAt,
	)
	if err != nil {
		return nil, repository.MapError(err)
	}
	return &obj, nil
}

func (r *PostgresRepository) FindByObjectKey(ctx context.Context, objectKey string) (*StoredObject, error) {
	query := `
		SELECT id, team_id, object_key, storage_driver, mime_type, size_bytes, hash, created_at, updated_at
		FROM stored_objects
		WHERE object_key = $1 AND deleted_at IS NULL
	`
	var obj StoredObject
	err := r.db.Pool.QueryRow(ctx, query, objectKey).Scan(
		&obj.ID, &obj.TeamID, &obj.ObjectKey, &obj.StorageDriver, &obj.MimeType, &obj.SizeBytes, &obj.Hash, &obj.CreatedAt, &obj.UpdatedAt,
	)
	if err != nil {
		return nil, repository.MapError(err)
	}
	return &obj, nil
}

func (r *PostgresRepository) SoftDelete(ctx context.Context, teamID, id uuid.UUID) error {
	tag, err := r.db.Pool.Exec(ctx, `
		UPDATE stored_objects
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND team_id = $2 AND deleted_at IS NULL
	`, id, teamID)
	if err != nil {
		return repository.MapError(err)
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}
