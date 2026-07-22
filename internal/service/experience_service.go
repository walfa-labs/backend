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

type ExperienceService struct {
	repo port.ExperienceRepo
}

func NewExperienceService(repo port.ExperienceRepo) *ExperienceService {
	return &ExperienceService{repo: repo}
}

func (s *ExperienceService) List(ctx context.Context) ([]domain.Experience, error) {
	return s.repo.List(ctx)
}

func (s *ExperienceService) Get(ctx context.Context, id uuid.UUID) (*domain.Experience, error) {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return e, nil
}

func (s *ExperienceService) Create(ctx context.Context, input port.ExperienceInput) (*domain.Experience, error) {
	if err := validateExperienceInput(input); err != nil {
		return nil, err
	}
	e := buildExperience(input)
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *ExperienceService) Update(ctx context.Context, id uuid.UUID, input port.ExperienceInput) (*domain.Experience, error) {
	if err := validateExperienceInput(input); err != nil {
		return nil, err
	}
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	applyExperienceInput(e, input)
	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *ExperienceService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func validateExperienceInput(input port.ExperienceInput) error {
	var fields []string
	if input.ExperienceType != domain.ExperienceTypeWork && input.ExperienceType != domain.ExperienceTypeEducation {
		fields = append(fields, "experienceType", "must be 'work' or 'education'")
	}
	if strings.TrimSpace(input.Organization) == "" {
		fields = append(fields, "organization", "required")
	}
	if strings.TrimSpace(input.RoleTitle) == "" {
		fields = append(fields, "roleTitle", "required")
	}
	if input.StartDate.IsZero() {
		fields = append(fields, "startDate", "required")
	}
	if input.EndDate != nil && input.EndDate.Before(input.StartDate) {
		fields = append(fields, "endDate", "must be after start date")
	}
	if len(fields) > 0 {
		return domain.NewValidationError(fields...)
	}
	return nil
}

func buildExperience(input port.ExperienceInput) *domain.Experience {
	e := &domain.Experience{}
	applyExperienceInput(e, input)
	e.ID = uuid.New()
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
	return e
}

func applyExperienceInput(e *domain.Experience, input port.ExperienceInput) {
	e.ExperienceType = input.ExperienceType
	e.Organization = input.Organization
	e.RoleTitle = input.RoleTitle
	e.Location = input.Location
	e.StartDate = input.StartDate
	e.EndDate = input.EndDate
	e.Current = input.Current
	e.SummaryMarkdown = input.SummaryMarkdown
	e.SortOrder = input.SortOrder
	e.Highlights = make([]domain.ExperienceHighlight, len(input.Highlights))
	for i, h := range input.Highlights {
		e.Highlights[i] = domain.ExperienceHighlight{
			ID:           uuid.New(),
			BodyMarkdown: h.BodyMarkdown,
			SortOrder:    h.SortOrder,
		}
	}
	e.UpdatedAt = time.Now()
}
