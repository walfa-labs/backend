package memory

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

// Store aggregates in-memory repositories for DAST, offline development, and testing.
type Store struct {
	Admin        *AdminRepo
	Experience   *ExperienceRepo
	Project      *ProjectRepo
	Post         *PostRepo
	Tag          *TagRepo
	Asset        *AssetRepo
	Profile      *ProfileRepo
	Stats        *StatsRepo
	Analytics    *AnalyticsStore
	AssetStorage *AssetStore
}

// NewStore initializes an in-memory store populated with demo data.
func NewStore(adminPasswordHash string) *Store {
	now := time.Now().UTC()

	adminUser := &domain.AdminUser{
		ID:           uuid.New(),
		Username:     "admin",
		PasswordHash: adminPasswordHash,
		CreatedAt:    now,
	}

	tagGo := domain.Tag{ID: uuid.New(), Name: "Go", Slug: "go"}
	tagFiber := domain.Tag{ID: uuid.New(), Name: "Fiber", Slug: "fiber"}
	tagOracle := domain.Tag{ID: uuid.New(), Name: "Oracle Cloud", Slug: "oracle-cloud"}

	demoProfile := &domain.Profile{
		Name:        "Walfa Portfolio",
		Email:       "contact@walfa.dev",
		Tagline:     "Software Engineer & Cloud Architect",
		BioMarkdown: "Passionate about high-performance backends and cloud infrastructure.",
		AvatarURL:   "/api/v1/assets/avatar.png",
		GitHubURL:   "https://github.com/walfa-labs",
		LinkedInURL: "https://linkedin.com/in/walfarid",
		TwitterURL:  "https://twitter.com/walfarid",
		Location:    "Jakarta, Indonesia",
		UpdatedAt:   now,
	}

	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	expID := uuid.New()
	demoExp := domain.Experience{
		ID:              expID,
		ExperienceType:  domain.ExperienceTypeWork,
		Organization:    "Walfa Labs",
		RoleTitle:       "Lead Backend Engineer",
		Location:        "Jakarta, Indonesia",
		StartDate:       startDate,
		Current:         true,
		SummaryMarkdown: "Architecting cloud-native solutions on Oracle Cloud ATP & ADW.",
		SortOrder:       1,
		Highlights: []domain.ExperienceHighlight{
			{ID: uuid.New(), ExperienceID: expID, BodyMarkdown: "Designed and implemented Go Fiber v3 REST APIs", SortOrder: 1},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	projID := uuid.New()
	demoProject := domain.Project{
		ID:                  projID,
		Title:               "Portfolio Backend",
		Slug:                "portfolio-backend",
		Tagline:             "Hexagonal Go REST API powering portfolio",
		DescriptionMarkdown: "High performance Go API on Oracle ATP and ADW",
		TechStack:           []string{"Go", "Fiber", "Oracle Cloud", "Docker"},
		RepoURL:             "https://github.com/walfa-labs/backend",
		DemoURL:             "https://walfa.dev",
		CoverImageURL:       "/api/v1/assets/project-cover.png",
		Status:              domain.StatusPublished,
		Featured:            true,
		SortOrder:           1,
		Links: []domain.ProjectLink{
			{ID: uuid.New(), ProjectID: projID, Label: "API Docs", URL: "http://localhost:8080/swagger/", Kind: domain.LinkKindDocs},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	demoPost := domain.BlogPost{
		ID:            uuid.New(),
		Title:         "Migrating to Oracle Cloud Free Tier",
		Slug:          "migrating-to-oracle-cloud-free-tier",
		Excerpt:       "How we migrated from PostgreSQL to Oracle ATP & ADW.",
		BodyMarkdown:  "# Migrating to Oracle Cloud\n\nPolyglot persistence on ATP & ADW.",
		CoverImageURL: "/api/v1/assets/blog-cover.png",
		Status:        domain.StatusPublished,
		PublishedAt:   &now,
		ViewCount:     42,
		Tags:          []domain.Tag{tagGo, tagOracle},
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	adminRepo := &AdminRepo{user: adminUser}
	expRepo := &ExperienceRepo{data: map[uuid.UUID]domain.Experience{demoExp.ID: demoExp}}
	projectRepo := &ProjectRepo{data: map[uuid.UUID]domain.Project{demoProject.ID: demoProject}}
	postRepo := &PostRepo{data: map[uuid.UUID]domain.BlogPost{demoPost.ID: demoPost}}
	tagRepo := &TagRepo{data: map[string]domain.Tag{tagGo.Slug: tagGo, tagFiber.Slug: tagFiber, tagOracle.Slug: tagOracle}}
	assetRepo := &AssetRepo{data: make(map[string]domain.Asset)}
	profileRepo := &ProfileRepo{profile: demoProfile}
	analyticsStore := &AnalyticsStore{
		views: []port.PostView{
			{PostID: demoPost.ID, Slug: demoPost.Slug, Title: demoPost.Title, ViewedAt: now},
		},
	}
	statsRepo := &StatsRepo{
		expRepo:     expRepo,
		projectRepo: projectRepo,
		postRepo:    postRepo,
		analytics:   analyticsStore,
	}
	assetStore := &AssetStore{files: make(map[string][]byte)}

	return &Store{
		Admin:        adminRepo,
		Experience:   expRepo,
		Project:      projectRepo,
		Post:         postRepo,
		Tag:          tagRepo,
		Asset:        assetRepo,
		Profile:      profileRepo,
		Stats:        statsRepo,
		Analytics:    analyticsStore,
		AssetStorage: assetStore,
	}
}

// AdminRepo in-memory
type AdminRepo struct {
	mu   sync.RWMutex
	user *domain.AdminUser
}

func (r *AdminRepo) GetByUsername(ctx context.Context, username string) (*domain.AdminUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.user != nil && r.user.Username == username {
		u := *r.user
		return &u, nil
	}
	return nil, domain.ErrNotFound
}

// ExperienceRepo in-memory
type ExperienceRepo struct {
	mu   sync.RWMutex
	data map[uuid.UUID]domain.Experience
}

func (r *ExperienceRepo) List(ctx context.Context) ([]domain.Experience, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Experience, 0, len(r.data))
	for _, e := range r.data {
		list = append(list, e)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].SortOrder < list[j].SortOrder
	})
	return list, nil
}

func (r *ExperienceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Experience, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &e, nil
}

func (r *ExperienceRepo) Create(ctx context.Context, e *domain.Experience) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[e.ID] = *e
	return nil
}

func (r *ExperienceRepo) Update(ctx context.Context, e *domain.Experience) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[e.ID]; !ok {
		return domain.ErrNotFound
	}
	r.data[e.ID] = *e
	return nil
}

