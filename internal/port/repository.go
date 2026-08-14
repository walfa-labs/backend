package port

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
)

// ExperienceRepo persists Experience aggregates.
type ExperienceRepo interface {
	List(ctx context.Context) ([]domain.Experience, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Experience, error)
	Create(ctx context.Context, e *domain.Experience) error
	Update(ctx context.Context, e *domain.Experience) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProjectRepo persists Project aggregates.
type ProjectRepo interface {
	ListPublished(ctx context.Context, filter ProjectFilter) ([]domain.Project, error)
	ListAll(ctx context.Context) ([]domain.Project, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	Create(ctx context.Context, p *domain.Project) error
	Update(ctx context.Context, p *domain.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProjectFilter narrows a published project listing.
type ProjectFilter struct {
	Featured bool
	HasFeat  bool // true if Featured filter is active
}

// PostRepo persists BlogPost aggregates.
type PostRepo interface {
	ListPublished(ctx context.Context, filter PostFilter) ([]PostSummary, error)
	CountPublished(ctx context.Context, filter PostFilter) (int64, error)
	ListAll(ctx context.Context) ([]domain.BlogPost, error)
	GetPublishedBySlug(ctx context.Context, slug string) (*domain.BlogPost, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BlogPost, error)
	Create(ctx context.Context, p *domain.BlogPost) error
	Update(ctx context.Context, p *domain.BlogPost) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	SumViews(ctx context.Context) (int64, error)
}

// PostFilter narrows a published post listing.
type PostFilter struct {
	Tag     string
	Page    int
	PerPage int
}

// PostSummary is a lightweight post representation for list views.
type PostSummary struct {
	ID            uuid.UUID
	Slug          string
	Title         string
	Excerpt       string
	CoverImageURL string
	PublishedAt   *time.Time
	Tags          []domain.Tag
}

// TagRepo persists Tags.
type TagRepo interface {
	List(ctx context.Context) ([]domain.Tag, error)
	GetOrCreate(ctx context.Context, name, slug string) (*domain.Tag, error)
}

// AssetRepo persists asset metadata in the database.
type AssetRepo interface {
	Create(ctx context.Context, a *domain.Asset) error
	GetByKey(ctx context.Context, key string) (*domain.Asset, error)
	DeleteByKey(ctx context.Context, key string) error
}

// AdminRepo persists the admin user.
type AdminRepo interface {
	GetByUsername(ctx context.Context, username string) (*domain.AdminUser, error)
	Upsert(ctx context.Context, u *domain.AdminUser) error
}

// ProfileRepo persists the singleton Profile aggregate.
type ProfileRepo interface {
	Get(ctx context.Context) (*domain.Profile, error)
	Upsert(ctx context.Context, p *domain.Profile) error
}

// AssetStore is the Strategy interface for object storage backends (S3, local FS).
type AssetStore interface {
	Upload(ctx context.Context, key string, r io.Reader, contentType string, size int64) (url string, err error)
	Presign(ctx context.Context, key string) (url string, err error)
	Delete(ctx context.Context, key string) error
}

// StatsRepo reads aggregate statistics from the operational (OLTP) store.
type StatsRepo interface {
	Summary(ctx context.Context) (StatsSummary, error)
}

// AnalyticsStore records and aggregates post-view analytics events in the
// analytical (ADW) store. Implementations are expected to denormalize the
// post slug/title into the warehouse so analytics queries need no cross-store
// joins.
type AnalyticsStore interface {
	RecordPostView(ctx context.Context, view PostView) error
	TotalViews(ctx context.Context) (int64, error)
	ViewsTimeSeries(ctx context.Context, from, to time.Time, bucket string) ([]ViewsBucket, error)
	TopPosts(ctx context.Context, limit int) ([]TopPost, error)
}

// PostView is a single blog-post view event recorded for analytics.
type PostView struct {
	PostID   uuid.UUID
	Slug     string
	Title    string
	ViewedAt time.Time
}

// StatsSummary is the public-facing counts payload.
type StatsSummary struct {
	PublishedPosts    int
	PublishedProjects int
	FeaturedProjects  int
	YearsExperience   int
	TotalPostViews    int64
}

// ViewsBucket is one data point in a view-count time series.
type ViewsBucket struct {
	Bucket time.Time
	Views  int64
}

// TopPost is a post ranked by view count.
type TopPost struct {
	ID    uuid.UUID
	Slug  string
	Title string
	Views int64
}
