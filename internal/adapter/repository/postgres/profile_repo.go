package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/walfa-labs/backend/internal/domain"
)

// ProfileRepo implements port.ProfileRepo against PostgreSQL.
type ProfileRepo struct {
	db *sql.DB
}

// NewProfileRepo constructs a ProfileRepo bound to the given pool.
func NewProfileRepo(db *sql.DB) *ProfileRepo {
	return &ProfileRepo{db: db}
}

// profileColumns lists the profiles columns in scan order.
const profileColumns = `name, email, tagline, bio_markdown, location,
	avatar_url, github_url, linkedin_url, twitter_url, updated_at`

func scanProfile(row rowScanner) (*domain.Profile, error) {
	var p domain.Profile
	var name, email, location, avatarURL, githubURL, linkedinURL, twitterURL sql.NullString
	err := row.Scan(
		&name, &email, &p.Tagline, &p.BioMarkdown, &location,
		&avatarURL, &githubURL, &linkedinURL, &twitterURL, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Name = nullStr(name)
	p.Email = nullStr(email)
	p.Location = nullStr(location)
	p.AvatarURL = nullStr(avatarURL)
	p.GitHubURL = nullStr(githubURL)
	p.LinkedInURL = nullStr(linkedinURL)
	p.TwitterURL = nullStr(twitterURL)
	return &p, nil
}

// Get returns the singleton profile row. If no row exists, it returns a
// zero-value Profile with empty strings (NOT ErrNotFound) so callers can
// serve a 200 with empty fields.
func (r *ProfileRepo) Get(ctx context.Context) (*domain.Profile, error) {
	p, err := scanProfile(r.db.QueryRowContext(ctx,
		`SELECT `+profileColumns+` FROM profiles WHERE profile_id = 1`))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.Profile{}, nil
		}
		return nil, err
	}
	return p, nil
}

// Upsert creates or updates the singleton profile row using PostgreSQL
// INSERT ... ON CONFLICT DO UPDATE instead of Oracle MERGE INTO.
func (r *ProfileRepo) Upsert(ctx context.Context, p *domain.Profile) error {
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO profiles
		    (profile_id, name, email, tagline, bio_markdown, location,
		     avatar_url, github_url, linkedin_url, twitter_url)
		VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (profile_id) DO UPDATE SET
		    name         = EXCLUDED.name,
		    email        = EXCLUDED.email,
		    tagline      = EXCLUDED.tagline,
		    bio_markdown = EXCLUDED.bio_markdown,
		    location     = EXCLUDED.location,
		    avatar_url   = EXCLUDED.avatar_url,
		    github_url   = EXCLUDED.github_url,
		    linkedin_url = EXCLUDED.linkedin_url,
		    twitter_url  = EXCLUDED.twitter_url,
		    updated_at   = NOW()`,
		p.Name, p.Email, p.Tagline, p.BioMarkdown, p.Location,
		p.AvatarURL, p.GitHubURL, p.LinkedInURL, p.TwitterURL,
	); err != nil {
		return err
	}

	updated, err := scanProfile(r.db.QueryRowContext(ctx,
		`SELECT `+profileColumns+` FROM profiles WHERE profile_id = 1`))
	if err != nil {
		return err
	}
	*p = *updated
	return nil
}
