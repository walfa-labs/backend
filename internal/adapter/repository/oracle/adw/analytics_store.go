// Package adw implements the analytics port against Oracle Autonomous Data
// Warehouse (ADW) using database/sql with the godror driver.
package adw

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/port"
)

// AnalyticsStore implements port.AnalyticsStore against the ADW star schema
// (dim_posts + fact_post_views). Unlike the OLTP stats repo, it counts
// individual view events, not a running counter.
type AnalyticsStore struct {
	db *sql.DB
}

// NewAnalyticsStore constructs an AnalyticsStore bound to the given handle.
func NewAnalyticsStore(db *sql.DB) *AnalyticsStore {
	return &AnalyticsStore{db: db}
}

// RecordPostView upserts the post dimension row and appends one view fact in a
// single transaction. Called best-effort on every public post read.
func (s *AnalyticsStore) RecordPostView(ctx context.Context, view port.PostView) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	if _, err := tx.ExecContext(ctx, `
		MERGE INTO dim_posts d
		USING (SELECT :1 AS post_id, :2 AS slug, :3 AS title FROM dual) s
		ON (d.post_id = s.post_id)
		WHEN MATCHED THEN UPDATE SET d.slug = s.slug, d.title = s.title
		WHEN NOT MATCHED THEN INSERT (post_id, slug, title) VALUES (s.post_id, s.slug, s.title)`,
		view.PostID.String(), view.Slug, view.Title); err != nil {
		return fmt.Errorf("merge dim_posts: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO fact_post_views (view_id, post_id, viewed_at) VALUES (:1, :2, :3)`,
		uuid.New().String(), view.PostID.String(), view.ViewedAt); err != nil {
		return fmt.Errorf("insert fact_post_views: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// TotalViews returns the total number of recorded view events.
func (s *AnalyticsStore) TotalViews(ctx context.Context) (int64, error) {
	var total int64
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM fact_post_views`).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// ViewsTimeSeries returns cumulative view-count buckets across [from, to].
// View events are grouped by truncating viewed_at to the requested interval;
// each bucket holds the running total up to and including that bucket.
func (s *AnalyticsStore) ViewsTimeSeries(ctx context.Context, from, to time.Time, bucket string) ([]port.ViewsBucket, error) {
	unit, ok := truncUnits[bucket]
	if !ok {
		return nil, fmt.Errorf("unsupported bucket: %q", bucket)
	}

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT TRUNC(viewed_at, '%s') AS b, COUNT(*)
		FROM fact_post_views
		WHERE viewed_at >= :1 AND viewed_at <= :2
		GROUP BY TRUNC(viewed_at, '%s')
		ORDER BY b ASC`, unit, unit), from, to)
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

// TopPosts returns the most-viewed posts, ranked by recorded view events.
func (s *AnalyticsStore) TopPosts(ctx context.Context, limit int) ([]port.TopPost, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT f.post_id, d.slug, d.title, COUNT(*) AS views
		FROM fact_post_views f
		JOIN dim_posts d ON d.post_id = f.post_id
		GROUP BY f.post_id, d.slug, d.title
		ORDER BY views DESC
		FETCH FIRST :1 ROWS ONLY`, limit)
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

// truncUnits maps the public bucket name to an Oracle TRUNC unit.
var truncUnits = map[string]string{
	"hour":  "HH",
	"day":   "DD",
	"week":  "IW",
	"month": "MM",
	"year":  "YYYY",
}
