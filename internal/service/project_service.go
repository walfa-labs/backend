package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

type ProjectService struct {
	repo port.ProjectRepo
}

func NewProjectService(repo port.ProjectRepo) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) ListPublished(ctx context.Context, featured *bool) ([]domain.Project, error) {
	f := port.ProjectFilter{}
	if featured != nil {
		f.Featured = *featured
		f.HasFeat = true
	}
	return s.repo.ListPublished(ctx, f)
}

func (s *ProjectService) ListAll(ctx context.Context) ([]domain.Project, error) {
	return s.repo.ListAll(ctx)
}

func (s *ProjectService) GetPublishedBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	p, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if p.Status != domain.StatusPublished {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (s *ProjectService) Get(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProjectService) Create(ctx context.Context, input port.ProjectInput) (*domain.Project, error) {
	if err := validateProjectInput(input); err != nil {
		return nil, err
	}
	p := buildProject(input)
	if p.Status == domain.StatusPublished {
		now := time.Now()
		p.PublishedAt = &now
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, input port.ProjectInput) (*domain.Project, error) {
	if err := validateProjectInput(input); err != nil {
		return nil, err
	}
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	applyProjectInput(p, input)
	// Set published_at when transitioning to published.
	if p.Status == domain.StatusPublished && p.PublishedAt == nil {
		now := time.Now()
		p.PublishedAt = &now
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func validateProjectInput(input port.ProjectInput) error {
	var fields []string
	if strings.TrimSpace(input.Slug) == "" {
		fields = append(fields, "slug", "required")
	}
	if strings.TrimSpace(input.Title) == "" {
		fields = append(fields, "title", "required")
	}
	if input.Status != domain.StatusDraft && input.Status != domain.StatusPublished {
		fields = append(fields, "status", "must be 'draft' or 'published'")
	}
	if len(fields) > 0 {
		return domain.NewValidationError(fields...)
	}
	return nil
}

func buildProject(input port.ProjectInput) *domain.Project {
	p := &domain.Project{}
	applyProjectInput(p, input)
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return p
}

func applyProjectInput(p *domain.Project, input port.ProjectInput) {
	p.Slug = input.Slug
	p.Title = input.Title
	p.Tagline = input.Tagline
	p.DescriptionMarkdown = input.DescriptionMarkdown
	p.CoverImageURL = input.CoverImageURL
	p.RepoURL = input.RepoURL
	p.DemoURL = input.DemoURL
	p.TechStack = input.TechStack
	p.Status = input.Status
	p.Featured = input.Featured
	p.SortOrder = input.SortOrder
	p.Links = make([]domain.ProjectLink, len(input.Links))
	for i, l := range input.Links {
		p.Links[i] = domain.ProjectLink{
			ID:        uuid.New(),
			Label:     l.Label,
			URL:       l.URL,
			Kind:      l.Kind,
		}
	}
	p.UpdatedAt = time.Now()
}
