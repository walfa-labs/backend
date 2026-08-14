package domain

import (
	"time"

	"github.com/google/uuid"
)

// ContentStatus governs visibility: only published items appear on public endpoints.
type ContentStatus string

// Content status values for draft and published items.
const (
	StatusDraft     ContentStatus = "draft"
	StatusPublished ContentStatus = "published"
)

// LinkKind categorizes a project's external link.
type LinkKind string

// Recognized project link kinds.
const (
	LinkKindRepo  LinkKind = "repo"
	LinkKindDemo  LinkKind = "demo"
	LinkKindDocs  LinkKind = "docs"
	LinkKindOther LinkKind = "other"
)

// Project is a portfolio project entry.
type Project struct {
	ID                  uuid.UUID
	Slug                string
	Title               string
	Tagline             string
	DescriptionMarkdown string
	CoverImageURL       string
	RepoURL             string
	DemoURL             string
	TechStack           []string
	Status              ContentStatus
	Featured            bool
	SortOrder           int
	PublishedAt         *time.Time
	Links               []ProjectLink
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// ProjectLink is an external link associated with a project.
type ProjectLink struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Label     string
	URL       string
	Kind      LinkKind
}
