package storage

import "testing"

func TestValidateUpload(t *testing.T) {
	if err := ValidateUpload("image/jpeg", 1024); err != nil {
		t.Fatalf("ValidateUpload() error = %v", err)
	}
	if err := ValidateUpload("image/gif", 1024); err != ErrUnsupportedMIME {
		t.Fatalf("expected unsupported mime, got %v", err)
	}
	if err := ValidateUpload("image/jpeg", 0); err != ErrEmptyUpload {
		t.Fatalf("expected empty upload, got %v", err)
	}
}
