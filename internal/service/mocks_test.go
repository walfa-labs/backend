package service_test

import (
	"context"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

// MockAdminRepo implements port.AdminRepo
type MockAdminRepo struct {
	mu    sync.RWMutex
	users map[string]*domain.AdminUser
}

func NewMockAdminRepo() *MockAdminRepo {
	return &MockAdminRepo{users: make(map[string]*domain.AdminUser)}
}

func (m *MockAdminRepo) GetByUsername(ctx context.Context, username string) (*domain.AdminUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[username]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (m *MockAdminRepo) AddUser(u *domain.AdminUser) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[u.Username] = u
}

func (m *MockAdminRepo) Upsert(ctx context.Context, u *domain.AdminUser) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[u.Username] = u
	return nil
}

// MockExperienceRepo implements port.ExperienceRepo
type MockExperienceRepo struct {
	mu   sync.RWMutex
	data map[uuid.UUID]domain.Experience
}

func NewMockExperienceRepo() *MockExperienceRepo {
	return &MockExperienceRepo{data: make(map[uuid.UUID]domain.Experience)}
}

func (m *MockExperienceRepo) List(ctx context.Context) ([]domain.Experience, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []domain.Experience
	for _, e := range m.data {
		list = append(list, e)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].SortOrder < list[j].SortOrder
	})
	return list, nil
}

func (m *MockExperienceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Experience, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &e, nil
}

func (m *MockExperienceRepo) Create(ctx context.Context, e *domain.Experience) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[e.ID] = *e
	return nil
}

func (m *MockExperienceRepo) Update(ctx context.Context, e *domain.Experience) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[e.ID]; !ok {
		return domain.ErrNotFound
	}
	m.data[e.ID] = *e
	return nil
}

func (m *MockExperienceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.data, id)
	return nil
}

// MockProjectRepo implements port.ProjectRepo
type MockProjectRepo struct {
	mu   sync.RWMutex
	data map[uuid.UUID]domain.Project
}

func NewMockProjectRepo() *MockProjectRepo {
	return &MockProjectRepo{data: make(map[uuid.UUID]domain.Project)}
}

func (m *MockProjectRepo) ListPublished(ctx context.Context, filter port.ProjectFilter) ([]domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []domain.Project
	for _, p := range m.data {
		if p.Status == domain.StatusPublished {
			if filter.HasFeat && p.Featured != filter.Featured {
				continue
			}
			list = append(list, p)
		}
	}
	return list, nil
}

func (m *MockProjectRepo) ListAll(ctx context.Context) ([]domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []domain.Project
	for _, p := range m.data {
		list = append(list, p)
	}
	return list, nil
}

func (m *MockProjectRepo) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.data {
		if p.Slug == slug {
			return &p, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockProjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &p, nil
}

func (m *MockProjectRepo) Create(ctx context.Context, p *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, existing := range m.data {
		if existing.Slug == p.Slug {
			return domain.ErrConflict
		}
	}
	m.data[p.ID] = *p
	return nil
}

func (m *MockProjectRepo) Update(ctx context.Context, p *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[p.ID]; !ok {
		return domain.ErrNotFound
	}
	m.data[p.ID] = *p
	return nil
}

func (m *MockProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.data, id)
	return nil
}

// MockTagRepo implements port.TagRepo
type MockTagRepo struct {
	mu   sync.RWMutex
	tags map[string]domain.Tag
}

func NewMockTagRepo() *MockTagRepo {
	return &MockTagRepo{tags: make(map[string]domain.Tag)}
}

func (m *MockTagRepo) List(ctx context.Context) ([]domain.Tag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []domain.Tag
	for _, t := range m.tags {
		list = append(list, t)
	}
	return list, nil
}

func (m *MockTagRepo) GetOrCreate(ctx context.Context, name, slug string) (*domain.Tag, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.tags[slug]; ok {
		return &t, nil
	}
	t := domain.Tag{ID: uuid.New(), Name: name, Slug: slug}
	m.tags[slug] = t
	return &t, nil
}

// MockPostRepo implements port.PostRepo
type MockPostRepo struct {
	mu   sync.RWMutex
	data map[uuid.UUID]domain.BlogPost
}

func NewMockPostRepo() *MockPostRepo {
	return &MockPostRepo{data: make(map[uuid.UUID]domain.BlogPost)}
}

