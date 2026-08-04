package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/walfa-labs/backend/internal/domain"
)

// AssetRepo implements port.AssetRepo against PostgreSQL.
type AssetRepo struct {
	pool *pgxpool.Pool
}

// NewAssetRepo constructs an AssetRepo bound to the given pool.
func NewAssetRepo(pool *pgxpool.Pool) *AssetRepo {
	return &AssetRepo{pool: pool}
}

// Create inserts asset metadata and returns the generated ID/timestamp.
func (r *AssetRepo) Create(ctx context.Context, a *domain.Asset) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO assets (key, url, content_type, size_bytes)
		VALUES ($1, $2, $3, $4)
		RETURNING asset_id, uploaded_at`,
		a.Key, a.URL, a.ContentType, a.SizeBytes,
	).Scan(&a.ID, &a.UploadedAt)
	return err
}

// GetByKey returns the asset metadata for a given storage key.
func (r *AssetRepo) GetByKey(ctx context.Context, key string) (*domain.Asset, error) {
	var a domain.Asset
	err := r.pool.QueryRow(ctx, `
		SELECT asset_id, key, url, content_type, size_bytes, uploaded_at
		FROM assets WHERE key = $1`, key).Scan(
		&a.ID, &a.Key, &a.URL, &a.ContentType, &a.SizeBytes, &a.UploadedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

// DeleteByKey removes asset metadata by storage key.
func (r *AssetRepo) DeleteByKey(ctx context.Context, key string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM assets WHERE key = $1`, key)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
