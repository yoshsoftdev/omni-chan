package storage

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ValidateObjectKey ensures keys are provider-independent and path-safe.
func ValidateObjectKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}
	if strings.HasPrefix(key, "/") || strings.Contains(key, "\\") || strings.Contains(key, "..") {
		return ErrInvalidKey
	}
	clean := path.Clean(key)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return ErrInvalidKey
	}
	return nil
}

// BuildObjectKey generates a provider-independent object key.
// Example: attachments/{team_id}/{yyyy}/{mm}/{uuid}.jpg
func BuildObjectKey(namespace string, teamID uuid.UUID, originalFilename string) (string, error) {
	namespace = strings.Trim(strings.TrimSpace(namespace), "/")
	if namespace == "" {
		return "", ErrInvalidKey
	}
	if teamID == uuid.Nil {
		return "", ErrInvalidKey
	}

	now := time.Now().UTC()
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext != "" && !isSafeExtension(ext) {
		ext = ""
	}

	key := fmt.Sprintf("%s/%s/%04d/%02d/%s%s", namespace, teamID.String(), now.Year(), int(now.Month()), uuid.NewString(), ext)
	if err := ValidateObjectKey(key); err != nil {
		return "", err
	}
	return key, nil
}

func isSafeExtension(ext string) bool {
	ext = strings.TrimPrefix(ext, ".")
	if ext == "" || len(ext) > 10 {
		return false
	}
	for _, ch := range ext {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			continue
		}
		return false
	}
	return true
}
