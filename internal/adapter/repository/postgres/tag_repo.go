package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/walfa-labs/backend/internal/domain"
)

// TagRepo implements port.TagRepo against PostgreSQL.
type TagRepo struct {
	pool *pgxpool.Pool
}

// NewTagRepo constructs a TagRepo bound to the given pool.
func NewTagRepo(pool *pgxpool.Pool) *TagRepo {
	return &TagRepo{pool: pool}
}

// List returns all tags ordered by name.
func (r *TagRepo) List(ctx context.Context) ([]domain.Tag, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, slug FROM tag ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
func (r *TagRepo) GetOrCreate(ctx context.Context, name, slug string) (*domain.Tag, error) {
	var t domain.Tag
	err := r.pool.QueryRow(ctx, `
		INSERT INTO tag (name, slug)
		VALUES ($1, $2)
		ON CONFLICT (slug) DO NOTHING
		RETURNING id, name, slug`, name, slug).Scan(&t.ID, &t.Name, &t.Slug)
	if err == nil {
		return &t, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	// Row was not inserted because a tag with this slug already exists.
	err = r.pool.QueryRow(ctx, `SELECT id, name, slug FROM tag WHERE slug = $1`, slug).Scan(&t.ID, &t.Name, &t.Slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}
