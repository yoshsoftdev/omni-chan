package storage

import "errors"

var (
	ErrInvalidKey          = errors.New("invalid object key")
	ErrNotFound            = errors.New("object not found")
	ErrDriverNotImplemented = errors.New("storage driver not implemented")
)
