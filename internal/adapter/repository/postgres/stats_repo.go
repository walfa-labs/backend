package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/walfa-labs/backend/internal/port"
)

// StatsRepo implements port.StatsRepo against PostgreSQL.
type StatsRepo struct {
	pool *pgxpool.Pool
}

// NewStatsRepo constructs a StatsRepo bound to the given pool.
func NewStatsRepo(pool *pgxpool.Pool) *StatsRepo {
	return &StatsRepo{pool: pool}
}

// Summary returns aggregate counts for the public stats endpoint.
func (r *StatsRepo) Summary(ctx context.Context) (port.StatsSummary, error) {
	var s port.StatsSummary
	var minStart *time.Time

	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM blog_post WHERE status = 'published') AS published_posts,
			(SELECT COUNT(*) FROM project WHERE status = 'published') AS published_projects,
			(SELECT COUNT(*) FROM project WHERE featured = true AND status = 'published') AS featured_projects,
			(SELECT COALESCE(SUM(view_count), 0) FROM blog_post) AS total_views,
			(SELECT MIN(start_date) FROM experience WHERE experience_type = 'work') AS min_start`).Scan(
		&s.PublishedPosts, &s.PublishedProjects, &s.FeaturedProjects, &s.TotalPostViews, &minStart,
	)
	if err != nil {
		return s, err
	}

	if minStart != nil {
		years := time.Since(*minStart).Hours() / 24 / 365.25
		s.YearsExperience = int(years)
	}
	return s, nil
}

// ViewsTimeSeries returns cumulative view-count buckets across [from, to].
// Since we only track a running view_count (not per-view events), we bucket
// by truncating created_at to the requested interval and sum within each bucket.
func (r *StatsRepo) ViewsTimeSeries(ctx context.Context, from, to time.Time, bucket string) ([]port.ViewsBucket, error) {
	truncFn, err := bucketTruncator(bucket)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT date_trunc('%s', created_at) AS b, COALESCE(SUM(view_count), 0)
		FROM blog_post
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY b
		ORDER BY b ASC`, truncFn), from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buckets []port.ViewsBucket
	var cumulative int64
	for rows.Next() {
		var b time.Time
		var v int64
		if err := rows.Scan(&b, &v); err != nil {
			return nil, err
		}
		cumulative += v
		buckets = append(buckets, port.ViewsBucket{Bucket: b, Views: cumulative})
	}
	return buckets, rows.Err()
}

// TopPosts returns the most-viewed published posts.
func (r *StatsRepo) TopPosts(ctx context.Context, limit int) ([]port.TopPost, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, title, view_count
		FROM blog_post
		WHERE status = 'published'
		ORDER BY view_count DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []port.TopPost
	for rows.Next() {
		var t port.TopPost
		if err := rows.Scan(&t.ID, &t.Slug, &t.Title, &t.Views); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// bucketTruncator maps the public bucket name to a Postgres date_trunc unit.
func bucketTruncator(bucket string) (string, error) {
	switch bucket {
	case "hour":
		return "hour", nil
	case "day":
		return "day", nil
	case "week":
		return "week", nil
	case "month":
		return "month", nil
	case "year":
		return "year", nil
	default:
		return "", fmt.Errorf("unsupported bucket: %q", bucket)
	}
}
