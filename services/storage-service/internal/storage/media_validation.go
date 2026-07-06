package storage

import (
	"fmt"
	"strings"
)

const (
	MaxImageBytes    int64 = 5 * 1024 * 1024
	MaxVideoBytes    int64 = 16 * 1024 * 1024
	MaxAudioBytes    int64 = 16 * 1024 * 1024
	MaxDocumentBytes int64 = 100 * 1024 * 1024
)

var (
	ErrUnsupportedMIME = fmt.Errorf("unsupported mime type")
	ErrFileTooLarge    = fmt.Errorf("file exceeds allowed size")
	ErrEmptyUpload     = fmt.Errorf("empty upload")
)

func ValidateUpload(mimeType string, sizeBytes int64) error {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	if sizeBytes <= 0 {
		return ErrEmptyUpload
	}
	limit, ok := maxSizeForMIME(mimeType)
	if !ok {
		return ErrUnsupportedMIME
	}
	if sizeBytes > limit {
		return ErrFileTooLarge
	}
	return nil
}

func maxSizeForMIME(mimeType string) (int64, bool) {
	switch mimeType {
	case "image/jpeg", "image/png", "image/webp":
		return MaxImageBytes, true
	case "video/mp4", "video/3gpp":
		return MaxVideoBytes, true
	case "audio/aac", "audio/mp4", "audio/mpeg", "audio/amr", "audio/ogg", "audio/ogg; codecs=opus":
		return MaxAudioBytes, true
	case "application/pdf",
		"application/vnd.ms-excel",
		"application/vnd.ms-powerpoint",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return MaxDocumentBytes, true
	default:
		return 0, false
	}
}
