package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

type AssetService struct {
	assetRepo  port.AssetRepo
	assetStore port.AssetStore
}

func NewAssetService(assetRepo port.AssetRepo, assetStore port.AssetStore) *AssetService {
	return &AssetService{
		assetRepo:  assetRepo,
		assetStore: assetStore,
	}
}

func (s *AssetService) Upload(ctx context.Context, r io.ReadSeeker, contentType string, size int64) (*domain.Asset, error) {
	if err := domain.ValidateAssetContentType(contentType); err != nil {
		return nil, err
	}
	if size <= 0 {
		return nil, domain.NewValidationError("file", "empty or missing")
	}

	prefix := domain.AssetKeyPrefix(contentType)
	ext := filepath.Ext(contentTypeToExt(contentType))
	key := fmt.Sprintf("%s/%s%s", prefix, uuid.New().String(), ext)

	url, err := s.assetStore.Upload(ctx, key, r, contentType, size)
	if err != nil {
		return nil, err
	}

	asset := &domain.Asset{
		ID:          uuid.New(),
		Key:         key,
		URL:         url,
		ContentType: contentType,
		SizeBytes:   size,
	}
	if err := s.assetRepo.Create(ctx, asset); err != nil {
		// Best-effort cleanup of orphaned object.
		_ = s.assetStore.Delete(ctx, key)
		return nil, err
	}
	return asset, nil
}

func (s *AssetService) GetURL(ctx context.Context, key string) (string, error) {
	asset, err := s.assetRepo.GetByKey(ctx, key)
	if err != nil {
		return "", err
	}
	return s.assetStore.Presign(ctx, asset.Key)
}

func (s *AssetService) Delete(ctx context.Context, key string) error {
	if err := s.assetStore.Delete(ctx, key); err != nil {
		// Log but proceed — the DB record is the source of truth.
	}
	return s.assetRepo.DeleteByKey(ctx, key)
}

func contentTypeToExt(ct string) string {
	ct = strings.ToLower(ct)
	switch {
	case strings.Contains(ct, "jpeg") || strings.Contains(ct, "jpg"):
		return ".jpg"
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "gif"):
		return ".gif"
	case strings.Contains(ct, "webp"):
		return ".webp"
	case strings.Contains(ct, "avif"):
		return ".avif"
	case strings.Contains(ct, "svg"):
		return ".svg"
	default:
		return ""
	}
}
