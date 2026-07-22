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

// ProjectRepo implements port.ProjectRepo against PostgreSQL.
type ProjectRepo struct {
	pool *pgxpool.Pool
}

// NewProjectRepo constructs a ProjectRepo bound to the given pool.
func NewProjectRepo(pool *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{pool: pool}
}

const projectColumns = `id, slug, title, tagline, description_markdown, cover_image_url,
	repo_url, demo_url, tech_stack, status, featured, sort_order, published_at,
	created_at, updated_at`

func scanProject(row pgx.Row) (*domain.Project, error) {
	var p domain.Project
	var publishedAt *time.Time
	err := row.Scan(
		&p.ID, &p.Slug, &p.Title, &p.Tagline, &p.DescriptionMarkdown, &p.CoverImageURL,
		&p.RepoURL, &p.DemoURL, &p.TechStack, &p.Status, &p.Featured, &p.SortOrder,
		&publishedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.PublishedAt = publishedAt
	return &p, nil
}

// ListPublished returns published projects, optionally filtered by featured.
func (r *ProjectRepo) ListPublished(ctx context.Context, filter port.ProjectFilter) ([]domain.Project, error) {
	q := `SELECT ` + projectColumns + ` FROM project WHERE status = 'published'`
	args := []interface{}{}
	if filter.HasFeat {
		q += ` AND featured = $1`
		args = append(args, filter.Featured)
	}
	q += ` ORDER BY sort_order ASC, published_at DESC NULLS LAST`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		p.Links, err = r.fetchLinks(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// ListAll returns all projects including drafts, ordered by sort_order.
func (r *ProjectRepo) ListAll(ctx context.Context) ([]domain.Project, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+projectColumns+` FROM project ORDER BY sort_order ASC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		p.Links, err = r.fetchLinks(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// GetBySlug returns a single project with its links.
func (r *ProjectRepo) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	p, err := scanProject(r.pool.QueryRow(ctx, `SELECT `+projectColumns+` FROM project WHERE slug = $1`, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	p.Links, err = r.fetchLinks(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// GetByID returns a single project with its links.
func (r *ProjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	p, err := scanProject(r.pool.QueryRow(ctx, `SELECT `+projectColumns+` FROM project WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	p.Links, err = r.fetchLinks(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Create inserts a project and its links in a single transaction.
func (r *ProjectRepo) Create(ctx context.Context, p *domain.Project) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO project
		    (slug, title, tagline, description_markdown, cover_image_url,
		     repo_url, demo_url, tech_stack, status, featured, sort_order, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`,
		p.Slug, p.Title, p.Tagline, p.DescriptionMarkdown, p.CoverImageURL,
		p.RepoURL, p.DemoURL, p.TechStack, p.Status, p.Featured, p.SortOrder, p.PublishedAt,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	if err := r.replaceLinks(ctx, tx, p.ID, p.Links); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Update modifies a project and re-syncs its links.
func (r *ProjectRepo) Update(ctx context.Context, p *domain.Project) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE project SET
		    slug = $1, title = $2, tagline = $3, description_markdown = $4,
		    cover_image_url = $5, repo_url = $6, demo_url = $7, tech_stack = $8,
		    status = $9, featured = $10, sort_order = $11, published_at = $12,
		    updated_at = now()
		WHERE id = $13`,
		p.Slug, p.Title, p.Tagline, p.DescriptionMarkdown, p.CoverImageURL,
		p.RepoURL, p.DemoURL, p.TechStack, p.Status, p.Featured, p.SortOrder,
		p.PublishedAt, p.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	if err := r.replaceLinks(ctx, tx, p.ID, p.Links); err != nil {
		return err
	}

	if err := tx.QueryRow(ctx, `SELECT updated_at FROM project WHERE id = $1`, p.ID).Scan(&p.UpdatedAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Delete removes a project; links cascade.
func (r *ProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM project WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProjectRepo) fetchLinks(ctx context.Context, projectID uuid.UUID) ([]domain.ProjectLink, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, project_id, label, url, kind
		FROM project_link
		WHERE project_id = $1
		ORDER BY id ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.ProjectLink
	for rows.Next() {
		var l domain.ProjectLink
		if err := rows.Scan(&l.ID, &l.ProjectID, &l.Label, &l.URL, &l.Kind); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *ProjectRepo) replaceLinks(ctx context.Context, tx pgx.Tx, projectID uuid.UUID, links []domain.ProjectLink) error {
	if _, err := tx.Exec(ctx, `DELETE FROM project_link WHERE project_id = $1`, projectID); err != nil {
		return err
	}
	for _, l := range links {
		_, err := tx.Exec(ctx, `
			INSERT INTO project_link (project_id, label, url, kind)
			VALUES ($1, $2, $3, $4)`,
			projectID, l.Label, l.URL, l.Kind)
		if err != nil {
			return err
		}
	}
	return nil
}
