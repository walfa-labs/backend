package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

// PostRepo implements port.PostRepo against PostgreSQL.
type PostRepo struct {
	pool *pgxpool.Pool
}

// NewPostRepo constructs a PostRepo bound to the given pool.
func NewPostRepo(pool *pgxpool.Pool) *PostRepo {
	return &PostRepo{pool: pool}
}

const postColumns = `id, slug, title, COALESCE(excerpt, ''), COALESCE(body_markdown, ''), COALESCE(cover_image_url, ''), status,
	view_count, published_at, created_at, updated_at`

func scanPost(row pgx.Row) (*domain.BlogPost, error) {
	var p domain.BlogPost
	var publishedAt *time.Time
	err := row.Scan(
		&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.BodyMarkdown, &p.CoverImageURL,
		&p.Status, &p.ViewCount, &publishedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.PublishedAt = publishedAt
	return &p, nil
}

// ListPublished returns a paginated list of PostSummary for published posts,
// optionally filtered by tag slug.
func (r *PostRepo) ListPublished(ctx context.Context, filter port.PostFilter) ([]port.PostSummary, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	if filter.Tag == "" {
		rows, err := r.pool.Query(ctx, `
			SELECT id, slug, title, COALESCE(excerpt, ''), COALESCE(cover_image_url, ''), published_at
			FROM blog_post
			WHERE status = 'published'
			ORDER BY published_at DESC NULLS LAST
			LIMIT $1 OFFSET $2`, perPage, offset)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanSummaries(ctx, r.pool, rows)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT bp.id, bp.slug, bp.title, COALESCE(bp.excerpt, ''), COALESCE(bp.cover_image_url, ''), bp.published_at
		FROM blog_post bp
		JOIN post_tag pt ON pt.post_id = bp.id
		JOIN tag t ON t.id = pt.tag_id
		WHERE bp.status = 'published' AND t.slug = $1
		ORDER BY bp.published_at DESC NULLS LAST
		LIMIT $2 OFFSET $3`, filter.Tag, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSummaries(ctx, r.pool, rows)
}

func scanSummaries(ctx context.Context, pool interface {
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
}, rows pgx.Rows) ([]port.PostSummary, error) {
	var out []port.PostSummary
	for rows.Next() {
		var s port.PostSummary
		var publishedAt *time.Time
		if err := rows.Scan(&s.ID, &s.Slug, &s.Title, &s.Excerpt, &s.CoverImageURL, &publishedAt); err != nil {
			return nil, err
		}
		s.PublishedAt = publishedAt
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range out {
		tags, err := fetchTags(ctx, pool, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Tags = tags
	}
	return out, nil
}

func fetchTags(ctx context.Context, pool interface {
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
}, postID uuid.UUID) ([]domain.Tag, error) {
	rows, err := pool.Query(ctx, `
		SELECT t.id, t.name, t.slug
		FROM tag t
		JOIN post_tag pt ON pt.tag_id = t.id
		WHERE pt.post_id = $1
		ORDER BY t.name ASC`, postID)
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

// ListAll returns all posts including drafts, ordered by created_at desc.
func (r *PostRepo) ListAll(ctx context.Context) ([]domain.BlogPost, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+postColumns+` FROM blog_post ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.BlogPost
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		p.Tags, err = fetchTags(ctx, r.pool, p.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// GetPublishedBySlug returns a published post with its tags.
func (r *PostRepo) GetPublishedBySlug(ctx context.Context, slug string) (*domain.BlogPost, error) {
	p, err := scanPost(r.pool.QueryRow(ctx, `SELECT `+postColumns+` FROM blog_post WHERE slug = $1 AND status = 'published'`, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	p.Tags, err = fetchTags(ctx, r.pool, p.ID)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// GetByID returns a post with its tags (any status).
func (r *PostRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.BlogPost, error) {
	p, err := scanPost(r.pool.QueryRow(ctx, `SELECT `+postColumns+` FROM blog_post WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	p.Tags, err = fetchTags(ctx, r.pool, p.ID)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Create inserts a blog post and syncs its tag associations in a transaction.
func (r *PostRepo) Create(ctx context.Context, p *domain.BlogPost) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO blog_post
		    (slug, title, excerpt, body_markdown, cover_image_url, status,
		     view_count, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`,
		p.Slug, p.Title, p.Excerpt, p.BodyMarkdown, p.CoverImageURL, p.Status,
		p.ViewCount, p.PublishedAt,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	if err := r.syncPostTags(ctx, tx, p.ID, p.Tags); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Update modifies a blog post and re-syncs its tag associations.
func (r *PostRepo) Update(ctx context.Context, p *domain.BlogPost) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE blog_post SET
		    slug = $1, title = $2, excerpt = $3, body_markdown = $4,
		    cover_image_url = $5, status = $6, view_count = $7, published_at = $8,
		    updated_at = now()
		WHERE id = $9`,
		p.Slug, p.Title, p.Excerpt, p.BodyMarkdown, p.CoverImageURL, p.Status,
		p.ViewCount, p.PublishedAt, p.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	if err := r.syncPostTags(ctx, tx, p.ID, p.Tags); err != nil {
		return err
	}

	if err := tx.QueryRow(ctx, `SELECT updated_at FROM blog_post WHERE id = $1`, p.ID).Scan(&p.UpdatedAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Delete removes a blog post; post_tag associations cascade.
func (r *PostRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM blog_post WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// IncrementViewCount atomically bumps view_count.
func (r *PostRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE blog_post SET view_count = view_count + 1 WHERE id = $1`, id)
	return err
}

// SumViews returns the total view count across all posts.
func (r *PostRepo) SumViews(ctx context.Context) (int64, error) {
	var sum int64
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(SUM(view_count), 0) FROM blog_post`).Scan(&sum)
	return sum, err
}

// CountPublished returns the number of published posts, optionally filtered
// by tag slug, using the same filter semantics as ListPublished.
func (r *PostRepo) CountPublished(ctx context.Context, filter port.PostFilter) (int64, error) {
	if filter.Tag == "" {
		var count int64
		err := r.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM blog_post WHERE status = 'published'`).Scan(&count)
		return count, err
	}

	var count int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT bp.id)
		FROM blog_post bp
		JOIN post_tag pt ON pt.post_id = bp.id
		JOIN tag t ON t.id = pt.tag_id
		WHERE bp.status = 'published' AND t.slug = $1`, filter.Tag).Scan(&count)
	return count, err
}

// syncPostTags replaces a post's tag associations.
func (r *PostRepo) syncPostTags(ctx context.Context, tx pgx.Tx, postID uuid.UUID, tags []domain.Tag) error {
	if _, err := tx.Exec(ctx, `DELETE FROM post_tag WHERE post_id = $1`, postID); err != nil {
		return err
	}
	for _, t := range tags {
		_, err := tx.Exec(ctx, `INSERT INTO post_tag (post_id, tag_id) VALUES ($1, $2)`, postID, t.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
