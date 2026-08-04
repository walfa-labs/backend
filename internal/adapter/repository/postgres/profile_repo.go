package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/walfa-labs/backend/internal/domain"
)

// ProfileRepo implements port.ProfileRepo against PostgreSQL.
type ProfileRepo struct {
	pool *pgxpool.Pool
}

// NewProfileRepo constructs a ProfileRepo bound to the given pool.
func NewProfileRepo(pool *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{pool: pool}
}

const profileColumns = `name, email, tagline, bio_markdown, location,
	avatar_url, github_url, linkedin_url, twitter_url, updated_at`

func scanProfile(row pgx.Row) (*domain.Profile, error) {
	var p domain.Profile
	err := row.Scan(
		&p.Name, &p.Email, &p.Tagline, &p.BioMarkdown, &p.Location,
		&p.AvatarURL, &p.GitHubURL, &p.LinkedInURL, &p.TwitterURL, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Get returns the singleton profile row. If no row exists, it returns a
// zero-value Profile with empty strings (NOT ErrNotFound) so callers can
// serve a 200 with empty fields.
func (r *ProfileRepo) Get(ctx context.Context) (*domain.Profile, error) {
	p, err := scanProfile(r.pool.QueryRow(ctx, `SELECT `+profileColumns+` FROM profiles WHERE profile_id = 1`))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.Profile{}, nil
		}
		return nil, err
	}
	return p, nil
}

// Upsert creates or updates the singleton profile row, returning the result.
func (r *ProfileRepo) Upsert(ctx context.Context, p *domain.Profile) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO profiles (profile_id, name, email, tagline, bio_markdown, location,
		    avatar_url, github_url, linkedin_url, twitter_url)
		VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (profile_id) DO UPDATE SET
		    name = EXCLUDED.name,
		    email = EXCLUDED.email,
		    tagline = EXCLUDED.tagline,
		    bio_markdown = EXCLUDED.bio_markdown,
		    location = EXCLUDED.location,
		    avatar_url = EXCLUDED.avatar_url,
		    github_url = EXCLUDED.github_url,
		    linkedin_url = EXCLUDED.linkedin_url,
		    twitter_url = EXCLUDED.twitter_url,
		    updated_at = now()
		RETURNING `+profileColumns,
		p.Name, p.Email, p.Tagline, p.BioMarkdown, p.Location,
		p.AvatarURL, p.GitHubURL, p.LinkedInURL, p.TwitterURL,
	).Scan(
		&p.Name, &p.Email, &p.Tagline, &p.BioMarkdown, &p.Location,
		&p.AvatarURL, &p.GitHubURL, &p.LinkedInURL, &p.TwitterURL, &p.UpdatedAt,
	)
}
