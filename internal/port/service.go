package port

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
)

// ExperienceService is the use-case port for experience management.
type ExperienceService interface {
	List(ctx context.Context) ([]domain.Experience, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Experience, error)
	Create(ctx context.Context, input ExperienceInput) (*domain.Experience, error)
	Update(ctx context.Context, id uuid.UUID, input ExperienceInput) (*domain.Experience, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ExperienceInput is the write payload for create/update.
type ExperienceInput struct {
	ExperienceType  domain.ExperienceType
	Organization    string
	RoleTitle       string
	Location        string
	StartDate       time.Time
	EndDate         *time.Time
	Current         bool
	SummaryMarkdown string
	SortOrder       int
	Highlights      []HighlightInput
}

// HighlightInput is the write payload for an experience highlight.
type HighlightInput struct {
	BodyMarkdown string
	SortOrder    int
}

// ProjectService is the use-case port for project management.
type ProjectService interface {
	ListPublished(ctx context.Context, featured *bool) ([]domain.Project, error)
	ListAll(ctx context.Context) ([]domain.Project, error)
	GetPublishedBySlug(ctx context.Context, slug string) (*domain.Project, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	Create(ctx context.Context, input ProjectInput) (*domain.Project, error)
	Update(ctx context.Context, id uuid.UUID, input ProjectInput) (*domain.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProjectInput is the write payload for create/update.
type ProjectInput struct {
	Slug                string
	Title               string
	Tagline             string
	DescriptionMarkdown string
	CoverImageURL       string
	RepoURL             string
	DemoURL             string
	TechStack           []string
	Status              domain.ContentStatus
	Featured            bool
	SortOrder           int
	Links               []LinkInput
}

// LinkInput is the write payload for a project link.
type LinkInput struct {
	Label string
	URL   string
	Kind  domain.LinkKind
}

// PostService is the use-case port for blog post management.
type PostService interface {
	ListPublished(ctx context.Context, filter PostFilter) ([]PostSummary, int64, error)
	ListAll(ctx context.Context) ([]domain.BlogPost, error)
	GetPublishedBySlug(ctx context.Context, slug string) (*domain.BlogPost, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.BlogPost, error)
	Create(ctx context.Context, input PostInput) (*domain.BlogPost, error)
	Update(ctx context.Context, id uuid.UUID, input PostInput) (*domain.BlogPost, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SetStatus(ctx context.Context, id uuid.UUID, status domain.ContentStatus) (*domain.BlogPost, error)
}

// PostInput is the write payload for create/update.
type PostInput struct {
	Slug          string
	Title         string
	Excerpt       string
	BodyMarkdown  string
	CoverImageURL string
	Status        domain.ContentStatus
	Tags          []TagInput
}

// TagInput is the write payload for a tag association.
type TagInput struct {
	Name string
	Slug string
}

// AuthTokens is the token pair returned by Login.
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}

// AuthService handles authentication and token issuance (§4.6).
type AuthService interface {
	Login(ctx context.Context, username, password string) (*AuthTokens, error)
	Refresh(ctx context.Context, refreshToken string) (string, error)
}

// AssetService handles asset uploads and retrieval.
type AssetService interface {
	Upload(ctx context.Context, r io.ReadSeeker, contentType string, size int64) (*domain.Asset, error)
	GetURL(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

// StatsService is the use-case port for analytics.
type StatsService interface {
	Summary(ctx context.Context) (StatsSummary, error)
	ViewsTimeSeries(ctx context.Context, from, to time.Time, bucket string) ([]ViewsBucket, error)
	TopPosts(ctx context.Context, limit int) ([]TopPost, error)
}

// ProfileService is the use-case port for the singleton profile.
type ProfileService interface {
	Get(ctx context.Context) (*domain.Profile, error)
	Update(ctx context.Context, input ProfileInput) (*domain.Profile, error)
}

// ProfileInput is the write payload for the singleton profile.
type ProfileInput struct {
	Name        string
	Email       string
	Tagline     string
	BioMarkdown string
	Location    string
	AvatarURL   string
	GitHubURL   string
	LinkedInURL string
	TwitterURL  string
}
