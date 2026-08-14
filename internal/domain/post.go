package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BlogPost is a blog article authored via the CMS.
type BlogPost struct {
	ID            uuid.UUID
	Slug          string
	Title         string
	Excerpt       string
	BodyMarkdown  string
	CoverImageURL string
	Status        ContentStatus
	ViewCount     int
	PublishedAt   *time.Time
	Tags          []Tag
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ETag returns a weak ETag derived from UpdatedAt, suitable for If-None-Match.
func (p *BlogPost) ETag() string {
	h := sha256.Sum256([]byte(p.ID.String() + p.UpdatedAt.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(h[:8])
}

// Tag is a label attached to blog posts (many-to-many).
type Tag struct {
	ID   uuid.UUID
	Name string
	Slug string
}

// AdminUser is the single portfolio owner with dashboard access.
type AdminUser struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

// Asset is a binary file (image) stored in object storage, referenced by URL in Markdown.
type Asset struct {
	ID          uuid.UUID
	Key         string
	URL         string
	ContentType string
	SizeBytes   int64
	UploadedAt  time.Time
}

// AssetKeyPrefix infers a storage key prefix from the content type.
func AssetKeyPrefix(contentType string) string {
	switch {
	case len(contentType) > 5 && contentType[:5] == "image":
		return "images"
	default:
		return "files"
	}
}

// ValidateAssetContentType returns an error if the content type is not an allowed image type.
func ValidateAssetContentType(ct string) error {
	allowed := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/gif": true,
		"image/webp": true, "image/avif": true, "image/svg+xml": true,
	}
	if !allowed[ct] {
		return fmt.Errorf("%w: unsupported content type %q", ErrValidation, ct)
	}
	return nil
}
