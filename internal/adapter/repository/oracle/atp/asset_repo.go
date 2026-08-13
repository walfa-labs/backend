package atp

import (
	"context"
	"database/sql"
	"errors"

	"github.com/walfa-labs/backend/internal/domain"
)

// AssetRepo implements port.AssetRepo against Oracle ATP.
type AssetRepo struct {
	db *sql.DB
}

// NewAssetRepo constructs an AssetRepo bound to the given pool.
func NewAssetRepo(db *sql.DB) *AssetRepo {
	return &AssetRepo{db: db}
}

// Create inserts asset metadata; uploaded_at is set by the DB default.
func (r *AssetRepo) Create(ctx context.Context, a *domain.Asset) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO assets (asset_id, key, url, content_type, size_bytes)
		VALUES (:1, :2, :3, :4, :5)`,
		a.ID.String(), a.Key, a.URL, a.ContentType, a.SizeBytes,
	)
	if err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx,
		`SELECT uploaded_at FROM assets WHERE asset_id = :1`, a.ID.String(),
	).Scan(&a.UploadedAt)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetByKey returns the asset metadata for a given storage key.
func (r *AssetRepo) GetByKey(ctx context.Context, key string) (*domain.Asset, error) {
	var a domain.Asset
	var url, contentType sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT asset_id, key, url, content_type, size_bytes, uploaded_at
		FROM assets WHERE key = :1`, key).Scan(
		&a.ID, &a.Key, &url, &contentType, &a.SizeBytes, &a.UploadedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	a.URL = nullStr(url)
	a.ContentType = nullStr(contentType)
	return &a, nil
}

// DeleteByKey removes asset metadata by storage key.
func (r *AssetRepo) DeleteByKey(ctx context.Context, key string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM assets WHERE key = :1`, key)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
