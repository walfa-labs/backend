package atp

import (
	"context"
	"database/sql"
	"time"

	"github.com/walfa-labs/backend/internal/port"
)

// StatsRepo implements port.StatsRepo against Oracle ATP. It covers only the
// operational counts; view analytics (TotalPostViews, time series, top posts)
// live in the ADW analytics store and are composed by the service layer.
type StatsRepo struct {
	db *sql.DB
}

// NewStatsRepo constructs a StatsRepo bound to the given pool.
func NewStatsRepo(db *sql.DB) *StatsRepo {
	return &StatsRepo{db: db}
}

// Summary returns aggregate counts for the public stats endpoint.
// TotalPostViews is intentionally left zero here — the service layer fills it
// from the analytics store (ADW), the analytical source of truth for views.
func (r *StatsRepo) Summary(ctx context.Context) (port.StatsSummary, error) {
	var s port.StatsSummary
	var minStart sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT
		    (SELECT COUNT(*) FROM blog_posts WHERE status = 'published'),
		    (SELECT COUNT(*) FROM projects WHERE status = 'published'),
		    (SELECT COUNT(*) FROM projects WHERE featured = 1 AND status = 'published'),
		    (SELECT MIN(start_date) FROM experiences WHERE experience_type = 'work')
		FROM dual`).Scan(
		&s.PublishedPosts, &s.PublishedProjects, &s.FeaturedProjects, &minStart,
	)
	if err != nil {
		return s, err
	}

	if minStart.Valid {
		years := time.Since(minStart.Time).Hours() / 24 / 365.25
		s.YearsExperience = int(years)
	}
	return s, nil
}
