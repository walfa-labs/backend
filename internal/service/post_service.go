package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

type PostService struct {
	repo port.PostRepo
}

func NewPostService(repo port.PostRepo) *PostService {
	return &PostService{repo: repo}
}

// ListPublished returns published post summaries and a total count.
// The repo fetches with LIMIT/OFFSET; total comes from a SELECT COUNT(*) with the same WHERE.
func (s *PostService) ListPublished(ctx context.Context, filter port.PostFilter) ([]port.PostSummary, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 || filter.PerPage > 100 {
		filter.PerPage = 10
	}
	posts, err := s.repo.ListPublished(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountPublished(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (s *PostService) ListAll(ctx context.Context) ([]domain.BlogPost, error) {
	return s.repo.ListAll(ctx)
}

func (s *PostService) GetPublishedBySlug(ctx context.Context, slug string) (*domain.BlogPost, error) {
	p, err := s.repo.GetPublishedBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	// Best-effort view count increment (§6.4 approach #1).
	_ = s.repo.IncrementViewCount(ctx, p.ID)
	return p, nil
}

func (s *PostService) Get(ctx context.Context, id uuid.UUID) (*domain.BlogPost, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PostService) Create(ctx context.Context, input port.PostInput) (*domain.BlogPost, error) {
	if err := validatePostInput(input); err != nil {
		return nil, err
	}
	p := buildPost(input)
	if p.Status == domain.StatusPublished {
		now := time.Now()
		p.PublishedAt = &now
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PostService) Update(ctx context.Context, id uuid.UUID, input port.PostInput) (*domain.BlogPost, error) {
	if err := validatePostInput(input); err != nil {
		return nil, err
	}
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	applyPostInput(p, input)
	if p.Status == domain.StatusPublished && p.PublishedAt == nil {
		now := time.Now()
		p.PublishedAt = &now
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PostService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *PostService) SetStatus(ctx context.Context, id uuid.UUID, status domain.ContentStatus) (*domain.BlogPost, error) {
	if status != domain.StatusDraft && status != domain.StatusPublished {
		return nil, domain.NewValidationError("status", "must be 'draft' or 'published'")
	}
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Status = status
	if status == domain.StatusPublished && p.PublishedAt == nil {
		now := time.Now()
		p.PublishedAt = &now
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func validatePostInput(input port.PostInput) error {
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

func buildPost(input port.PostInput) *domain.BlogPost {
	p := &domain.BlogPost{}
	applyPostInput(p, input)
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return p
}

func applyPostInput(p *domain.BlogPost, input port.PostInput) {
	p.Slug = input.Slug
	p.Title = input.Title
	p.Excerpt = input.Excerpt
	p.BodyMarkdown = input.BodyMarkdown
	p.CoverImageURL = input.CoverImageURL
	p.Status = input.Status
	p.Tags = make([]domain.Tag, len(input.Tags))
	for i, t := range input.Tags {
		p.Tags[i] = domain.Tag{
			ID:   uuid.New(),
			Name: t.Name,
			Slug: t.Slug,
		}
	}
	p.UpdatedAt = time.Now()
}