func (r *ExperienceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.data, id)
	return nil
}

// ProjectRepo in-memory
type ProjectRepo struct {
	mu   sync.RWMutex
	data map[uuid.UUID]domain.Project
}

func (r *ProjectRepo) ListPublished(ctx context.Context, filter port.ProjectFilter) ([]domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.Project
	for _, p := range r.data {
		if p.Status == domain.StatusPublished {
			if filter.HasFeat && p.Featured != filter.Featured {
				continue
			}
			list = append(list, p)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].SortOrder < list[j].SortOrder
	})
	return list, nil
}

func (r *ProjectRepo) ListAll(ctx context.Context) ([]domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Project, 0, len(r.data))
	for _, p := range r.data {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].SortOrder < list[j].SortOrder
	})
	return list, nil
}

func (r *ProjectRepo) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.data {
		if p.Slug == slug && p.Status == domain.StatusPublished {
			cp := p
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *ProjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &p, nil
}

func (r *ProjectRepo) Create(ctx context.Context, p *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[p.ID] = *p
	return nil
}

func (r *ProjectRepo) Update(ctx context.Context, p *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[p.ID]; !ok {
		return domain.ErrNotFound
	}
	r.data[p.ID] = *p
	return nil
}

func (r *ProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.data, id)
	return nil
}

// PostRepo in-memory
type PostRepo struct {
	mu   sync.RWMutex
	data map[uuid.UUID]domain.BlogPost
}

func (r *PostRepo) ListPublished(ctx context.Context, filter port.PostFilter) ([]port.PostSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []port.PostSummary
	for _, p := range r.data {
		if p.Status == domain.StatusPublished {
			if filter.Tag != "" {
				matched := false
				for _, t := range p.Tags {
					if t.Slug == filter.Tag {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}
			list = append(list, port.PostSummary{
				ID:            p.ID,
				Slug:          p.Slug,
				Title:         p.Title,
				Excerpt:       p.Excerpt,
				CoverImageURL: p.CoverImageURL,
				PublishedAt:   p.PublishedAt,
				Tags:          p.Tags,
			})
		}
	}
	return list, nil
}

func (r *PostRepo) CountPublished(ctx context.Context, filter port.PostFilter) (int64, error) {
	list, err := r.ListPublished(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int64(len(list)), nil
}

func (r *PostRepo) ListAll(ctx context.Context) ([]domain.BlogPost, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.BlogPost, 0, len(r.data))
	for _, p := range r.data {
		list = append(list, p)
	}
	return list, nil
}

func (r *PostRepo) GetPublishedBySlug(ctx context.Context, slug string) (*domain.BlogPost, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.data {
		if p.Slug == slug && p.Status == domain.StatusPublished {
			cp := p
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *PostRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.BlogPost, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &p, nil
}

func (r *PostRepo) Create(ctx context.Context, p *domain.BlogPost) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[p.ID] = *p
	return nil
}

func (r *PostRepo) Update(ctx context.Context, p *domain.BlogPost) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[p.ID]; !ok {
		return domain.ErrNotFound
	}
	r.data[p.ID] = *p
	return nil
}

func (r *PostRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.data, id)
	return nil
}

func (r *PostRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.data[id]
	if !ok {
		return domain.ErrNotFound
	}
	p.ViewCount++
	r.data[id] = p
	return nil
}

func (r *PostRepo) SumViews(ctx context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var total int64
	for _, p := range r.data {
		total += int64(p.ViewCount)
	}
	return total, nil
}

// TagRepo in-memory
type TagRepo struct {
	mu   sync.RWMutex
	data map[string]domain.Tag
}

func (r *TagRepo) List(ctx context.Context) ([]domain.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Tag, 0, len(r.data))
	for _, t := range r.data {
		list = append(list, t)
	}
	return list, nil
}

