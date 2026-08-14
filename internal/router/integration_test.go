package router_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/walfa-labs/backend/internal/adapter/handler"
	"github.com/walfa-labs/backend/internal/adapter/middleware"
	"github.com/walfa-labs/backend/internal/config"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/platform"
	"github.com/walfa-labs/backend/internal/port"
	"github.com/walfa-labs/backend/internal/router"
	"github.com/walfa-labs/backend/internal/service"
)

// --- Mocks for Router Integration Test ---

type mockAdminRepo struct {
	user *domain.AdminUser
}

func (m *mockAdminRepo) GetByUsername(ctx context.Context, username string) (*domain.AdminUser, error) {
	if m.user != nil && m.user.Username == username {
		return m.user, nil
	}
	return nil, domain.ErrNotFound
}

type mockExpRepo struct {
	mu   sync.RWMutex
	data map[uuid.UUID]domain.Experience
}

func (m *mockExpRepo) List(ctx context.Context) ([]domain.Experience, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var l []domain.Experience
	for _, e := range m.data {
		l = append(l, e)
	}
	return l, nil
}
func (m *mockExpRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Experience, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &e, nil
}
func (m *mockExpRepo) Create(ctx context.Context, e *domain.Experience) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[e.ID] = *e
	return nil
}
func (m *mockExpRepo) Update(ctx context.Context, e *domain.Experience) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[e.ID] = *e
	return nil
}
func (m *mockExpRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, id)
	return nil
}

type mockProjectRepo struct {
	mu   sync.RWMutex
	data map[uuid.UUID]domain.Project
}

