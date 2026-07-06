package handler

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/internal/middleware"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/internal/storage"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/pkg/repository"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/pkg/response"
)

type StorageHandler struct {
	service *storage.Service
}

func NewStorageHandler(service *storage.Service) *StorageHandler {
	return &StorageHandler{service: service}
}

func (h *StorageHandler) Upload(c *fiber.Ctx) error {
	actor, err := middleware.ActorFromContext(c)
	if err != nil {
		return response.Unauthorized(c, "unauthorized")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "file is required")
	}
	src, err := file.Open()
	if err != nil {
		return response.BadRequest(c, "unable to read uploaded file")
	}
	defer src.Close()

	namespace := strings.TrimSpace(c.FormValue("namespace"))
	if namespace == "" {
		namespace = "attachments"
	}
	contentType := strings.TrimSpace(file.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	obj, err := h.service.Upload(c.Context(), storage.UploadRequest{
		TeamID:      actor.TeamID,
		Namespace:   namespace,
		Filename:    file.Filename,
		Body:        src,
		Size:        file.Size,
		ContentType: contentType,
	})
	if err != nil {
		return mapStorageError(c, err)
	}
	return response.Created(c, "file uploaded", fiber.Map{"object": obj})
}

func (h *StorageHandler) GetObject(c *fiber.Ctx) error {
	actor, err := middleware.ActorFromContext(c)
	if err != nil {
		return response.Unauthorized(c, "unauthorized")
	}
	objectKey := strings.TrimPrefix(c.Params("*"), "/")
	if objectKey == "" {
		return response.BadRequest(c, "object key is required")
	}

	reader, info, err := h.service.OpenForTeam(c.Context(), actor.TeamID, objectKey)
	if err != nil {
		return mapStorageError(c, err)
	}

	contentType := "application/octet-stream"
	if info != nil && info.ContentType != "" {
		contentType = info.ContentType
	} else if meta, metaErr := h.service.GetByObjectKey(c.Context(), objectKey); metaErr == nil {
		contentType = meta.MimeType
	}
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "private, max-age=300")
	if info != nil && info.Size > 0 {
		return c.SendStream(reader, int(info.Size))
	}
	return c.SendStream(reader)
}

func (h *StorageHandler) GetObjectContent(c *fiber.Ctx) error {
	actor, err := middleware.ActorFromContext(c)
	if err != nil {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuidParam(c, "id")
	if err != nil {
		return response.BadRequest(c, "invalid object id")
	}
	meta, reader, info, err := h.service.OpenForTeamByID(c.Context(), actor.TeamID, id)
	if err != nil {
		return mapStorageError(c, err)
	}

	contentType := meta.MimeType
	if info != nil && info.ContentType != "" {
		contentType = info.ContentType
	}
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "private, max-age=300")
	size := int(meta.SizeBytes)
	if info != nil && info.Size > 0 {
		size = int(info.Size)
	}
	if size > 0 {
		return c.SendStream(reader, size)
	}
	return c.SendStream(reader)
}

func (h *StorageHandler) GetMetadata(c *fiber.Ctx) error {
	actor, err := middleware.ActorFromContext(c)
	if err != nil {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuidParam(c, "id")
	if err != nil {
		return response.BadRequest(c, "invalid object id")
	}
	obj, err := h.service.GetByID(c.Context(), actor.TeamID, id)
	if err != nil {
		return mapStorageError(c, err)
	}
	return response.OK(c, "ok", fiber.Map{"object": obj})
}

func (h *StorageHandler) Delete(c *fiber.Ctx) error {
	actor, err := middleware.ActorFromContext(c)
	if err != nil {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuidParam(c, "id")
	if err != nil {
		return response.BadRequest(c, "invalid object id")
	}
	if err := h.service.Delete(c.Context(), actor.TeamID, id); err != nil {
		return mapStorageError(c, err)
	}
	return response.OK(c, "object deleted", nil)
}

func mapStorageError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, storage.ErrNotFound), errors.Is(err, repository.ErrNotFound):
		return response.NotFound(c, "object not found")
	case errors.Is(err, storage.ErrInvalidKey),
		errors.Is(err, storage.ErrUnsupportedMIME),
		errors.Is(err, storage.ErrFileTooLarge),
		errors.Is(err, storage.ErrEmptyUpload):
		return response.BadRequest(c, err.Error())
	case errors.Is(err, storage.ErrDriverNotImplemented):
		return response.InternalError(c, err.Error())
	default:
		return response.InternalError(c, "internal server error")
	}
}

func uuidParam(c *fiber.Ctx, name string) (uuid.UUID, error) {
	return uuid.Parse(c.Params(name))
}