func (m *MockPostRepo) ListPublished(ctx context.Context, filter port.PostFilter) ([]port.PostSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []port.PostSummary
	for _, p := range m.data {
		if p.Status == domain.StatusPublished {
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

func (m *MockPostRepo) CountPublished(ctx context.Context, filter port.PostFilter) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	for _, p := range m.data {
		if p.Status == domain.StatusPublished {
			count++
		}
	}
	return count, nil
}

func (m *MockPostRepo) ListAll(ctx context.Context) ([]domain.BlogPost, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []domain.BlogPost
	for _, p := range m.data {
		list = append(list, p)
	}
	return list, nil
}

func (m *MockPostRepo) GetPublishedBySlug(ctx context.Context, slug string) (*domain.BlogPost, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.data {
		if p.Slug == slug && p.Status == domain.StatusPublished {
			return &p, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockPostRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.BlogPost, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &p, nil
}

func (m *MockPostRepo) Create(ctx context.Context, p *domain.BlogPost) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, existing := range m.data {
		if existing.Slug == p.Slug {
			return domain.ErrConflict
		}
	}
	m.data[p.ID] = *p
	return nil
}

func (m *MockPostRepo) Update(ctx context.Context, p *domain.BlogPost) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[p.ID]; !ok {
		return domain.ErrNotFound
	}
	m.data[p.ID] = *p
	return nil
}

func (m *MockPostRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.data, id)
	return nil
}

func (m *MockPostRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.data[id]
	if !ok {
		return domain.ErrNotFound
	}
	p.ViewCount++
	m.data[id] = p
	return nil
}

func (m *MockPostRepo) SumViews(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var sum int64
	for _, p := range m.data {
		sum += int64(p.ViewCount)
	}
	return sum, nil
}

// MockAnalyticsStore implements port.AnalyticsStore
type MockAnalyticsStore struct {
	mu    sync.RWMutex
	views []port.PostView
}

func NewMockAnalyticsStore() *MockAnalyticsStore {
	return &MockAnalyticsStore{}
}

func (m *MockAnalyticsStore) RecordPostView(ctx context.Context, view port.PostView) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.views = append(m.views, view)
	return nil
}

func (m *MockAnalyticsStore) TotalViews(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.views)), nil
}

func (m *MockAnalyticsStore) ViewsTimeSeries(ctx context.Context, from, to time.Time, bucket string) ([]port.ViewsBucket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return []port.ViewsBucket{
		{Bucket: time.Now(), Views: int64(len(m.views))},
	}, nil
}

func (m *MockAnalyticsStore) TopPosts(ctx context.Context, limit int) ([]port.TopPost, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	counts := make(map[string]*port.TopPost)
	for _, v := range m.views {
		if cur, ok := counts[v.Slug]; ok {
			cur.Views++
		} else {
			counts[v.Slug] = &port.TopPost{ID: v.PostID, Slug: v.Slug, Title: v.Title, Views: 1}
		}
	}
	var res []port.TopPost
	for _, p := range counts {
		res = append(res, *p)
	}
	return res, nil
}

// MockProfileRepo implements port.ProfileRepo
type MockProfileRepo struct {
	mu      sync.RWMutex
	profile *domain.Profile
}

func NewMockProfileRepo() *MockProfileRepo {
	return &MockProfileRepo{}
}

func (m *MockProfileRepo) Get(ctx context.Context) (*domain.Profile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.profile == nil {
		return nil, domain.ErrNotFound
	}
	return m.profile, nil
}

func (m *MockProfileRepo) Upsert(ctx context.Context, p *domain.Profile) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.profile = p
	return nil
}

// MockAssetRepo implements port.AssetRepo
type MockAssetRepo struct {
	mu     sync.RWMutex
	assets map[string]domain.Asset
}

func NewMockAssetRepo() *MockAssetRepo {
	return &MockAssetRepo{assets: make(map[string]domain.Asset)}
}

func (m *MockAssetRepo) Create(ctx context.Context, a *domain.Asset) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.assets[a.Key] = *a
	return nil
}

func (m *MockAssetRepo) GetByKey(ctx context.Context, key string) (*domain.Asset, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.assets[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &a, nil
}

func (m *MockAssetRepo) DeleteByKey(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.assets[key]; !ok {
		return domain.ErrNotFound
	}
	delete(m.assets, key)
	return nil
}

// MockAssetStore implements port.AssetStore
type MockAssetStore struct {
	mu      sync.RWMutex
	objects map[string][]byte
}

func NewMockAssetStore() *MockAssetStore {
	return &MockAssetStore{objects: make(map[string][]byte)}
}

func (m *MockAssetStore) Upload(ctx context.Context, key string, r io.Reader, contentType string, size int64) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	m.objects[key] = data
	return "https://storage.example.com/" + key, nil
}

func (m *MockAssetStore) Presign(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.objects[key]; !ok {
		return "", domain.ErrNotFound
	}
	return "https://storage.example.com/presigned/" + key, nil
}

func (m *MockAssetStore) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, key)
	return nil
}

// MockStatsRepo implements port.StatsRepo
type MockStatsRepo struct {
	summary port.StatsSummary
}

func NewMockStatsRepo(s port.StatsSummary) *MockStatsRepo {
	return &MockStatsRepo{summary: s}
}

func (m *MockStatsRepo) Summary(ctx context.Context) (port.StatsSummary, error) {
	return m.summary, nil
}
