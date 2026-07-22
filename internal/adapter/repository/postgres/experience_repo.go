package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/walfa-labs/backend/internal/domain"
)

// ExperienceRepo implements port.ExperienceRepo against PostgreSQL.
type ExperienceRepo struct {
	pool *pgxpool.Pool
}

// NewExperienceRepo constructs an ExperienceRepo bound to the given pool.
func NewExperienceRepo(pool *pgxpool.Pool) *ExperienceRepo {
	return &ExperienceRepo{pool: pool}
}

// List returns all experiences ordered by sort_order.
func (r *ExperienceRepo) List(ctx context.Context) ([]domain.Experience, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, experience_type, organization, role_title, location,
		       start_date, end_date, current, summary_markdown, sort_order,
		       created_at, updated_at
		FROM experience
		ORDER BY sort_order ASC, start_date DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Experience
	for rows.Next() {
		var e domain.Experience
		var endDate *time.Time
		if err := rows.Scan(
			&e.ID, &e.ExperienceType, &e.Organization, &e.RoleTitle, &e.Location,
			&e.StartDate, &endDate, &e.Current, &e.SummaryMarkdown, &e.SortOrder,
			&e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		e.EndDate = endDate
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetByID returns a single experience with its highlights.
func (r *ExperienceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Experience, error) {
	var e domain.Experience
	var endDate *time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT id, experience_type, organization, role_title, location,
		       start_date, end_date, current, summary_markdown, sort_order,
		       created_at, updated_at
		FROM experience
		WHERE id = $1`, id).Scan(
		&e.ID, &e.ExperienceType, &e.Organization, &e.RoleTitle, &e.Location,
		&e.StartDate, &endDate, &e.Current, &e.SummaryMarkdown, &e.SortOrder,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	e.EndDate = endDate

	highlights, err := r.fetchHighlights(ctx, id)
	if err != nil {
		return nil, err
	}
	e.Highlights = highlights
	return &e, nil
}

// Create inserts an experience and its highlights in a single transaction.
func (r *ExperienceRepo) Create(ctx context.Context, e *domain.Experience) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO experience
		    (experience_type, organization, role_title, location, start_date,
		     end_date, current, summary_markdown, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`,
		e.ExperienceType, e.Organization, e.RoleTitle, e.Location, e.StartDate,
		e.EndDate, e.Current, e.SummaryMarkdown, e.SortOrder,
	).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return err
	}

	if err := r.replaceHighlights(ctx, tx, e.ID, e.Highlights); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Update modifies an experience and re-syncs its highlights.
func (r *ExperienceRepo) Update(ctx context.Context, e *domain.Experience) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE experience SET
		    experience_type = $1, organization = $2, role_title = $3,
		    location = $4, start_date = $5, end_date = $6, current = $7,
		    summary_markdown = $8, sort_order = $9, updated_at = now()
		WHERE id = $10`,
		e.ExperienceType, e.Organization, e.RoleTitle, e.Location, e.StartDate,
		e.EndDate, e.Current, e.SummaryMarkdown, e.SortOrder, e.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	if err := r.replaceHighlights(ctx, tx, e.ID, e.Highlights); err != nil {
		return err
	}

	if err := tx.QueryRow(ctx, `SELECT updated_at FROM experience WHERE id = $1`, e.ID).Scan(&e.UpdatedAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Delete removes an experience; highlights cascade.
func (r *ExperienceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM experience WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ExperienceRepo) fetchHighlights(ctx context.Context, experienceID uuid.UUID) ([]domain.ExperienceHighlight, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, experience_id, body_markdown, sort_order
		FROM experience_highlight
		WHERE experience_id = $1
		ORDER BY sort_order ASC`, experienceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.ExperienceHighlight
	for rows.Next() {
		var h domain.ExperienceHighlight
		if err := rows.Scan(&h.ID, &h.ExperienceID, &h.BodyMarkdown, &h.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (r *ExperienceRepo) replaceHighlights(ctx context.Context, tx pgx.Tx, experienceID uuid.UUID, highlights []domain.ExperienceHighlight) error {
	if _, err := tx.Exec(ctx, `DELETE FROM experience_highlight WHERE experience_id = $1`, experienceID); err != nil {
		return err
	}
	for _, h := range highlights {
		_, err := tx.Exec(ctx, `
			INSERT INTO experience_highlight (experience_id, body_markdown, sort_order)
			VALUES ($1, $2, $3)`,
			experienceID, h.BodyMarkdown, h.SortOrder)
		if err != nil {
			return err
		}
	}
	return nil
}
