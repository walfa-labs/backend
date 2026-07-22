package domain

import (
	"time"

	"github.com/google/uuid"
)

// ExperienceType distinguishes employment from study.
type ExperienceType string

const (
	ExperienceTypeWork      ExperienceType = "work"
	ExperienceTypeEducation ExperienceType = "education"
)

// Experience is a work or education entry on the portfolio timeline.
type Experience struct {
	ID              uuid.UUID
	ExperienceType  ExperienceType
	Organization    string
	RoleTitle       string
	Location        string
	StartDate       time.Time
	EndDate         *time.Time // nil = present
	Current         bool
	SummaryMarkdown string
	SortOrder       int
	Highlights      []ExperienceHighlight
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ExperienceHighlight is a bullet point under an experience entry.
type ExperienceHighlight struct {
	ID           uuid.UUID
	ExperienceID uuid.UUID
	BodyMarkdown string
	SortOrder    int
}
