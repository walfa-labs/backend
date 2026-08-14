package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
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
	"github.com/walfa-labs/backend/internal/service"
)

// --- In-memory repos for handler tests ---

type mockAdminRepo struct {
	user *domain.AdminUser
}

func (m *mockAdminRepo) GetByUsername(ctx context.Context, username string) (*domain.AdminUser, error) {
	if m.user != nil && m.user.Username == username {
		return m.user, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockAdminRepo) Upsert(ctx context.Context, u *domain.AdminUser) error {
	m.user = u
	return nil
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
	if _, ok := m.data[e.ID]; !ok {
		return domain.ErrNotFound
	}
	m.data[e.ID] = *e
	return nil
}
func (m *mockExpRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return domain.ErrNotFound
	}
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
	if _, ok := m.data[p.ID]; !ok {
		return domain.ErrNotFound
	}
	m.data[p.ID] = *p
	return nil
}
func (m *mockProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return domain.ErrNotFound
	}
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
	if _, ok := m.data[p.ID]; !ok {
		return domain.ErrNotFound
	}
	m.data[p.ID] = *p
	return nil
}
func (m *mockPostRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return domain.ErrNotFound
	}
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
	return []domain.Tag{{ID: uuid.New(), Name: "Go", Slug: "go"}}, nil
}
func (m *mockTagRepo) GetOrCreate(ctx context.Context, name, slug string) (*domain.Tag, error) {
	return &domain.Tag{ID: uuid.New(), Name: name, Slug: slug}, nil
}

type mockAnalyticsStore struct{}

func (m *mockAnalyticsStore) RecordPostView(ctx context.Context, v port.PostView) error { return nil }
func (m *mockAnalyticsStore) TotalViews(ctx context.Context) (int64, error)             { return 100, nil }
func (m *mockAnalyticsStore) ViewsTimeSeries(ctx context.Context, from, to time.Time, bucket string) ([]port.ViewsBucket, error) {
	return []port.ViewsBucket{{Bucket: time.Now(), Views: 50}}, nil
}
func (m *mockAnalyticsStore) TopPosts(ctx context.Context, limit int) ([]port.TopPost, error) {
	return []port.TopPost{{ID: uuid.New(), Slug: "test-post", Title: "Test Post", Views: 50}}, nil
}

type mockProfileRepo struct {
	profile *domain.Profile
}

func (m *mockProfileRepo) Get(ctx context.Context) (*domain.Profile, error) {
	if m.profile == nil {
		return nil, domain.ErrNotFound
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
	if _, ok := m.assets[key]; !ok {
		return domain.ErrNotFound
	}
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

func setupTestApp() (*fiber.App, *mockExpRepo, *mockProjectRepo, *mockPostRepo, *mockProfileRepo) {
	logger := platform.NewLogger("development")
	cfg := &config.Config{
		AppEnv:  "development",
		AppPort: ":8080",
	}

	app := platform.NewServer(cfg, logger)
	app.Use(middleware.RequestID())

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

	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.MinCost)
	adminRepo := &mockAdminRepo{user: &domain.AdminUser{ID: uuid.New(), Username: "admin", PasswordHash: string(hash)}}
	authSvc := service.NewAuthService(adminRepo, "secret-test", 15*time.Minute, 24*time.Hour)
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

	// Routes
	app.Get("/experiences", expH.List)
	app.Get("/experiences/:id", expH.Get)
	app.Post("/experiences", expH.Create)
	app.Put("/experiences/:id", expH.Update)
	app.Delete("/experiences/:id", expH.Delete)

	app.Get("/projects", projectH.List)
	app.Get("/projects/:slug", projectH.GetBySlug)
	app.Get("/admin/projects", projectH.AdminList)
	app.Get("/admin/projects/:id", projectH.AdminGet)
	app.Post("/projects", projectH.Create)
	app.Put("/projects/:id", projectH.Update)
	app.Delete("/projects/:id", projectH.Delete)

	app.Get("/posts", postH.List)
	app.Get("/posts/:slug", postH.GetBySlug)
	app.Get("/admin/posts", postH.AdminList)
	app.Get("/admin/posts/:id", postH.AdminGet)
	app.Post("/posts", postH.Create)
	app.Put("/posts/:id", postH.Update)
	app.Delete("/posts/:id", postH.Delete)
	app.Patch("/posts/:id/status", postH.SetStatus)

	app.Post("/auth/login", authH.Login)
	app.Post("/auth/refresh", authH.Refresh)

	app.Get("/profile", profileH.Get)
	app.Get("/admin/profile", profileH.AdminGet)
	app.Put("/profile", profileH.Update)

	app.Get("/stats/summary", statsH.Summary)
	app.Get("/stats/views", statsH.ViewsTimeSeries)
	app.Get("/stats/top-posts", statsH.TopPosts)
	app.Get("/tags", statsH.Tags)

	app.Get("/assets/*", assetH.Redirect)
	app.Post("/admin/assets", assetH.Upload)
	app.Delete("/admin/assets/*", assetH.Delete)

	return app, expRepo, projectRepo, postRepo, profileRepo
}

func TestExperienceHandler(t *testing.T) {
	app, expRepo, _, _, _ := setupTestApp()
	expID := uuid.New()
	expRepo.data[expID] = domain.Experience{
		ID:             expID,
		ExperienceType: domain.ExperienceTypeWork,
		Organization:   "Walfa Labs",
		RoleTitle:      "Lead",
		StartDate:      time.Now(),
	}

	t.Run("GET /experiences returns list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/experiences", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("GET /experiences/:id returns single item", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/experiences/"+expID.String(), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("GET /experiences/invalid-uuid returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/experiences/not-a-uuid", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 for invalid UUID, got %d", resp.StatusCode)
		}
	})

	t.Run("POST /experiences creates experience", func(t *testing.T) {
		payload := map[string]any{
			"experienceType": "work",
			"organization":   "New Org",
			"roleTitle":      "Senior Dev",
			"startDate":      "2024-01-01",
			"current":        true,
		}
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/experiences", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("expected 201 Created, got %d", resp.StatusCode)
		}
	})

	t.Run("PUT /experiences/:id updates experience", func(t *testing.T) {
		payload := map[string]any{
			"experienceType": "work",
			"organization":   "Updated Org",
			"roleTitle":      "Principal Dev",
			"startDate":      "2024-01-01",
			"current":        true,
		}
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPut, "/experiences/"+expID.String(), bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("DELETE /experiences/:id deletes experience", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/experiences/"+expID.String(), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected 204 NoContent, got %d", resp.StatusCode)
		}
	})
}

