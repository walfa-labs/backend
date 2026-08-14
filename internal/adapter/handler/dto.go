package handler

import (
	"time"

	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

// --- Experience DTOs ---

type ExperienceResponse struct {
	ID              string                   `json:"id"`
	ExperienceType  string                   `json:"experienceType"`
	Organization    string                   `json:"organization"`
	RoleTitle       string                   `json:"roleTitle"`
	Location        string                   `json:"location"`
	StartDate       string                   `json:"startDate"`
	EndDate         *string                  `json:"endDate"`
	Current         bool                     `json:"current"`
	SummaryMarkdown string                   `json:"summaryMarkdown"`
	SortOrder       int                      `json:"sortOrder"`
	Highlights      []HighlightResponse     `json:"highlights"`
}

type HighlightResponse struct {
	ID           string `json:"id"`
	BodyMarkdown string `json:"bodyMarkdown"`
	SortOrder    int    `json:"sortOrder"`
}

func toExperienceResponse(e *domain.Experience) ExperienceResponse {
	resp := ExperienceResponse{
		ID:              e.ID.String(),
		ExperienceType:  string(e.ExperienceType),
		Organization:    e.Organization,
		RoleTitle:       e.RoleTitle,
		Location:        e.Location,
		StartDate:       e.StartDate.Format("2006-01-02"),
		Current:         e.Current,
		SummaryMarkdown: e.SummaryMarkdown,
		SortOrder:       e.SortOrder,
		Highlights:      make([]HighlightResponse, 0, len(e.Highlights)),
	}
	if e.EndDate != nil {
		s := e.EndDate.Format("2006-01-02")
		resp.EndDate = &s
	}
	for _, h := range e.Highlights {
		resp.Highlights = append(resp.Highlights, HighlightResponse{
			ID:           h.ID.String(),
			BodyMarkdown: h.BodyMarkdown,
			SortOrder:    h.SortOrder,
		})
	}
	return resp
}

type experienceRequest struct {
	ExperienceType  string              `json:"experienceType" validate:"required,oneof=work education"`
	Organization    string              `json:"organization" validate:"required"`
	RoleTitle       string              `json:"roleTitle" validate:"required"`
	Location        string              `json:"location"`
	StartDate       string              `json:"startDate" validate:"required"`
	EndDate         *string             `json:"endDate"`
	Current         bool                `json:"current"`
	SummaryMarkdown string              `json:"summaryMarkdown"`
	SortOrder       int                 `json:"sortOrder"`
	Highlights      []highlightRequest  `json:"highlights"`
}

type highlightRequest struct {
	BodyMarkdown string `json:"bodyMarkdown"`
	SortOrder    int    `json:"sortOrder"`
}

func (r experienceRequest) toInput() (port.ExperienceInput, error) {
	st, err := time.Parse("2006-01-02", r.StartDate)
	if err != nil {
		return port.ExperienceInput{}, domain.NewValidationError("startDate", "must be YYYY-MM-DD")
	}
	var end *time.Time
	if r.EndDate != nil && *r.EndDate != "" {
		t, err := time.Parse("2006-01-02", *r.EndDate)
		if err != nil {
			return port.ExperienceInput{}, domain.NewValidationError("endDate", "must be YYYY-MM-DD")
		}
		end = &t
	}
	highlights := make([]port.HighlightInput, len(r.Highlights))
	for i, h := range r.Highlights {
		highlights[i] = port.HighlightInput{
			BodyMarkdown: h.BodyMarkdown,
			SortOrder:    h.SortOrder,
		}
	}
	return port.ExperienceInput{
		ExperienceType:  domain.ExperienceType(r.ExperienceType),
		Organization:    r.Organization,
		RoleTitle:       r.RoleTitle,
		Location:        r.Location,
		StartDate:       st,
		EndDate:         end,
		Current:         r.Current,
		SummaryMarkdown: r.SummaryMarkdown,
		SortOrder:       r.SortOrder,
		Highlights:      highlights,
	}, nil
}

// --- Project DTOs ---

type ProjectResponse struct {
	ID                  string           `json:"id"`
	Slug                string           `json:"slug"`
	Title               string           `json:"title"`
	Tagline             string           `json:"tagline"`
	DescriptionMarkdown string           `json:"descriptionMarkdown"`
	CoverImageURL       string           `json:"coverImageUrl"`
	RepoURL             string           `json:"repoUrl"`
	DemoURL             string           `json:"demoUrl"`
	TechStack           []string         `json:"techStack"`
	Status              string           `json:"status"`
	Featured            bool             `json:"featured"`
	SortOrder           int              `json:"sortOrder"`
	PublishedAt         *string          `json:"publishedAt"`
	Links               []LinkResponse   `json:"links"`
}

type LinkResponse struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url"`
	Kind  string `json:"kind"`
}

