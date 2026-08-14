package service_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
	"github.com/walfa-labs/backend/internal/service"
)

func TestAuthService(t *testing.T) {
	ctx := context.Background()
	adminRepo := NewMockAdminRepo()
	secret := "test-jwt-secret-key"
	accessTTL := 15 * time.Minute
	refreshTTL := 24 * time.Hour

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	adminRepo.AddUser(&domain.AdminUser{
		ID:           uuid.New(),
		Username:     "admin",
		PasswordHash: string(hash),
	})

	authSvc := service.NewAuthService(adminRepo, secret, accessTTL, refreshTTL)

	t.Run("Login with correct credentials succeeds", func(t *testing.T) {
		tokens, err := authSvc.Login(ctx, "admin", "correct-password")
		if err != nil {
			t.Fatalf("unexpected login error: %v", err)
		}
		if tokens.AccessToken == "" || tokens.RefreshToken == "" {
			t.Errorf("expected non-empty tokens, got %+v", tokens)
		}
	})

	t.Run("Login with empty username or password returns ErrUnauthorized", func(t *testing.T) {
		_, err := authSvc.Login(ctx, " ", "password")
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized for empty username, got %v", err)
		}
		_, err = authSvc.Login(ctx, "admin", "")
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized for empty password, got %v", err)
		}
	})

	t.Run("Login with invalid password returns ErrUnauthorized", func(t *testing.T) {
		_, err := authSvc.Login(ctx, "admin", "wrong-password")
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("Login with unknown username returns ErrUnauthorized", func(t *testing.T) {
		_, err := authSvc.Login(ctx, "nonexistent", "password")
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("Refresh with valid refresh token returns new access token", func(t *testing.T) {
		tokens, err := authSvc.Login(ctx, "admin", "correct-password")
		if err != nil {
			t.Fatalf("login error: %v", err)
		}

		newAccessToken, err := authSvc.Refresh(ctx, tokens.RefreshToken)
		if err != nil {
			t.Fatalf("unexpected refresh error: %v", err)
		}
		if newAccessToken == "" {
			t.Errorf("expected non-empty access token on refresh")
		}
	})

	t.Run("Refresh with access token instead of refresh token fails", func(t *testing.T) {
		tokens, err := authSvc.Login(ctx, "admin", "correct-password")
		if err != nil {
			t.Fatalf("login error: %v", err)
		}

		_, err = authSvc.Refresh(ctx, tokens.AccessToken)
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized when using access token as refresh token, got %v", err)
		}
	})

	t.Run("Refresh with malformed token returns ErrUnauthorized", func(t *testing.T) {
		_, err := authSvc.Refresh(ctx, "malformed.jwt.token")
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized for malformed token, got %v", err)
		}
	})
}

func TestExperienceService(t *testing.T) {
	ctx := context.Background()
	expRepo := NewMockExperienceRepo()
	expSvc := service.NewExperienceService(expRepo)

	now := time.Now()
	input := port.ExperienceInput{
		ExperienceType:  domain.ExperienceTypeWork,
		Organization:    "Walfa Labs",
		RoleTitle:       "Software Architect",
		Location:        "Remote",
		StartDate:       now.AddDate(-2, 0, 0),
		Current:         true,
		SummaryMarkdown: "Building distributed systems",
		SortOrder:       1,
		Highlights: []port.HighlightInput{
			{BodyMarkdown: "Designed CI/CD system", SortOrder: 1},
		},
	}

	t.Run("Create validation errors on empty fields", func(t *testing.T) {
		badInput := input
		badInput.Organization = ""
		badInput.RoleTitle = ""
		badInput.ExperienceType = ""
		_, err := expSvc.Create(ctx, badInput)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
		var valErr *domain.ValidationError
		if !errors.As(err, &valErr) {
			t.Errorf("expected *domain.ValidationError, got %T", err)
		}
	})

	t.Run("Create validation error when endDate before startDate for non-current experience", func(t *testing.T) {
		badInput := input
		badInput.Current = false
		end := now.AddDate(-3, 0, 0) // Before start date (-2 yrs)
		badInput.EndDate = &end

		_, err := expSvc.Create(ctx, badInput)
		if err == nil {
			t.Fatal("expected validation error for invalid end date, got nil")
		}
	})

	t.Run("Create and Get experience", func(t *testing.T) {
		created, err := expSvc.Create(ctx, input)
		if err != nil {
			t.Fatalf("unexpected create error: %v", err)
		}

		if created.ID == uuid.Nil {
			t.Errorf("expected valid generated UUID")
		}
		if len(created.Highlights) != 1 {
			t.Errorf("expected 1 highlight, got %d", len(created.Highlights))
		}

		fetched, err := expSvc.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected get error: %v", err)
		}
		if fetched.Organization != "Walfa Labs" {
			t.Errorf("expected organization 'Walfa Labs', got '%s'", fetched.Organization)
		}
	})

	t.Run("Get non-existent experience returns ErrNotFound", func(t *testing.T) {
		_, err := expSvc.Get(ctx, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("List experiences", func(t *testing.T) {
		list, err := expSvc.List(ctx)
		if err != nil {
			t.Fatalf("unexpected list error: %v", err)
		}
		if len(list) == 0 {
			t.Errorf("expected non-empty experience list")
		}
	})

	t.Run("Update experience", func(t *testing.T) {
		created, _ := expSvc.Create(ctx, input)
		updateInput := input
		updateInput.RoleTitle = "Principal Engineer"

		updated, err := expSvc.Update(ctx, created.ID, updateInput)
		if err != nil {
			t.Fatalf("unexpected update error: %v", err)
		}
		if updated.RoleTitle != "Principal Engineer" {
			t.Errorf("expected 'Principal Engineer', got '%s'", updated.RoleTitle)
		}
	})

	t.Run("Update non-existent experience returns ErrNotFound", func(t *testing.T) {
		_, err := expSvc.Update(ctx, uuid.New(), input)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Delete experience", func(t *testing.T) {
		created, _ := expSvc.Create(ctx, input)
		err := expSvc.Delete(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected delete error: %v", err)
		}

		_, err = expSvc.Get(ctx, created.ID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound after deletion, got %v", err)
		}
	})

	t.Run("Delete non-existent experience returns ErrNotFound", func(t *testing.T) {
		err := expSvc.Delete(ctx, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestProjectService(t *testing.T) {
	ctx := context.Background()
	projRepo := NewMockProjectRepo()
	projSvc := service.NewProjectService(projRepo)

	input := port.ProjectInput{
		Slug:                "portfolio-backend",
		Title:               "Portfolio Backend API",
		Tagline:             "High throughput Go API",
		DescriptionMarkdown: "Hexagonal Go architecture",
		CoverImageURL:       "https://example.com/cover.jpg",
		RepoURL:             "https://github.com/walfa-labs/backend",
		DemoURL:             "https://api.walfa.dev",
		TechStack:           []string{"Go", "Fiber", "Docker"},
		Status:              domain.StatusPublished,
		Featured:            true,
		SortOrder:           1,
		Links: []port.LinkInput{
			{Label: "GitHub", URL: "https://github.com/walfa-labs/backend", Kind: domain.LinkKindRepo},
		},
	}

	t.Run("Create validation errors on empty fields or invalid status", func(t *testing.T) {
		badInput := input
		badInput.Slug = ""
		badInput.Title = ""
		badInput.Status = "invalid"

		_, err := projSvc.Create(ctx, badInput)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("Create and GetPublishedBySlug project", func(t *testing.T) {
		created, err := projSvc.Create(ctx, input)
		if err != nil {
			t.Fatalf("unexpected create error: %v", err)
		}
		if created.Slug != "portfolio-backend" {
			t.Errorf("expected slug 'portfolio-backend', got '%s'", created.Slug)
		}

		fetched, err := projSvc.GetPublishedBySlug(ctx, "portfolio-backend")
		if err != nil {
			t.Fatalf("unexpected get by slug error: %v", err)
		}
		if fetched.Title != input.Title {
			t.Errorf("expected title '%s', got '%s'", input.Title, fetched.Title)
		}

		byID, err := projSvc.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected get by id error: %v", err)
		}
		if byID.Slug != "portfolio-backend" {
			t.Errorf("expected slug 'portfolio-backend', got '%s'", byID.Slug)
		}
	})

	t.Run("Get non-existent project returns ErrNotFound", func(t *testing.T) {
		_, err := projSvc.Get(ctx, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
		_, err = projSvc.GetPublishedBySlug(ctx, "nonexistent-slug")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("ListPublished with and without featured filter", func(t *testing.T) {
		featTrue := true
		featFalse := false

		featuredList, err := projSvc.ListPublished(ctx, &featTrue)
		if err != nil {
			t.Fatalf("list error: %v", err)
		}
		if len(featuredList) != 1 {
			t.Errorf("expected 1 featured project, got %d", len(featuredList))
		}

		nonFeatList, err := projSvc.ListPublished(ctx, &featFalse)
		if err != nil {
			t.Fatalf("list error: %v", err)
		}
		if len(nonFeatList) != 0 {
			t.Errorf("expected 0 non-featured projects, got %d", len(nonFeatList))
		}

		allPubList, err := projSvc.ListPublished(ctx, nil)
		if err != nil {
			t.Fatalf("list error: %v", err)
		}
		if len(allPubList) != 1 {
			t.Errorf("expected 1 published project, got %d", len(allPubList))
		}
	})

	t.Run("ListAll projects", func(t *testing.T) {
		all, err := projSvc.ListAll(ctx)
		if err != nil {
			t.Fatalf("list all error: %v", err)
		}
		if len(all) != 1 {
			t.Errorf("expected 1 project in list all, got %d", len(all))
		}
	})

	t.Run("Update project", func(t *testing.T) {
		all, _ := projSvc.ListAll(ctx)
		targetID := all[0].ID
		updateInput := input
		updateInput.Title = "Updated Title"

		updated, err := projSvc.Update(ctx, targetID, updateInput)
		if err != nil {
			t.Fatalf("update error: %v", err)
		}
		if updated.Title != "Updated Title" {
			t.Errorf("expected 'Updated Title', got '%s'", updated.Title)
		}
	})

	t.Run("Update non-existent project returns ErrNotFound", func(t *testing.T) {
		_, err := projSvc.Update(ctx, uuid.New(), input)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Delete project", func(t *testing.T) {
		all, _ := projSvc.ListAll(ctx)
		targetID := all[0].ID
		err := projSvc.Delete(ctx, targetID)
		if err != nil {
			t.Fatalf("delete error: %v", err)
		}
		_, err = projSvc.Get(ctx, targetID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound after deletion, got %v", err)
		}
	})
}

func TestPostService(t *testing.T) {
	ctx := context.Background()
	postRepo := NewMockPostRepo()
	analyticsStore := NewMockAnalyticsStore()
	postSvc := service.NewPostService(postRepo, analyticsStore)

	input := port.PostInput{
		Slug:          "devsecops-guide",
		Title:         "Complete DevSecOps Guide",
		Excerpt:       "How to secure CI/CD pipelines",
		BodyMarkdown:  "# DevSecOps\n\nSecurity automated.",
		CoverImageURL: "https://example.com/post.jpg",
		Status:        domain.StatusPublished,
		Tags: []port.TagInput{
			{Name: "Security", Slug: "security"},
			{Name: "CI/CD", Slug: "cicd"},
		},
	}

	t.Run("Create validation errors on empty fields or invalid status", func(t *testing.T) {
		badInput := input
		badInput.Slug = ""
		badInput.Title = ""
		badInput.Status = "unknown"

		_, err := postSvc.Create(ctx, badInput)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("Create post with tags", func(t *testing.T) {
		created, err := postSvc.Create(ctx, input)
		if err != nil {
			t.Fatalf("unexpected create error: %v", err)
		}
		if created.Slug != "devsecops-guide" {
			t.Errorf("expected slug 'devsecops-guide', got '%s'", created.Slug)
		}
		if len(created.Tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(created.Tags))
		}
	})

	t.Run("ListPublished with pagination and filters", func(t *testing.T) {
		posts, total, err := postSvc.ListPublished(ctx, port.PostFilter{Page: 1, PerPage: 10})
		if err != nil {
			t.Fatalf("list published error: %v", err)
		}
		if total != 1 || len(posts) != 1 {
			t.Errorf("expected 1 post, got total=%d len=%d", total, len(posts))
		}

		// Defaults when Page/PerPage are 0
		postsDefault, totalDefault, err := postSvc.ListPublished(ctx, port.PostFilter{})
		if err != nil {
			t.Fatalf("list published with default filter error: %v", err)
		}
		if totalDefault != 1 || len(postsDefault) != 1 {
			t.Errorf("expected 1 post with default filter")
		}
	})

	t.Run("ListAll posts", func(t *testing.T) {
		all, err := postSvc.ListAll(ctx)
		if err != nil {
			t.Fatalf("list all error: %v", err)
		}
		if len(all) != 1 {
			t.Errorf("expected 1 post, got %d", len(all))
		}
	})

	t.Run("Get by ID and GetPublishedBySlug", func(t *testing.T) {
		all, _ := postSvc.ListAll(ctx)
		postID := all[0].ID

		p, err := postSvc.Get(ctx, postID)
		if err != nil {
			t.Fatalf("get error: %v", err)
		}
		if p.Slug != "devsecops-guide" {
			t.Errorf("expected 'devsecops-guide', got '%s'", p.Slug)
		}

		post, err := postSvc.GetPublishedBySlug(ctx, "devsecops-guide")
		if err != nil {
			t.Fatalf("unexpected get published by slug error: %v", err)
		}
		if post.Title != "Complete DevSecOps Guide" {
			t.Errorf("unexpected post title: %s", post.Title)
		}

		// Verify analytics store recorded view
		totalViews, err := analyticsStore.TotalViews(ctx)
		if err != nil {
			t.Fatalf("analytics error: %v", err)
		}
		if totalViews != 1 {
			t.Errorf("expected 1 analytics view recorded, got %d", totalViews)
		}
	})

	t.Run("GetPublishedBySlug not found", func(t *testing.T) {
		_, err := postSvc.GetPublishedBySlug(ctx, "nonexistent-post")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Update post", func(t *testing.T) {
		all, _ := postSvc.ListAll(ctx)
		postID := all[0].ID
		updateInput := input
		updateInput.Title = "Updated DevSecOps Guide"

		updated, err := postSvc.Update(ctx, postID, updateInput)
		if err != nil {
			t.Fatalf("update error: %v", err)
		}
		if updated.Title != "Updated DevSecOps Guide" {
			t.Errorf("expected 'Updated DevSecOps Guide', got '%s'", updated.Title)
		}
	})

	t.Run("Update non-existent post returns ErrNotFound", func(t *testing.T) {
		_, err := postSvc.Update(ctx, uuid.New(), input)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("SetStatus transitions post and updates published_at", func(t *testing.T) {
		draftInput := input
		draftInput.Slug = "draft-post"
		draftInput.Status = domain.StatusDraft

		draft, err := postSvc.Create(ctx, draftInput)
		if err != nil {
			t.Fatalf("create draft error: %v", err)
		}

		published, err := postSvc.SetStatus(ctx, draft.ID, domain.StatusPublished)
		if err != nil {
			t.Fatalf("set status error: %v", err)
		}
		if published.Status != domain.StatusPublished {
			t.Errorf("expected status 'published', got '%s'", published.Status)
		}
		if published.PublishedAt == nil {
			t.Errorf("expected non-nil PublishedAt timestamp on published post")
		}

		// Invalid status
		_, err = postSvc.SetStatus(ctx, draft.ID, "invalid-status")
		if err == nil {
			t.Fatal("expected error for invalid status, got nil")
		}
	})

	t.Run("Delete post", func(t *testing.T) {
		all, _ := postSvc.ListAll(ctx)
		postID := all[0].ID
		err := postSvc.Delete(ctx, postID)
		if err != nil {
			t.Fatalf("delete error: %v", err)
		}
	})
}

func TestProfileService(t *testing.T) {
	ctx := context.Background()
	profileRepo := NewMockProfileRepo()
	profileSvc := service.NewProfileService(profileRepo)

	t.Run("Get when not configured returns ErrNotFound", func(t *testing.T) {
		_, err := profileSvc.Get(ctx)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Update profile and Get returns updated profile", func(t *testing.T) {
		input := port.ProfileInput{
			Name:        "Walfa Dev",
			Email:       "walfa@example.com",
			Tagline:     "Cloud Engineer",
			BioMarkdown: "Passionate about DevSecOps",
			Location:    "Indonesia",
			AvatarURL:   "https://example.com/me.png",
			GitHubURL:   "https://github.com/walfa-labs",
		}

		updated, err := profileSvc.Update(ctx, input)
		if err != nil {
			t.Fatalf("unexpected update error: %v", err)
		}
		if updated.Name != "Walfa Dev" {
			t.Errorf("unexpected profile data: %+v", updated)
		}

		fetched, err := profileSvc.Get(ctx)
		if err != nil {
			t.Fatalf("unexpected get error: %v", err)
		}
		if fetched.Email != "walfa@example.com" {
			t.Errorf("expected email 'walfa@example.com', got '%s'", fetched.Email)
		}
	})
}

func TestAssetService(t *testing.T) {
	ctx := context.Background()
	assetRepo := NewMockAssetRepo()
	assetStore := NewMockAssetStore()
	assetSvc := service.NewAssetService(assetRepo, assetStore)

	t.Run("Upload rejects non-image MIME types", func(t *testing.T) {
		r := bytes.NewReader([]byte("not an image content"))
		_, err := assetSvc.Upload(ctx, r, "application/pdf", int64(r.Len()))
		if err == nil {
			t.Fatal("expected error uploading PDF, got nil")
		}
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for invalid MIME type, got %v", err)
		}
	})

	t.Run("Upload rejects empty or zero size payload", func(t *testing.T) {
		r := bytes.NewReader([]byte(""))
		_, err := assetSvc.Upload(ctx, r, "image/png", 0)
		if err == nil {
			t.Fatal("expected error for file with 0 size, got nil")
		}
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for 0 size, got %v", err)
		}
	})

	t.Run("Upload various image types succeeds and GetURL presigns", func(t *testing.T) {
		imageTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp", "image/avif", "image/svg+xml"}
		for _, it := range imageTypes {
			data := []byte("fake-image-bytes")
			r := bytes.NewReader(data)

			asset, err := assetSvc.Upload(ctx, r, it, int64(len(data)))
			if err != nil {
				t.Fatalf("unexpected upload error for %s: %v", it, err)
			}
			if asset.Key == "" {
				t.Errorf("expected non-empty asset key")
			}

			presignedURL, err := assetSvc.GetURL(ctx, asset.Key)
			if err != nil {
				t.Fatalf("unexpected get url error: %v", err)
			}
			if presignedURL == "" {
				t.Errorf("expected non-empty presigned URL")
			}

			// Delete
			err = assetSvc.Delete(ctx, asset.Key)
			if err != nil {
				t.Fatalf("unexpected delete error: %v", err)
			}
		}
	})

	t.Run("GetURL on non-existent key returns error", func(t *testing.T) {
		_, err := assetSvc.GetURL(ctx, "nonexistent-key")
		if err == nil {
			t.Fatal("expected error for non-existent key, got nil")
		}
	})
}

func TestStatsService(t *testing.T) {
	ctx := context.Background()
	statsRepo := NewMockStatsRepo(port.StatsSummary{
		PublishedPosts:    10,
		PublishedProjects: 5,
		FeaturedProjects:  3,
		YearsExperience:   4,
	})
	analyticsStore := NewMockAnalyticsStore()
	_ = analyticsStore.RecordPostView(ctx, port.PostView{PostID: uuid.New(), Slug: "test-post", Title: "Test Post", ViewedAt: time.Now()})

	statsSvc := service.NewStatsService(statsRepo, analyticsStore)

	t.Run("Summary aggregates stats from ATP and ADW", func(t *testing.T) {
		summary, err := statsSvc.Summary(ctx)
		if err != nil {
			t.Fatalf("unexpected summary error: %v", err)
		}
		if summary.PublishedPosts != 10 || summary.TotalPostViews != 1 {
			t.Errorf("unexpected summary data: %+v", summary)
		}
	})

	t.Run("ViewsTimeSeries and TopPosts delegates to analyticsStore", func(t *testing.T) {
		ts, err := statsSvc.ViewsTimeSeries(ctx, time.Now().Add(-24*time.Hour), time.Now(), "day")
		if err != nil {
			t.Fatalf("unexpected timeseries error: %v", err)
		}
		if len(ts) != 1 {
			t.Errorf("expected 1 timeseries bucket, got %d", len(ts))
		}

		top, err := statsSvc.TopPosts(ctx, 5)
		if err != nil {
			t.Fatalf("unexpected top posts error: %v", err)
		}
		if len(top) != 1 {
			t.Errorf("expected 1 top post, got %d", len(top))
		}
	})
}
