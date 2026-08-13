package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
)

func TestDomainErrors(t *testing.T) {
	t.Run("sentinel errors identity", func(t *testing.T) {
		sentinels := []error{
			domain.ErrNotFound,
			domain.ErrConflict,
			domain.ErrValidation,
			domain.ErrUnauthorized,
			domain.ErrForbidden,
		}

		for _, err := range sentinels {
			if err == nil {
				t.Fatalf("expected non-nil sentinel error")
			}
			if err.Error() == "" {
				t.Fatalf("expected non-empty error string for %v", err)
			}
		}
	})

	t.Run("NewValidationError builds correctly with valid pairs", func(t *testing.T) {
		valErr := domain.NewValidationError("username", "cannot be empty", "password", "too short")
		if valErr == nil {
			t.Fatal("expected non-nil ValidationError")
		}

		if len(valErr.Fields) != 2 {
			t.Fatalf("expected 2 field errors, got %d", len(valErr.Fields))
		}

		if valErr.Fields[0].Field != "username" || valErr.Fields[0].Issue != "cannot be empty" {
			t.Errorf("unexpected field 0: %+v", valErr.Fields[0])
		}

		if valErr.Fields[1].Field != "password" || valErr.Fields[1].Issue != "too short" {
			t.Errorf("unexpected field 1: %+v", valErr.Fields[1])
		}

		if !errors.Is(valErr, domain.ErrValidation) {
			t.Errorf("expected valErr to unwrap to ErrValidation")
		}

		if valErr.Error() != "validation failed: 2 field error(s)" {
			t.Errorf("unexpected error string: %s", valErr.Error())
		}
	})

	t.Run("NewValidationError panics on odd number of arguments", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic on odd number of arguments")
			}
		}()

		domain.NewValidationError("username", "cannot be empty", "dangling")
	})
}

func TestDomainEntities(t *testing.T) {
	t.Run("Experience entity creation", func(t *testing.T) {
		id := uuid.New()
		now := time.Now()
		exp := domain.Experience{
			ID:              id,
			ExperienceType:  domain.ExperienceTypeWork,
			Organization:    "Walfa Labs",
			RoleTitle:       "Lead Engineer",
			Location:        "Remote",
			StartDate:       now,
			Current:         true,
			SummaryMarkdown: "Leading cloud architecture",
			SortOrder:       1,
			Highlights: []domain.ExperienceHighlight{
				{
					ID:           uuid.New(),
					ExperienceID: id,
					BodyMarkdown: "Built CI/CD pipeline",
					SortOrder:    1,
				},
			},
		}

		if exp.ID != id || exp.ExperienceType != domain.ExperienceTypeWork || !exp.Current {
			t.Errorf("unexpected experience entity values: %+v", exp)
		}
	})

	t.Run("BlogPost entity creation and ETag", func(t *testing.T) {
		id := uuid.New()
		now := time.Now()
		post := domain.BlogPost{
			ID:            id,
			Slug:          "devsecops-pipeline",
			Title:         "Building a DevSecOps Pipeline",
			Excerpt:       "An overview of CI/CD security scanning",
			BodyMarkdown:  "# DevSecOps\n\nFull pipeline guide.",
			CoverImageURL: "https://example.com/cover.jpg",
			Status:        domain.StatusPublished,
			ViewCount:     42,
			PublishedAt:   &now,
			CreatedAt:     now,
			UpdatedAt:     now,
			Tags: []domain.Tag{
				{ID: uuid.New(), Name: "DevSecOps", Slug: "devsecops"},
			},
		}

		if post.ID != id || post.Slug != "devsecops-pipeline" || post.Status != domain.StatusPublished {
			t.Errorf("unexpected post entity values: %+v", post)
		}

		etag := post.ETag()
		if len(etag) != 16 {
			t.Errorf("expected 16-character hex ETag, got '%s'", etag)
		}
	})

	t.Run("Asset validation and prefix", func(t *testing.T) {
		if err := domain.ValidateAssetContentType("image/png"); err != nil {
			t.Errorf("expected image/png to be valid, got %v", err)
		}
		if err := domain.ValidateAssetContentType("application/pdf"); err == nil {
			t.Errorf("expected application/pdf to be invalid")
		}

		if domain.AssetKeyPrefix("image/png") != "images" {
			t.Errorf("expected 'images', got '%s'", domain.AssetKeyPrefix("image/png"))
		}
		if domain.AssetKeyPrefix("application/pdf") != "files" {
			t.Errorf("expected 'files', got '%s'", domain.AssetKeyPrefix("application/pdf"))
		}
	})

	t.Run("Project entity creation", func(t *testing.T) {
		id := uuid.New()
		proj := domain.Project{
			ID:                  id,
			Slug:                "portfolio-api",
			Title:               "Portfolio API",
			Tagline:             "High performance Go backend",
			DescriptionMarkdown: "Hexagonal architecture with Oracle Cloud",
			CoverImageURL:       "https://example.com/proj.jpg",
			RepoURL:             "https://github.com/walfa-labs/backend",
			DemoURL:             "https://walfa.dev",
			TechStack:           []string{"Go", "Fiber", "Oracle ATP", "ADW"},
			Status:              domain.StatusPublished,
			Featured:            true,
			SortOrder:           1,
			Links: []domain.ProjectLink{
				{ID: uuid.New(), ProjectID: id, Label: "Docs", URL: "https://walfa.dev/docs", Kind: domain.LinkKindDocs},
			},
		}

		if proj.ID != id || proj.Slug != "portfolio-api" || !proj.Featured {
			t.Errorf("unexpected project entity values: %+v", proj)
		}
	})

	t.Run("Profile entity creation", func(t *testing.T) {
		profile := domain.Profile{
			Name:        "Walfa Developer",
			Email:       "contact@walfa.dev",
			Tagline:     "Cloud & Security Engineer",
			BioMarkdown: "Passionate about robust architectures.",
			Location:    "Indonesia",
			AvatarURL:   "https://example.com/avatar.jpg",
			GitHubURL:   "https://github.com/walfa-labs",
			LinkedInURL: "https://linkedin.com/in/walfa",
			TwitterURL:  "https://x.com/walfa",
		}

		if profile.Name != "Walfa Developer" || profile.Email != "contact@walfa.dev" {
			t.Errorf("unexpected profile entity values: %+v", profile)
		}
	})
}
