package atp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

// ProjectRepo implements port.ProjectRepo against Oracle ATP.
type ProjectRepo struct {
	db *sql.DB
}

// NewProjectRepo constructs a ProjectRepo bound to the given pool.
func NewProjectRepo(db *sql.DB) *ProjectRepo {
	return &ProjectRepo{db: db}
}

const projectColumns = `project_id, slug, title, tagline, description_markdown, cover_image_url,
	repo_url, demo_url, tech_stack, status, featured, sort_order, published_at,
	created_at, updated_at`

func scanProject(row rowScanner) (*domain.Project, error) {
	var p domain.Project
	var tagline, coverURL, repoURL, demoURL sql.NullString
	var techStackJSON string
	var featured int
	var publishedAt sql.NullTime
	err := row.Scan(
		&p.ID, &p.Slug, &p.Title, &tagline, &p.DescriptionMarkdown, &coverURL,
		&repoURL, &demoURL, &techStackJSON, &p.Status, &featured, &p.SortOrder,
		&publishedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Tagline = nullStr(tagline)
	p.CoverImageURL = nullStr(coverURL)
	p.RepoURL = nullStr(repoURL)
	p.DemoURL = nullStr(demoURL)
	p.Featured = i2b(featured)
	p.PublishedAt = nullTime(publishedAt)

	techStack, err := techStackFromJSON(techStackJSON)
	if err != nil {
		return nil, err
	}
	p.TechStack = techStack
	return &p, nil
}

// techStackToJSON encodes the tech stack for the VARCHAR2(4000) IS JSON column.
func techStackToJSON(stack []string) (string, error) {
	if len(stack) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal(stack)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func techStackFromJSON(s string) ([]string, error) {
	if s == "" {
		return []string{}, nil
	}
	var out []string
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListPublished returns published projects, optionally filtered by featured.
func (r *ProjectRepo) ListPublished(ctx context.Context, filter port.ProjectFilter) ([]domain.Project, error) {
	q := `SELECT ` + projectColumns + ` FROM projects WHERE status = 'published'`
	var args []any
	if filter.HasFeat {
		q += ` AND featured = :1`
		args = append(args, b2i(filter.Featured))
	}
	q += ` ORDER BY sort_order ASC, published_at DESC NULLS LAST`

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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
	rows, err := r.db.QueryContext(ctx, `SELECT `+projectColumns+` FROM projects ORDER BY sort_order ASC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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
	p, err := scanProject(r.db.QueryRowContext(ctx, `SELECT `+projectColumns+` FROM projects WHERE slug = :1`, slug))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
	p, err := scanProject(r.db.QueryRowContext(ctx, `SELECT `+projectColumns+` FROM projects WHERE project_id = :1`, id.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
	techStack, err := techStackToJSON(p.TechStack)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO projects
		    (project_id, slug, title, tagline, description_markdown, cover_image_url,
		     repo_url, demo_url, tech_stack, status, featured, sort_order, published_at)
		VALUES (:1, :2, :3, :4, :5, :6, :7, :8, :9, :10, :11, :12, :13)`,
		p.ID.String(), p.Slug, p.Title, clob(p.Tagline), clob(p.DescriptionMarkdown), p.CoverImageURL,
		p.RepoURL, p.DemoURL, techStack, string(p.Status), b2i(p.Featured), p.SortOrder, p.PublishedAt,
	)
	if err != nil {
		return err
	}

	if err := r.replaceLinks(ctx, tx, p.ID, p.Links); err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx,
		`SELECT created_at, updated_at FROM projects WHERE project_id = :1`, p.ID.String(),
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Update modifies a project and re-syncs its links.
func (r *ProjectRepo) Update(ctx context.Context, p *domain.Project) error {
	techStack, err := techStackToJSON(p.TechStack)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		UPDATE projects SET
		    slug = :1, title = :2, tagline = :3, description_markdown = :4,
		    cover_image_url = :5, repo_url = :6, demo_url = :7, tech_stack = :8,
		    status = :9, featured = :10, sort_order = :11, published_at = :12,
		    updated_at = CURRENT_TIMESTAMP
		WHERE project_id = :13`,
		p.Slug, p.Title, clob(p.Tagline), clob(p.DescriptionMarkdown), p.CoverImageURL,
		p.RepoURL, p.DemoURL, techStack, string(p.Status), b2i(p.Featured), p.SortOrder,
		p.PublishedAt, p.ID.String(),
	)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return domain.ErrNotFound
	}

	if err := r.replaceLinks(ctx, tx, p.ID, p.Links); err != nil {
		return err
	}

	if err := tx.QueryRowContext(ctx,
		`SELECT updated_at FROM projects WHERE project_id = :1`, p.ID.String(),
	).Scan(&p.UpdatedAt); err != nil {
		return err
	}
	return tx.Commit()
}

// Delete removes a project; links cascade.
func (r *ProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE project_id = :1`, id.String())
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProjectRepo) fetchLinks(ctx context.Context, projectID uuid.UUID) ([]domain.ProjectLink, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT project_link_id, project_id, label, url, kind
		FROM project_links
		WHERE project_id = :1
		ORDER BY project_link_id ASC`, projectID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []domain.ProjectLink
	for rows.Next() {
		var l domain.ProjectLink
		var label, url sql.NullString
		if err := rows.Scan(&l.ID, &l.ProjectID, &label, &url, &l.Kind); err != nil {
			return nil, err
		}
		l.Label = nullStr(label)
		l.URL = nullStr(url)
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *ProjectRepo) replaceLinks(ctx context.Context, tx *sql.Tx, projectID uuid.UUID, links []domain.ProjectLink) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM project_links WHERE project_id = :1`, projectID.String()); err != nil {
		return err
	}
	for _, l := range links {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO project_links (project_link_id, project_id, label, url, kind)
			VALUES (:1, :2, :3, :4, :5)`,
			uuid.New().String(), projectID.String(), l.Label, l.URL, string(l.Kind))
		if err != nil {
			return err
		}
	}
	return nil
}
