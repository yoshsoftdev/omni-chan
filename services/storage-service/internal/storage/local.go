package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/omnichannel/omni-chan-v2/services/storage-service/internal/config"
)

type LocalStorageProvider struct {
	basePath   string
	publicBase string
}

func NewLocalStorageProvider(cfg config.StorageConfig, publicBaseURL string) (*LocalStorageProvider, error) {
	basePath := strings.TrimSpace(cfg.LocalPath)
	if basePath == "" {
		basePath = "./storage"
	}
	abs, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("resolve storage path: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}
	publicBase := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if publicBase == "" {
		publicBase = "http://localhost:8080"
	}
	return &LocalStorageProvider{basePath: abs, publicBase: publicBase}, nil
}

func (p *LocalStorageProvider) Driver() string { return DriverLocal }

func (p *LocalStorageProvider) Upload(ctx context.Context, input UploadInput) (*ObjectInfo, error) {
	if err := ValidateObjectKey(input.Key); err != nil {
		return nil, err
	}
	target, err := p.resolvePath(input.Key)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return nil, fmt.Errorf("create object directory: %w", err)
	}

	file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open object file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(file, hasher)
	written, err := io.Copy(writer, input.Body)
	if err != nil {
		_ = os.Remove(target)
		return nil, fmt.Errorf("write object: %w", err)
	}
	if input.Size > 0 && written != input.Size {
		_ = os.Remove(target)
		return nil, fmt.Errorf("upload size mismatch")
	}

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat object: %w", err)
	}
	_ = ctx

	return &ObjectInfo{
		Key:         input.Key,
		Size:        info.Size(),
		ContentType: input.ContentType,
		Hash:        hex.EncodeToString(hasher.Sum(nil)),
		ModTime:     info.ModTime(),
	}, nil
}

func (p *LocalStorageProvider) Open(ctx context.Context, key string) (io.ReadCloser, *ObjectInfo, error) {
	if err := ValidateObjectKey(key); err != nil {
		return nil, nil, err
	}
	target, err := p.resolvePath(key)
	if err != nil {
		return nil, nil, err
	}
	file, err := os.Open(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, fmt.Errorf("open object: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, nil, err
	}
	_ = ctx
	return file, &ObjectInfo{
		Key:     key,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}, nil
}

func (p *LocalStorageProvider) Delete(ctx context.Context, key string) error {
	if err := ValidateObjectKey(key); err != nil {
		return err
	}
	target, err := p.resolvePath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(target); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("delete object: %w", err)
	}
	_ = ctx
	return nil
}

func (p *LocalStorageProvider) Exists(ctx context.Context, key string) (bool, error) {
	if err := ValidateObjectKey(key); err != nil {
		return false, err
	}
	target, err := p.resolvePath(key)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	_ = ctx
	return true, nil
}

func (p *LocalStorageProvider) Stat(ctx context.Context, key string) (*ObjectInfo, error) {
	if err := ValidateObjectKey(key); err != nil {
		return nil, err
	}
	target, err := p.resolvePath(key)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	_ = ctx
	return &ObjectInfo{
		Key:     key,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}, nil
}

func (p *LocalStorageProvider) PresignGet(ctx context.Context, key string, expiry time.Duration) (*PresignResult, error) {
	if err := ValidateObjectKey(key); err != nil {
		return nil, err
	}
	exists, err := p.Exists(ctx, key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}
	return &PresignResult{
		URL:       fmt.Sprintf("%s/api/v1/storage/blob/%s", p.publicBase, escapeObjectKeyForURL(key)),
		ExpiresAt: time.Now().UTC().Add(expiry),
	}, nil
}

func (p *LocalStorageProvider) resolvePath(key string) (string, error) {
	if err := ValidateObjectKey(key); err != nil {
		return "", err
	}
	target := filepath.Join(p.basePath, filepath.FromSlash(key))
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	absBase, err := filepath.Abs(p.basePath)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(absTarget, absBase+string(os.PathSeparator)) && absTarget != absBase {
		return "", ErrInvalidKey
	}
	return absTarget, nil
}

func escapeObjectKeyForURL(key string) string {
	parts := strings.Split(key, "/")
	for i, part := range parts {
		parts[i] = strings.ReplaceAll(part, " ", "%20")
	}
	return strings.Join(parts, "/")
}