func toProjectResponse(p *domain.Project) ProjectResponse {
	resp := ProjectResponse{
		ID:                  p.ID.String(),
		Slug:                p.Slug,
		Title:               p.Title,
		Tagline:             p.Tagline,
		DescriptionMarkdown: p.DescriptionMarkdown,
		CoverImageURL:       p.CoverImageURL,
		RepoURL:             p.RepoURL,
		DemoURL:             p.DemoURL,
		TechStack:           p.TechStack,
		Status:              string(p.Status),
		Featured:            p.Featured,
		SortOrder:           p.SortOrder,
		Links:               make([]LinkResponse, 0, len(p.Links)),
	}
	if p.PublishedAt != nil {
		s := p.PublishedAt.Format(time.RFC3339)
		resp.PublishedAt = &s
	}
	if resp.TechStack == nil {
		resp.TechStack = []string{}
	}
	for _, l := range p.Links {
		resp.Links = append(resp.Links, LinkResponse{
			ID:    l.ID.String(),
			Label: l.Label,
			URL:   l.URL,
			Kind:  string(l.Kind),
		})
	}
	return resp
}

type projectRequest struct {
	Slug                string         `json:"slug" validate:"required"`
	Title               string         `json:"title" validate:"required"`
	Tagline             string         `json:"tagline"`
	DescriptionMarkdown string         `json:"descriptionMarkdown"`
	CoverImageURL       string         `json:"coverImageUrl"`
	RepoURL             string         `json:"repoUrl"`
	DemoURL             string         `json:"demoUrl"`
	TechStack           []string       `json:"techStack"`
	Status              string         `json:"status" validate:"oneof=draft published"`
	Featured            bool           `json:"featured"`
	SortOrder           int            `json:"sortOrder"`
	Links              []linkRequest   `json:"links"`
}

type linkRequest struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Kind  string `json:"kind"`
}

func (r projectRequest) toInput() (port.ProjectInput, error) {
	links := make([]port.LinkInput, len(r.Links))
	for i, l := range r.Links {
		links[i] = port.LinkInput{
			Label: l.Label,
			URL:   l.URL,
			Kind:  domain.LinkKind(l.Kind),
		}
	}
	status := domain.StatusDraft
	if r.Status == "published" {
		status = domain.StatusPublished
	}
	ts := r.TechStack
	if ts == nil {
		ts = []string{}
	}
	return port.ProjectInput{
		Slug:                r.Slug,
		Title:               r.Title,
		Tagline:             r.Tagline,
		DescriptionMarkdown: r.DescriptionMarkdown,
		CoverImageURL:       r.CoverImageURL,
		RepoURL:             r.RepoURL,
		DemoURL:             r.DemoURL,
		TechStack:           ts,
		Status:              status,
		Featured:            r.Featured,
		SortOrder:           r.SortOrder,
		Links:               links,
	}, nil
}

// --- Post DTOs ---

type PostSummaryResponse struct {
	ID            string      `json:"id"`
	Slug          string      `json:"slug"`
	Title         string      `json:"title"`
	Excerpt       string      `json:"excerpt"`
	CoverImageURL string      `json:"coverImageUrl"`
	PublishedAt   *string     `json:"publishedAt"`
	Tags          []TagResponse `json:"tags"`
}

type PostResponse struct {
	PostSummaryResponse
	BodyMarkdown string `json:"bodyMarkdown"`
	ViewCount    int    `json:"viewCount"`
	Status       string `json:"status"`
}

type TagResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func toPostSummaryResponse(s port.PostSummary) PostSummaryResponse {
	resp := PostSummaryResponse{
		ID:            s.ID.String(),
		Slug:          s.Slug,
		Title:         s.Title,
		Excerpt:       s.Excerpt,
		CoverImageURL: s.CoverImageURL,
		Tags:          make([]TagResponse, 0, len(s.Tags)),
	}
	if s.PublishedAt != nil {
		t := s.PublishedAt.Format(time.RFC3339)
		resp.PublishedAt = &t
	}
	for _, t := range s.Tags {
		resp.Tags = append(resp.Tags, TagResponse{ID: t.ID.String(), Name: t.Name, Slug: t.Slug})
	}
	return resp
}

