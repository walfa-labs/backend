package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

// PostRepo implements port.PostRepo against PostgreSQL.
type PostRepo struct {
	db *sql.DB
}

// NewPostRepo constructs a PostRepo bound to the given pool.
func NewPostRepo(db *sql.DB) *PostRepo {
	return &PostRepo{db: db}
}

const postColumns = `blog_post_id, slug, title, excerpt, body_markdown, cover_image_url, status,
	view_count, published_at, created_at, updated_at`

func scanPost(row rowScanner) (*domain.BlogPost, error) {
	var p domain.BlogPost
	var excerpt, coverURL sql.NullString
	var publishedAt sql.NullTime
	err := row.Scan(
		&p.ID, &p.Slug, &p.Title, &excerpt, &p.BodyMarkdown, &coverURL,
		&p.Status, &p.ViewCount, &publishedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Excerpt = nullStr(excerpt)
	p.CoverImageURL = nullStr(coverURL)
	p.PublishedAt = nullTime(publishedAt)
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

	const summaryColumns = `blog_post_id, slug, title, excerpt, cover_image_url, published_at`

	if filter.Tag == "" {
		rows, err := r.db.QueryContext(ctx, `
			SELECT `+summaryColumns+`
			FROM blog_posts
			WHERE status = 'published'
			ORDER BY published_at DESC NULLS LAST
			OFFSET $1 LIMIT $2`, offset, perPage)
		if err != nil {
			return nil, err
		}
		defer func() { _ = rows.Close() }()
		return r.scanSummaries(ctx, rows)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT bp.blog_post_id, bp.slug, bp.title, bp.excerpt, bp.cover_image_url, bp.published_at
		FROM blog_posts bp
		JOIN post_tags pt ON pt.blog_post_id = bp.blog_post_id
		JOIN tags t ON t.tag_id = pt.tag_id
		WHERE bp.status = 'published' AND t.slug = $1
		ORDER BY bp.published_at DESC NULLS LAST
		OFFSET $2 LIMIT $3`, filter.Tag, offset, perPage)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return r.scanSummaries(ctx, rows)
}

func (r *PostRepo) scanSummaries(ctx context.Context, rows *sql.Rows) ([]port.PostSummary, error) {
	var out []port.PostSummary
	for rows.Next() {
		var s port.PostSummary
		var excerpt, coverURL sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&s.ID, &s.Slug, &s.Title, &excerpt, &coverURL, &publishedAt); err != nil {
			return nil, err
		}
		s.Excerpt = nullStr(excerpt)
		s.CoverImageURL = nullStr(coverURL)
		s.PublishedAt = nullTime(publishedAt)
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range out {
		tags, err := r.fetchTags(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Tags = tags
	}
	return out, nil
}

func (r *PostRepo) fetchTags(ctx context.Context, postID uuid.UUID) ([]domain.Tag, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.tag_id, t.name, t.slug
		FROM tags t
		JOIN post_tags pt ON pt.tag_id = t.tag_id
		WHERE pt.blog_post_id = $1
		ORDER BY t.name ASC`, postID.String())
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

// ListAll returns all posts including drafts, ordered by created_at desc.
func (r *PostRepo) ListAll(ctx context.Context) ([]domain.BlogPost, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+postColumns+` FROM blog_posts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []domain.BlogPost
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		p.Tags, err = r.fetchTags(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// GetPublishedBySlug returns a published post with its tags.
func (r *PostRepo) GetPublishedBySlug(ctx context.Context, slug string) (*domain.BlogPost, error) {
	p, err := scanPost(r.db.QueryRowContext(ctx,
		`SELECT `+postColumns+` FROM blog_posts WHERE slug = $1 AND status = 'published'`, slug))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	p.Tags, err = r.fetchTags(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// GetByID returns a post with its tags (any status).
func (r *PostRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.BlogPost, error) {
	p, err := scanPost(r.db.QueryRowContext(ctx,
		`SELECT `+postColumns+` FROM blog_posts WHERE blog_post_id = $1`, id.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	p.Tags, err = r.fetchTags(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Create inserts a blog post and syncs its tag associations in a transaction.
func (r *PostRepo) Create(ctx context.Context, p *domain.BlogPost) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO blog_posts
		    (blog_post_id, slug, title, excerpt, body_markdown, cover_image_url, status,
		     view_count, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID.String(), p.Slug, p.Title, p.Excerpt, p.BodyMarkdown, p.CoverImageURL,
		string(p.Status), p.ViewCount, p.PublishedAt,
	)
	if err != nil {
		return err
	}

	if err := r.syncPostTags(ctx, tx, p.ID, p.Tags); err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx,
		`SELECT created_at, updated_at FROM blog_posts WHERE blog_post_id = $1`, p.ID.String(),
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Update modifies a blog post and re-syncs its tag associations.
func (r *PostRepo) Update(ctx context.Context, p *domain.BlogPost) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		UPDATE blog_posts SET
		    slug = $1, title = $2, excerpt = $3, body_markdown = $4,
		    cover_image_url = $5, status = $6, view_count = $7, published_at = $8,
		    updated_at = NOW()
		WHERE blog_post_id = $9`,
		p.Slug, p.Title, p.Excerpt, p.BodyMarkdown, p.CoverImageURL,
		string(p.Status), p.ViewCount, p.PublishedAt, p.ID.String(),
	)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return domain.ErrNotFound
	}

	if err := r.syncPostTags(ctx, tx, p.ID, p.Tags); err != nil {
		return err
	}

	if err := tx.QueryRowContext(ctx,
		`SELECT updated_at FROM blog_posts WHERE blog_post_id = $1`, p.ID.String(),
	).Scan(&p.UpdatedAt); err != nil {
		return err
	}
	return tx.Commit()
}

// Delete removes a blog post; post_tags associations cascade.
func (r *PostRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM blog_posts WHERE blog_post_id = $1`, id.String())
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// IncrementViewCount atomically bumps view_count.
func (r *PostRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE blog_posts SET view_count = view_count + 1 WHERE blog_post_id = $1`, id.String())
	return err
}

// SumViews returns the total view count across all posts.
// Uses COALESCE instead of Oracle's NVL.
func (r *PostRepo) SumViews(ctx context.Context) (int64, error) {
	var sum int64
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(view_count), 0) FROM blog_posts`).Scan(&sum)
	return sum, err
}

// CountPublished returns the number of published posts, optionally filtered by tag slug.
func (r *PostRepo) CountPublished(ctx context.Context, filter port.PostFilter) (int64, error) {
	if filter.Tag == "" {
		var count int64
		err := r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM blog_posts WHERE status = 'published'`).Scan(&count)
		return count, err
	}

	var count int64
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT bp.blog_post_id)
		FROM blog_posts bp
		JOIN post_tags pt ON pt.blog_post_id = bp.blog_post_id
		JOIN tags t ON t.tag_id = pt.tag_id
		WHERE bp.status = 'published' AND t.slug = $1`, filter.Tag).Scan(&count)
	return count, err
}

// syncPostTags replaces a post's tag associations.
func (r *PostRepo) syncPostTags(ctx context.Context, tx *sql.Tx, postID uuid.UUID, tags []domain.Tag) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM post_tags WHERE blog_post_id = $1`, postID.String()); err != nil {
		return err
	}
	for _, t := range tags {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO post_tags (blog_post_id, tag_id) VALUES ($1, $2)`, postID.String(), t.ID.String())
		if err != nil {
			return err
		}
	}
	return nil
}
