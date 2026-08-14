package service

import (
	"context"
	"strings"

	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

// ProfileService implements read/upsert for the singleton profile.
type ProfileService struct {
	repo port.ProfileRepo
}

// NewProfileService constructs a ProfileService bound to the given repo.
func NewProfileService(repo port.ProfileRepo) *ProfileService {
	return &ProfileService{repo: repo}
}

// Get returns the singleton profile.
func (s *ProfileService) Get(ctx context.Context) (*domain.Profile, error) {
	return s.repo.Get(ctx)
}

// Update validates the name and upserts the singleton profile.
func (s *ProfileService) Update(ctx context.Context, input port.ProfileInput) (*domain.Profile, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, domain.NewValidationError("name", "required")
	}
	p := &domain.Profile{
		Name:        input.Name,
		Email:       input.Email,
		Tagline:     input.Tagline,
		BioMarkdown: input.BioMarkdown,
		Location:    input.Location,
		AvatarURL:   input.AvatarURL,
		GitHubURL:   input.GitHubURL,
		LinkedInURL: input.LinkedInURL,
		TwitterURL:  input.TwitterURL,
	}
	if err := s.repo.Upsert(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
