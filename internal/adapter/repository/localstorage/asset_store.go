package localstorage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/walfa-labs/backend/internal/port"
)

var _ port.AssetStore = (*AssetStore)(nil)

// AssetStore implements port.AssetStore using local filesystem storage.
type AssetStore struct {
	baseDir string
	baseURL string
}

// NewAssetStore initializes a local disk asset store.
func NewAssetStore(baseDir, baseURL string) (*AssetStore, error) {
	if baseDir == "" {
		baseDir = "./uploads"
	}
	if baseURL == "" {
		baseURL = "/uploads"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("local asset store: create base dir: %w", err)
	}

	return &AssetStore{
		baseDir: baseDir,
		baseURL: baseURL,
	}, nil
}

// Upload writes the file stream to the local baseDir and returns its accessible URL.
func (s *AssetStore) Upload(ctx context.Context, key string, r io.Reader, contentType string, size int64) (string, error) {
	cleanKey := filepath.Clean(filepath.FromSlash(key))
	targetPath := filepath.Join(s.baseDir, cleanKey)

	// Ensure subdirectories exist
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return "", fmt.Errorf("local asset store: create dirs for %q: %w", key, err)
	}

	out, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("local asset store: create file %q: %w", key, err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, r); err != nil {
		return "", fmt.Errorf("local asset store: write file %q: %w", key, err)
	}

	return s.buildURL(key), nil
}

// Presign returns the direct URL to the static asset in local mode.
func (s *AssetStore) Presign(ctx context.Context, key string) (string, error) {
	return s.buildURL(key), nil
}

// Delete removes the asset from local disk.
func (s *AssetStore) Delete(ctx context.Context, key string) error {
	cleanKey := filepath.Clean(filepath.FromSlash(key))
	targetPath := filepath.Join(s.baseDir, cleanKey)

	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local asset store: delete %q: %w", key, err)
	}
	return nil
}

func (s *AssetStore) buildURL(key string) string {
	cleanSlashKey := strings.TrimPrefix(filepath.ToSlash(key), "/")
	return fmt.Sprintf("%s/%s", s.baseURL, cleanSlashKey)
}