func (m *mockProjectRepo) ListPublished(ctx context.Context, filter port.ProjectFilter) ([]domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var l []domain.Project
	for _, p := range m.data {
		if p.Status == domain.StatusPublished {
			l = append(l, p)
		}
	}
	return l, nil
}
func (m *mockProjectRepo) ListAll(ctx context.Context) ([]domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var l []domain.Project
	for _, p := range m.data {
		l = append(l, p)
	}
	return l, nil
}
func (m *mockProjectRepo) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.data {
		if p.Slug == slug {
			return &p, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (m *mockProjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &p, nil
}
func (m *mockProjectRepo) Create(ctx context.Context, p *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[p.ID] = *p
	return nil
}
func (m *mockProjectRepo) Update(ctx context.Context, p *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[p.ID] = *p
	return nil
}
func (m *mockProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, id)
	return nil
}

type mockPostRepo struct {
	mu   sync.RWMutex
	data map[uuid.UUID]domain.BlogPost
}

func (m *mockPostRepo) ListPublished(ctx context.Context, filter port.PostFilter) ([]port.PostSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var l []port.PostSummary
	for _, p := range m.data {
		if p.Status == domain.StatusPublished {
			l = append(l, port.PostSummary{
				ID:            p.ID,
				Slug:          p.Slug,
				Title:         p.Title,
				Excerpt:       p.Excerpt,
				CoverImageURL: p.CoverImageURL,
				PublishedAt:   p.PublishedAt,
			})
		}
	}
	return l, nil
}
func (m *mockPostRepo) CountPublished(ctx context.Context, filter port.PostFilter) (int64, error) {
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
func (m *mockPostRepo) ListAll(ctx context.Context) ([]domain.BlogPost, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var l []domain.BlogPost
	for _, p := range m.data {
		l = append(l, p)
	}
	return l, nil
}
func (m *mockPostRepo) GetPublishedBySlug(ctx context.Context, slug string) (*domain.BlogPost, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.data {
		if p.Slug == slug && p.Status == domain.StatusPublished {
			return &p, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (m *mockPostRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.BlogPost, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &p, nil
}
func (m *mockPostRepo) Create(ctx context.Context, p *domain.BlogPost) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[p.ID] = *p
	return nil
}
func (m *mockPostRepo) Update(ctx context.Context, p *domain.BlogPost) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[p.ID] = *p
	return nil
}
func (m *mockPostRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, id)
	return nil
}
func (m *mockPostRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
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
func (m *mockPostRepo) SumViews(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var s int64
	for _, p := range m.data {
		s += int64(p.ViewCount)
	}
	return s, nil
}

type mockTagRepo struct{}

func (m *mockTagRepo) List(ctx context.Context) ([]domain.Tag, error) {
	return []domain.Tag{{ID: uuid.New(), Name: "Security", Slug: "security"}}, nil
}
func (m *mockTagRepo) GetOrCreate(ctx context.Context, name, slug string) (*domain.Tag, error) {
	return &domain.Tag{ID: uuid.New(), Name: name, Slug: slug}, nil
}

type mockAnalyticsStore struct{}

func (m *mockAnalyticsStore) RecordPostView(ctx context.Context, v port.PostView) error { return nil }
func (m *mockAnalyticsStore) TotalViews(ctx context.Context) (int64, error)             { return 100, nil }
func (m *mockAnalyticsStore) ViewsTimeSeries(ctx context.Context, from, to time.Time, bucket string) ([]port.ViewsBucket, error) {
	return []port.ViewsBucket{{Bucket: time.Now(), Views: 100}}, nil
}
func (m *mockAnalyticsStore) TopPosts(ctx context.Context, limit int) ([]port.TopPost, error) {
	return []port.TopPost{{ID: uuid.New(), Slug: "top-post", Title: "Top Post", Views: 100}}, nil
}

type mockProfileRepo struct {
	profile *domain.Profile
}

func (m *mockProfileRepo) Get(ctx context.Context) (*domain.Profile, error) {
	if m.profile == nil {
		return &domain.Profile{
			Name:        "Walfa",
			Email:       "contact@walfa.dev",
			Tagline:     "Security Engineer",
			BioMarkdown: "Bio",
		}, nil
	}
	return m.profile, nil
}
func (m *mockProfileRepo) Upsert(ctx context.Context, p *domain.Profile) error {
	m.profile = p
	return nil
}

type mockAssetRepo struct {
	assets map[string]domain.Asset
}

func (m *mockAssetRepo) Create(ctx context.Context, a *domain.Asset) error {
	m.assets[a.Key] = *a
	return nil
}
func (m *mockAssetRepo) GetByKey(ctx context.Context, key string) (*domain.Asset, error) {
	a, ok := m.assets[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &a, nil
}
func (m *mockAssetRepo) DeleteByKey(ctx context.Context, key string) error {
	delete(m.assets, key)
	return nil
}

type mockAssetStore struct{}

func (m *mockAssetStore) Upload(ctx context.Context, key string, r io.Reader, contentType string, size int64) (string, error) {
	return "https://example.com/" + key, nil
}
func (m *mockAssetStore) Presign(ctx context.Context, key string) (string, error) {
	return "https://example.com/presigned/" + key, nil
}
func (m *mockAssetStore) Delete(ctx context.Context, key string) error { return nil }

type mockStatsRepo struct{}

func (m *mockStatsRepo) Summary(ctx context.Context) (port.StatsSummary, error) {
	return port.StatsSummary{PublishedPosts: 5, PublishedProjects: 3, FeaturedProjects: 2, YearsExperience: 4}, nil
}

func setupIntegrationServer(t *testing.T) (*fiber.App, string) {
	logger := platform.NewLogger("development")
	jwtSecret := "integration-test-secret"
	cfg := &config.Config{
		AppEnv:             "development",
		AppPort:            ":8080",
		JWTSecretKey:       jwtSecret,
		JWTAccessTTL:       15 * time.Minute,
		JWTRefreshTTL:      24 * time.Hour,
		CORSAllowedOrigins: []string{"http://localhost:3000"},
	}

	app := platform.NewServer(cfg, logger)

	expRepo := &mockExpRepo{data: make(map[uuid.UUID]domain.Experience)}
	expSvc := service.NewExperienceService(expRepo)
	expH := handler.NewExperienceHandler(expSvc)

	projectRepo := &mockProjectRepo{data: make(map[uuid.UUID]domain.Project)}
	projectSvc := service.NewProjectService(projectRepo)
	projectH := handler.NewProjectHandler(projectSvc)

	postRepo := &mockPostRepo{data: make(map[uuid.UUID]domain.BlogPost)}
	tagRepo := &mockTagRepo{}
	analyticsStore := &mockAnalyticsStore{}
	postSvc := service.NewPostService(postRepo, analyticsStore)
	postH := handler.NewPostHandler(postSvc)

	hash, _ := bcrypt.GenerateFromPassword([]byte("adminpass123"), bcrypt.MinCost)
	adminRepo := &mockAdminRepo{user: &domain.AdminUser{ID: uuid.New(), Username: "admin", PasswordHash: string(hash)}}
	authSvc := service.NewAuthService(adminRepo, jwtSecret, 15*time.Minute, 24*time.Hour)
	authH := handler.NewAuthHandler(authSvc, 24)

	profileRepo := &mockProfileRepo{}
	profileSvc := service.NewProfileService(profileRepo)
	profileH := handler.NewProfileHandler(profileSvc)

	assetRepo := &mockAssetRepo{assets: make(map[string]domain.Asset)}
	assetStore := &mockAssetStore{}
	assetSvc := service.NewAssetService(assetRepo, assetStore)
	assetH := handler.NewAssetHandler(assetSvc)

	statsRepo := &mockStatsRepo{}
	statsSvc := service.NewStatsService(statsRepo, analyticsStore)
	statsH := handler.NewStatsHandler(statsSvc, tagRepo)

	authMiddleware := middleware.Auth(cfg, logger)

	// Open DB handles for HealthHandler (mock DB)
	atpDB := &sql.DB{}
	adwDB := &sql.DB{}
	healthH := handler.NewHealthHandler(atpDB, adwDB)

	router.Register(app, router.Deps{
		Cfg:            cfg,
		Health:         healthH,
		Experience:     expH,
		Project:        projectH,
		Post:           postH,
		Auth:           authH,
		Asset:          assetH,
		Stats:          statsH,
		Profile:        profileH,
		Logger:         logger,
		AuthMiddleware: authMiddleware,
	})

	return app, jwtSecret
}

func TestRouterIntegration(t *testing.T) {
	app, _ := setupIntegrationServer(t)

	t.Run("GET /api/v1/experiences (Public read)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/experiences", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
		if resp.Header.Get(middleware.RequestIDHeader) == "" {
			t.Errorf("expected X-Request-Id header to be set")
		}
	})

	t.Run("GET /api/v1/projects (Public read)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("GET /api/v1/blog/posts (Public read)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/blog/posts", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("GET /api/v1/stats/summary (Public read)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/stats/summary", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("GET /api/v1/profile (Public read)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("POST /api/v1/auth/login and access Admin endpoints", func(t *testing.T) {
		loginPayload := map[string]string{
			"username": "admin",
			"password": "adminpass123",
		}
		b, _ := json.Marshal(loginPayload)
		loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(b))
		loginReq.Header.Set("Content-Type", "application/json")

		loginResp, err := app.Test(loginReq)
		if err != nil {
			t.Fatalf("login request error: %v", err)
		}
		defer loginResp.Body.Close()

		if loginResp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for login, got %d", loginResp.StatusCode)
		}

		var envelope struct {
			Data struct {
				AccessToken string `json:"accessToken"`
			} `json:"data"`
		}
		_ = json.NewDecoder(loginResp.Body).Decode(&envelope)
		accessToken := envelope.Data.AccessToken

		if accessToken == "" {
			t.Fatalf("expected non-empty accessToken")
		}

		// 1. Unauthenticated request to /api/v1/admin/experiences returns 401
		unauthReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/experiences", nil)
		unauthResp, err := app.Test(unauthReq)
		if err != nil {
			t.Fatalf("unauth request error: %v", err)
		}
		unauthResp.Body.Close()
		if unauthResp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized for admin route without token, got %d", unauthResp.StatusCode)
		}

		// 2. Authenticated request to /api/v1/admin/experiences succeeds
		expPayload := map[string]any{
			"experienceType": "work",
			"organization":   "Integration Org",
			"roleTitle":      "Lead Architect",
			"startDate":      "2023-01-01",
			"current":        true,
		}
		expBody, _ := json.Marshal(expPayload)
		authReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/experiences", bytes.NewReader(expBody))
		authReq.Header.Set("Content-Type", "application/json")
		authReq.Header.Set("Authorization", "Bearer "+accessToken)

		authResp, err := app.Test(authReq)
		if err != nil {
			t.Fatalf("auth request error: %v", err)
		}
		defer authResp.Body.Close()

		if authResp.StatusCode != http.StatusCreated {
			t.Errorf("expected 201 Created for authenticated admin experience create, got %d", authResp.StatusCode)
		}
	})
}
