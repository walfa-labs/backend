package atp

import (
	"context"
	"database/sql"
	"errors"

	"github.com/walfa-labs/backend/internal/domain"
)

// ProfileRepo implements port.ProfileRepo against Oracle ATP.
type ProfileRepo struct {
	db *sql.DB
}

// NewProfileRepo constructs a ProfileRepo bound to the given pool.
func NewProfileRepo(db *sql.DB) *ProfileRepo {
	return &ProfileRepo{db: db}
}

// profileColumns lists the profiles columns in scan order. All short text
// columns are nullable (Oracle stores ” as NULL); tagline/bio_markdown are
// NOT NULL CLOBs (see 0002_profile.up.sql).
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

// Upsert creates or updates the singleton profile row, returning the result.
func (r *ProfileRepo) Upsert(ctx context.Context, p *domain.Profile) error {
	if _, err := r.db.ExecContext(ctx, `
		MERGE INTO profiles p
		USING (SELECT 1 AS profile_id FROM dual) s
		ON (p.profile_id = s.profile_id)
		WHEN MATCHED THEN UPDATE SET
		    name = :1, email = :2, tagline = :3, bio_markdown = :4, location = :5,
		    avatar_url = :6, github_url = :7, linkedin_url = :8, twitter_url = :9,
		    updated_at = CURRENT_TIMESTAMP
		WHEN NOT MATCHED THEN INSERT
		    (profile_id, name, email, tagline, bio_markdown, location,
		     avatar_url, github_url, linkedin_url, twitter_url)
		VALUES (1, :1, :2, :3, :4, :5, :6, :7, :8, :9)`,
		p.Name, p.Email, clob(p.Tagline), clob(p.BioMarkdown), p.Location,
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