func toPostResponse(p *domain.BlogPost) PostResponse {
	resp := PostResponse{
		PostSummaryResponse: PostSummaryResponse{
			ID:            p.ID.String(),
			Slug:          p.Slug,
			Title:         p.Title,
			Excerpt:       p.Excerpt,
			CoverImageURL: p.CoverImageURL,
			Tags:          make([]TagResponse, 0, len(p.Tags)),
		},
		BodyMarkdown: p.BodyMarkdown,
		ViewCount:    p.ViewCount,
		Status:       string(p.Status),
	}
	if p.PublishedAt != nil {
		t := p.PublishedAt.Format(time.RFC3339)
		resp.PublishedAt = &t
	}
	for _, t := range p.Tags {
		resp.Tags = append(resp.Tags, TagResponse{ID: t.ID.String(), Name: t.Name, Slug: t.Slug})
	}
	return resp
}

type postRequest struct {
	Slug          string      `json:"slug" validate:"required"`
	Title         string      `json:"title" validate:"required"`
	Excerpt       string      `json:"excerpt"`
	BodyMarkdown  string      `json:"bodyMarkdown"`
	CoverImageURL string      `json:"coverImageUrl"`
	Status        string      `json:"status" validate:"oneof=draft published"`
	Tags          []tagInput  `json:"tags"`
}

type tagInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (r postRequest) toInput() port.PostInput {
	tags := make([]port.TagInput, len(r.Tags))
	for i, t := range r.Tags {
		tags[i] = port.TagInput{Name: t.Name, Slug: t.Slug}
	}
	status := domain.StatusDraft
	if r.Status == "published" {
		status = domain.StatusPublished
	}
	return port.PostInput{
		Slug:          r.Slug,
		Title:         r.Title,
		Excerpt:       r.Excerpt,
		BodyMarkdown:  r.BodyMarkdown,
		CoverImageURL: r.CoverImageURL,
		Status:        status,
		Tags:          tags,
	}
}

// --- Stats DTOs ---

type StatsSummaryResponse struct {
	PublishedPosts    int   `json:"publishedPosts"`
	PublishedProjects int   `json:"publishedProjects"`
	FeaturedProjects  int   `json:"featuredProjects"`
	YearsExperience   int   `json:"yearsExperience"`
	TotalPostViews    int64 `json:"totalPostViews"`
}

func toStatsSummaryResponse(s port.StatsSummary) StatsSummaryResponse {
	return StatsSummaryResponse{
		PublishedPosts:    s.PublishedPosts,
		PublishedProjects: s.PublishedProjects,
		FeaturedProjects:  s.FeaturedProjects,
		YearsExperience:   s.YearsExperience,
		TotalPostViews:    s.TotalPostViews,
	}
}

// --- Auth DTOs ---

type loginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type loginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type refreshResponse struct {
	AccessToken string `json:"accessToken"`
}

// --- Asset DTOs ---

type assetResponse struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	URL         string `json:"url"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}

func toAssetResponse(a *domain.Asset) assetResponse {
	return assetResponse{
		ID:          a.ID.String(),
		Key:         a.Key,
		URL:         a.URL,
		ContentType: a.ContentType,
		SizeBytes:   a.SizeBytes,
	}
}

// --- Profile DTOs ---

type ProfileResponse struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Tagline     string `json:"tagline"`
	BioMarkdown string `json:"bioMarkdown"`
	Location    string `json:"location"`
	AvatarURL   string `json:"avatarUrl"`
	GitHubURL   string `json:"githubUrl"`
	LinkedInURL string `json:"linkedinUrl"`
	TwitterURL  string `json:"twitterUrl"`
	UpdatedAt   string `json:"updatedAt"`
}

type profileRequest struct {
	Name        string `json:"name" validate:"required"`
	Email       string `json:"email"`
	Tagline     string `json:"tagline"`
	BioMarkdown string `json:"bioMarkdown"`
	Location    string `json:"location"`
	AvatarURL   string `json:"avatarUrl"`
	GitHubURL   string `json:"githubUrl"`
	LinkedInURL string `json:"linkedinUrl"`
	TwitterURL  string `json:"twitterUrl"`
}

func toProfileResponse(p *domain.Profile) ProfileResponse {
	return ProfileResponse{
		Name:        p.Name,
		Email:       p.Email,
		Tagline:     p.Tagline,
		BioMarkdown: p.BioMarkdown,
		Location:    p.Location,
		AvatarURL:   p.AvatarURL,
		GitHubURL:   p.GitHubURL,
		LinkedInURL: p.LinkedInURL,
		TwitterURL:  p.TwitterURL,
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

func (r profileRequest) toInput() port.ProfileInput {
	return port.ProfileInput{
		Name:        r.Name,
		Email:       r.Email,
		Tagline:     r.Tagline,
		BioMarkdown: r.BioMarkdown,
		Location:    r.Location,
		AvatarURL:   r.AvatarURL,
		GitHubURL:   r.GitHubURL,
		LinkedInURL: r.LinkedInURL,
		TwitterURL:  r.TwitterURL,
	}
}
