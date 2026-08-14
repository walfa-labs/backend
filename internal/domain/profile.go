package domain

import "time"

// Profile is the singleton portfolio-owner profile aggregate.
type Profile struct {
	Name        string
	Email       string
	Tagline     string
	BioMarkdown string
	Location    string
	AvatarURL   string
	GitHubURL   string
	LinkedInURL string
	TwitterURL  string
	UpdatedAt   time.Time
}
