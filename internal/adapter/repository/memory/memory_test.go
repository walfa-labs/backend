package memory_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/adapter/repository/memory"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

func TestMemoryStore_AllRepositories(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")

	// 1. Admin
	u, err := store.Admin.GetByUsername(ctx, "admin")
	if err != nil || u.Username != "admin" {
		t.Fatalf("unexpected admin user: %v, err: %v", u, err)
	}
	_, err = store.Admin.GetByUsername(ctx, "unknown")
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// 2. Experience
	exps, err := store.Experience.List(ctx)
	if err != nil || len(exps) == 0 {
		t.Fatalf("expected experiences, got %v, err: %v", exps, err)
	}
	firstExp := exps[0]
	expGot, err := store.Experience.GetByID(ctx, firstExp.ID)
	if err != nil || expGot.ID != firstExp.ID {
		t.Fatalf("unexpected get experience: %v", err)
	}
	newExp := domain.Experience{
		ID:              uuid.New(),
		ExperienceType:  domain.ExperienceTypeWork,
		Organization:    "Acme Corp",
		RoleTitle:       "Senior Engineer",
		Location:        "Remote",
		StartDate:       time.Now(),
		Current:         true,
		SummaryMarkdown: "Leading cloud engineering",
	}
	if err := store.Experience.Create(ctx, &newExp); err != nil {
		t.Fatalf("create experience failed: %v", err)
	}
	newExp.RoleTitle = "Staff Engineer"
	if err := store.Experience.Update(ctx, &newExp); err != nil {
		t.Fatalf("update experience failed: %v", err)
	}
	if err := store.Experience.Delete(ctx, newExp.ID); err != nil {
		t.Fatalf("delete experience failed: %v", err)
	}
	if err := store.Experience.Delete(ctx, uuid.New()); err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound on non-existent delete, got %v", err)
	}

	// 3. Project
	projs, err := store.Project.ListPublished(ctx, port.ProjectFilter{})
	if err != nil || len(projs) == 0 {
		t.Fatalf("expected projects, got %v", err)
	}
	pBySlug, err := store.Project.GetBySlug(ctx, projs[0].Slug)
	if err != nil || pBySlug.Slug != projs[0].Slug {
		t.Fatalf("unexpected get by slug: %v", err)
	}
	allProjs, err := store.Project.ListAll(ctx)
	if err != nil || len(allProjs) == 0 {
		t.Fatalf("expected all projects: %v", err)
	}
	newProj := domain.Project{
		ID:        uuid.New(),
		Title:     "New Project",
		Slug:      "new-project",
		Tagline:   "Summary",
		Status:    domain.StatusDraft,
		SortOrder: 2,
	}
	if err := store.Project.Create(ctx, &newProj); err != nil {
		t.Fatalf("create project failed: %v", err)
	}
	newProj.Status = domain.StatusPublished
	if err := store.Project.Update(ctx, &newProj); err != nil {
		t.Fatalf("update project failed: %v", err)
	}
	if err := store.Project.Delete(ctx, newProj.ID); err != nil {
		t.Fatalf("delete project failed: %v", err)
	}

	// 4. Post & Tag
	tags, err := store.Tag.List(ctx)
	if err != nil || len(tags) == 0 {
		t.Fatalf("expected tags: %v", err)
	}
	newTag, err := store.Tag.GetOrCreate(ctx, "Rust", "rust")
	if err != nil || newTag.Slug != "rust" {
		t.Fatalf("expected get or create tag: %v", err)
	}

	posts, err := store.Post.ListPublished(ctx, port.PostFilter{})
	if err != nil || len(posts) == 0 {
		t.Fatalf("expected published posts: %v", err)
	}
	pCount, err := store.Post.CountPublished(ctx, port.PostFilter{})
	if err != nil || pCount == 0 {
		t.Fatalf("expected post count > 0: %v", err)
	}
	allPosts, err := store.Post.ListAll(ctx)
	if err != nil || len(allPosts) == 0 {
		t.Fatalf("expected all posts: %v", err)
	}
	postBySlug, err := store.Post.GetPublishedBySlug(ctx, posts[0].Slug)
	if err != nil || postBySlug.Slug != posts[0].Slug {
		t.Fatalf("expected get by slug post: %v", err)
	}
	if err := store.Post.IncrementViewCount(ctx, postBySlug.ID); err != nil {
		t.Fatalf("increment view count failed: %v", err)
	}
	sumViews, err := store.Post.SumViews(ctx)
	if err != nil || sumViews == 0 {
		t.Fatalf("sum views: %v, err: %v", sumViews, err)
	}
	newPost := domain.BlogPost{
		ID:        uuid.New(),
		Title:     "Test Post",
		Slug:      "test-post",
		Status:    domain.StatusDraft,
		ViewCount: 1,
	}
	if err := store.Post.Create(ctx, &newPost); err != nil {
		t.Fatalf("create post: %v", err)
	}
	newPost.Status = domain.StatusPublished
	if err := store.Post.Update(ctx, &newPost); err != nil {
		t.Fatalf("update post: %v", err)
	}
	if err := store.Post.Delete(ctx, newPost.ID); err != nil {
		t.Fatalf("delete post: %v", err)
	}

	// 5. Profile
	prof, err := store.Profile.Get(ctx)
	if err != nil || prof.Name == "" {
		t.Fatalf("expected profile: %v", err)
	}
	prof.Name = "Updated Name"
	if err := store.Profile.Upsert(ctx, prof); err != nil {
		t.Fatalf("upsert profile: %v", err)
	}

	// 6. Analytics & Stats
	if err := store.Analytics.RecordPostView(ctx, port.PostView{
		PostID:   posts[0].ID,
		Slug:     posts[0].Slug,
		Title:    posts[0].Title,
		ViewedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("record view: %v", err)
	}
	totViews, err := store.Analytics.TotalViews(ctx)
	if err != nil || totViews == 0 {
		t.Fatalf("total views: %v", err)
	}
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now().Add(24 * time.Hour)
	ts, err := store.Analytics.ViewsTimeSeries(ctx, from, to, "day")
	if err != nil || len(ts) == 0 {
		t.Fatalf("timeseries: %v", err)
	}
	top, err := store.Analytics.TopPosts(ctx, 5)
	if err != nil || len(top) == 0 {
		t.Fatalf("top posts: %v", err)
	}
	summary, err := store.Stats.Summary(ctx)
	if err != nil || summary.PublishedPosts == 0 {
		t.Fatalf("summary: %v", err)
	}

	// 7. Asset & Storage
	url, err := store.AssetStorage.Upload(ctx, "test.png", bytes.NewReader([]byte("test")), "image/png", 4)
	if err != nil || url == "" {
		t.Fatalf("upload asset: %v", err)
	}
	presign, err := store.AssetStorage.Presign(ctx, "test.png")
	if err != nil || presign == "" {
		t.Fatalf("presign asset: %v", err)
	}
	if err := store.AssetStorage.Delete(ctx, "test.png"); err != nil {
		t.Fatalf("delete asset store: %v", err)
	}
	assetRecord := domain.Asset{
		ID:          uuid.New(),
		Key:         "test.png",
		URL:         "https://example.com/test.png",
		ContentType: "image/png",
		SizeBytes:   4,
		UploadedAt:  time.Now(),
	}
	if err := store.Asset.Create(ctx, &assetRecord); err != nil {
		t.Fatalf("create asset: %v", err)
	}
	getA, err := store.Asset.GetByKey(ctx, "test.png")
	if err != nil || getA.Key != "test.png" {
		t.Fatalf("get asset: %v", err)
	}
	if err := store.Asset.DeleteByKey(ctx, "test.png"); err != nil {
		t.Fatalf("delete asset: %v", err)
	}
}