func (r *TagRepo) GetOrCreate(ctx context.Context, name, slug string) (*domain.Tag, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.data[slug]; ok {
		return &t, nil
	}
	tag := domain.Tag{
		ID:   uuid.New(),
		Name: name,
		Slug: slug,
	}
	r.data[slug] = tag
	return &tag, nil
}

// AssetRepo in-memory
type AssetRepo struct {
	mu   sync.RWMutex
	data map[string]domain.Asset
}

func (r *AssetRepo) Create(ctx context.Context, a *domain.Asset) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[a.Key] = *a
	return nil
}

func (r *AssetRepo) GetByKey(ctx context.Context, key string) (*domain.Asset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.data[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &a, nil
}

func (r *AssetRepo) DeleteByKey(ctx context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[key]; !ok {
		return domain.ErrNotFound
	}
	delete(r.data, key)
	return nil
}

// ProfileRepo in-memory
type ProfileRepo struct {
	mu      sync.RWMutex
	profile *domain.Profile
}

func (r *ProfileRepo) Get(ctx context.Context) (*domain.Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.profile == nil {
		return nil, domain.ErrNotFound
	}
	p := *r.profile
	return &p, nil
}

func (r *ProfileRepo) Upsert(ctx context.Context, p *domain.Profile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profile = p
	return nil
}

// AnalyticsStore in-memory
type AnalyticsStore struct {
	mu    sync.RWMutex
	views []port.PostView
}

func (s *AnalyticsStore) RecordPostView(ctx context.Context, view port.PostView) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.views = append(s.views, view)
	return nil
}

func (s *AnalyticsStore) TotalViews(ctx context.Context) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return int64(len(s.views)), nil
}

func (s *AnalyticsStore) ViewsTimeSeries(ctx context.Context, from, to time.Time, bucket string) ([]port.ViewsBucket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m := make(map[string]int64)
	for _, v := range s.views {
		if (v.ViewedAt.Equal(from) || v.ViewedAt.After(from)) && (v.ViewedAt.Equal(to) || v.ViewedAt.Before(to)) {
			key := v.ViewedAt.Format("2006-01-02")
			m[key]++
		}
	}
	res := make([]port.ViewsBucket, 0, len(m))
	for k, count := range m {
		t, _ := time.Parse("2006-01-02", k)
		res = append(res, port.ViewsBucket{Bucket: t, Views: count})
	}
	return res, nil
}

func (s *AnalyticsStore) TopPosts(ctx context.Context, limit int) ([]port.TopPost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	counts := make(map[uuid.UUID]*port.TopPost)
	for _, v := range s.views {
		if entry, ok := counts[v.PostID]; ok {
			entry.Views++
		} else {
			counts[v.PostID] = &port.TopPost{
				ID:    v.PostID,
				Slug:  v.Slug,
				Title: v.Title,
				Views: 1,
			}
		}
	}
	res := make([]port.TopPost, 0, len(counts))
	for _, p := range counts {
		res = append(res, *p)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Views > res[j].Views
	})
	if limit > 0 && len(res) > limit {
		res = res[:limit]
	}
	return res, nil
}

// StatsRepo in-memory
type StatsRepo struct {
	expRepo     *ExperienceRepo
	projectRepo *ProjectRepo
	postRepo    *PostRepo
	analytics   *AnalyticsStore
}

func (r *StatsRepo) Summary(ctx context.Context) (port.StatsSummary, error) {
	expList, _ := r.expRepo.List(ctx)
	projList, _ := r.projectRepo.ListPublished(ctx, port.ProjectFilter{})
	featList, _ := r.projectRepo.ListPublished(ctx, port.ProjectFilter{HasFeat: true, Featured: true})
	postCount, _ := r.postRepo.CountPublished(ctx, port.PostFilter{})
	views, _ := r.analytics.TotalViews(ctx)

	years := 0
	if len(expList) > 0 {
		years = 3
	}

	return port.StatsSummary{
		PublishedPosts:    int(postCount),
		PublishedProjects: len(projList),
		FeaturedProjects:  len(featList),
		YearsExperience:   years,
		TotalPostViews:    views,
	}, nil
}

// AssetStore in-memory
type AssetStore struct {
	mu    sync.RWMutex
	files map[string][]byte
}

func (s *AssetStore) Upload(ctx context.Context, key string, r io.Reader, contentType string, size int64) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	s.files[key] = data
	return fmt.Sprintf("https://mock-objectstorage.oraclecloud.com/%s", key), nil
}

func (s *AssetStore) Presign(ctx context.Context, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.files[key]; !ok {
		if !strings.HasSuffix(key, ".png") && !strings.HasSuffix(key, ".jpg") {
			return "", domain.ErrNotFound
		}
	}
	return fmt.Sprintf("https://mock-objectstorage.oraclecloud.com/signed/%s", key), nil
}

func (s *AssetStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.files, key)
	return nil
}
