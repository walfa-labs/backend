package atp

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
)

// ExperienceRepo implements port.ExperienceRepo against Oracle ATP.
type ExperienceRepo struct {
	db *sql.DB
}

// NewExperienceRepo constructs an ExperienceRepo bound to the given pool.
func NewExperienceRepo(db *sql.DB) *ExperienceRepo {
	return &ExperienceRepo{db: db}
}

// experienceColumns lists the experiences columns in scan order. "current" is
// an Oracle reserved word and must stay double-quoted.
const experienceColumns = `experience_id, experience_type, organization, role_title, location,
	start_date, end_date, "current", summary_markdown, sort_order, created_at, updated_at`

func scanExperience(row rowScanner) (*domain.Experience, error) {
	var e domain.Experience
	var org, role, loc sql.NullString
	var endDate sql.NullTime
	var current int
	err := row.Scan(
		&e.ID, &e.ExperienceType, &org, &role, &loc,
		&e.StartDate, &endDate, &current, &e.SummaryMarkdown, &e.SortOrder,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	e.Organization = nullStr(org)
	e.RoleTitle = nullStr(role)
	e.Location = nullStr(loc)
	e.EndDate = nullTime(endDate)
	e.Current = i2b(current)
	return &e, nil
}

// List returns all experiences ordered by sort_order.
func (r *ExperienceRepo) List(ctx context.Context) ([]domain.Experience, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+experienceColumns+` FROM experiences ORDER BY sort_order ASC, start_date DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []domain.Experience
	var ids []uuid.UUID
	for rows.Next() {
		e, err := scanExperience(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *e)
		ids = append(ids, e.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return out, nil
	}

	hlMap, err := r.fetchHighlightsForExperiences(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		if hl, ok := hlMap[out[i].ID]; ok {
			out[i].Highlights = hl
		} else {
			out[i].Highlights = []domain.ExperienceHighlight{}
		}
	}
	return out, nil
}

// GetByID returns a single experience with its highlights.
func (r *ExperienceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Experience, error) {
	e, err := scanExperience(r.db.QueryRowContext(ctx,
		`SELECT `+experienceColumns+` FROM experiences WHERE experience_id = :1`, id.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	highlights, err := r.fetchHighlights(ctx, id)
	if err != nil {
		return nil, err
	}
	e.Highlights = highlights
	return e, nil
}

// Create inserts an experience and its highlights in a single transaction.
func (r *ExperienceRepo) Create(ctx context.Context, e *domain.Experience) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO experiences
		    (experience_id, experience_type, organization, role_title, location, start_date,
		     end_date, "current", summary_markdown, sort_order)
		VALUES (:1, :2, :3, :4, :5, :6, :7, :8, :9, :10)`,
		e.ID.String(), string(e.ExperienceType), e.Organization, e.RoleTitle, e.Location, e.StartDate,
		e.EndDate, b2i(e.Current), clob(e.SummaryMarkdown), e.SortOrder,
	)
	if err != nil {
		return err
	}

	if err := r.replaceHighlights(ctx, tx, e.ID, e.Highlights); err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx,
		`SELECT created_at, updated_at FROM experiences WHERE experience_id = :1`, e.ID.String(),
	).Scan(&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Update modifies an experience and re-syncs its highlights.
func (r *ExperienceRepo) Update(ctx context.Context, e *domain.Experience) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		UPDATE experiences SET
		    experience_type = :1, organization = :2, role_title = :3,
		    location = :4, start_date = :5, end_date = :6, "current" = :7,
		    summary_markdown = :8, sort_order = :9, updated_at = CURRENT_TIMESTAMP
		WHERE experience_id = :10`,
		string(e.ExperienceType), e.Organization, e.RoleTitle, e.Location, e.StartDate,
		e.EndDate, b2i(e.Current), clob(e.SummaryMarkdown), e.SortOrder, e.ID.String(),
	)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return domain.ErrNotFound
	}

	if err := r.replaceHighlights(ctx, tx, e.ID, e.Highlights); err != nil {
		return err
	}

	if err := tx.QueryRowContext(ctx,
		`SELECT updated_at FROM experiences WHERE experience_id = :1`, e.ID.String(),
	).Scan(&e.UpdatedAt); err != nil {
		return err
	}
	return tx.Commit()
}

// Delete removes an experience; highlights cascade.
func (r *ExperienceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM experiences WHERE experience_id = :1`, id.String())
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ExperienceRepo) fetchHighlightsForExperiences(ctx context.Context, expIDs []uuid.UUID) (map[uuid.UUID][]domain.ExperienceHighlight, error) {
	result := make(map[uuid.UUID][]domain.ExperienceHighlight, len(expIDs))
	if len(expIDs) == 0 {
		return result, nil
	}

	args := make([]any, len(expIDs))
	for i, id := range expIDs {
		args[i] = id.String()
		result[id] = []domain.ExperienceHighlight{}
	}

	// #nosec G202 -- SQL query uses parameterized positional placeholders
	q := `
		SELECT experience_highlight_id, experience_id, body_markdown, sort_order
		FROM experience_highlights
		WHERE experience_id IN (` + inPlaceholders(len(expIDs), 1) + `)
		ORDER BY sort_order ASC`

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var h domain.ExperienceHighlight
		var expIDStr string
		if err := rows.Scan(&h.ID, &expIDStr, &h.BodyMarkdown, &h.SortOrder); err != nil {
			return nil, err
		}
		expID, err := uuid.Parse(expIDStr)
		if err == nil {
			h.ExperienceID = expID
			result[expID] = append(result[expID], h)
		}
	}
	return result, rows.Err()
}

func (r *ExperienceRepo) fetchHighlights(ctx context.Context, experienceID uuid.UUID) ([]domain.ExperienceHighlight, error) {
	m, err := r.fetchHighlightsForExperiences(ctx, []uuid.UUID{experienceID})
	if err != nil {
		return nil, err
	}
	if highlights, ok := m[experienceID]; ok {
		return highlights, nil
	}
	return []domain.ExperienceHighlight{}, nil
}

func (r *ExperienceRepo) replaceHighlights(ctx context.Context, tx *sql.Tx, experienceID uuid.UUID, highlights []domain.ExperienceHighlight) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM experience_highlights WHERE experience_id = :1`, experienceID.String()); err != nil {
		return err
	}
	for _, h := range highlights {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO experience_highlights (experience_highlight_id, experience_id, body_markdown, sort_order)
			VALUES (:1, :2, :3, :4)`,
			uuid.New().String(), experienceID.String(), clob(h.BodyMarkdown), h.SortOrder)
		if err != nil {
			return err
		}
	}
	return nil
}
