package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
)

// TagRepo implements port.TagRepo against PostgreSQL.
type TagRepo struct {
	db *sql.DB
}

// NewTagRepo constructs a TagRepo bound to the given pool.
func NewTagRepo(db *sql.DB) *TagRepo {
	return &TagRepo{db: db}
}

// List returns all tags ordered by name.
func (r *TagRepo) List(ctx context.Context) ([]domain.Tag, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT tag_id, name, slug FROM tags ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GetOrCreate inserts a tag with the given name/slug if absent and returns it.
// Uses PostgreSQL INSERT ... ON CONFLICT DO NOTHING instead of Oracle MERGE.
func (r *TagRepo) GetOrCreate(ctx context.Context, name, slug string) (*domain.Tag, error) {
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO tags (tag_id, name, slug)
		VALUES ($1, $2, $3)
		ON CONFLICT (slug) DO NOTHING`,
		uuid.New().String(), name, slug); err != nil {
		return nil, err
	}

	var t domain.Tag
	err := r.db.QueryRowContext(ctx,
		`SELECT tag_id, name, slug FROM tags WHERE slug = $1`, slug).Scan(&t.ID, &t.Name, &t.Slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}