func TestProjectHandler(t *testing.T) {
	app, _, projRepo, _, _ := setupTestApp()
	projID := uuid.New()
	projRepo.data[projID] = domain.Project{
		ID:        projID,
		Slug:      "test-proj",
		Title:     "Test Project",
		Status:    domain.StatusPublished,
		TechStack: []string{"Go"},
	}

	t.Run("GET /projects", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("GET /projects/:slug", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects/test-proj", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("GET /admin/projects and /admin/projects/:id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/projects", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		getReq := httptest.NewRequest(http.MethodGet, "/admin/projects/"+projID.String(), nil)
		getResp, err := app.Test(getReq)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		getResp.Body.Close()
		if getResp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", getResp.StatusCode)
		}
	})
}

func TestPostHandler(t *testing.T) {
	app, _, _, postRepo, _ := setupTestApp()
	postID := uuid.New()
	postRepo.data[postID] = domain.BlogPost{
		ID:     postID,
		Slug:   "test-post",
		Title:  "Test Post",
		Status: domain.StatusPublished,
	}

	t.Run("GET /posts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/posts", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("GET /posts/:slug", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/posts/test-post", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("GET /admin/posts/:id includes status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/posts/"+postID.String(), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		var envelope struct {
			Data struct {
				Status string `json:"status"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if envelope.Data.Status != "published" {
			t.Errorf("expected status published, got %q", envelope.Data.Status)
		}
	})

	t.Run("PATCH /posts/:id/status", func(t *testing.T) {
		payload := map[string]string{"status": "published"}
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPatch, "/posts/"+postID.String()+"/status", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestAuthHandler(t *testing.T) {
	app, _, _, _, _ := setupTestApp()

	t.Run("POST /auth/login with correct password returns tokens", func(t *testing.T) {
		payload := map[string]string{"username": "admin", "password": "pass123"}
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		cookies := resp.Cookies()
		var foundCookie bool
		for _, c := range cookies {
			if c.Name == "refresh_token" && c.HttpOnly {
				foundCookie = true
			}
		}
		if !foundCookie {
			t.Errorf("expected httpOnly refresh_token cookie")
		}
	})

	t.Run("POST /auth/login with wrong password returns 401", func(t *testing.T) {
		payload := map[string]string{"username": "admin", "password": "wrongpassword"}
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})
}

func TestStatsAndProfileHandlers(t *testing.T) {
	app, _, _, _, _ := setupTestApp()

	t.Run("GET /stats/summary returns aggregate stats", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/stats/summary", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("GET /stats/views and /stats/top-posts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/stats/views?from=2024-01-01T00:00:00Z&to=2024-12-31T23:59:59Z&bucket=day", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		topReq := httptest.NewRequest(http.MethodGet, "/stats/top-posts?limit=5", nil)
		topResp, err := app.Test(topReq)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		topResp.Body.Close()
		if topResp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", topResp.StatusCode)
		}
	})

	t.Run("GET /tags returns tag list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tags", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("PUT and GET /profile", func(t *testing.T) {
		payload := map[string]string{
			"name":        "Walfa",
			"email":       "walfa@example.com",
			"tagline":     "DevSecOps",
			"bioMarkdown": "Bio",
		}
		b, _ := json.Marshal(payload)
		putReq := httptest.NewRequest(http.MethodPut, "/profile", bytes.NewReader(b))
		putReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(putReq)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK for PUT /profile, got %d", resp.StatusCode)
		}

		getReq := httptest.NewRequest(http.MethodGet, "/profile", nil)
		getResp, err := app.Test(getReq)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer getResp.Body.Close()

		if getResp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK for GET /profile, got %d", getResp.StatusCode)
		}
	})
}

func TestAssetHandler(t *testing.T) {
	app, _, _, _, _ := setupTestApp()

	t.Run("POST /admin/assets multipart upload", func(t *testing.T) {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		partHeader := make(map[string][]string)
		partHeader["Content-Disposition"] = []string{`form-data; name="file"; filename="test.png"`}
		partHeader["Content-Type"] = []string{"image/png"}
		fw, err := w.CreatePart(partHeader)
		if err != nil {
			t.Fatalf("create form file error: %v", err)
		}
		_, _ = fw.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
		w.Close()

		req := httptest.NewRequest(http.MethodPost, "/admin/assets", &b)
		req.Header.Set("Content-Type", w.FormDataContentType())

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("upload request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("expected 201 Created for asset upload, got %d", resp.StatusCode)
		}
	})
}
