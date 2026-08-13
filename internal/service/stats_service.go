package service

import (
	"context"
	"time"

	"github.com/walfa-labs/backend/internal/port"
)

// StatsService composes operational counts (ATP, via port.StatsRepo) with
// view analytics (ADW, via port.AnalyticsStore).
type StatsService struct {
	repo      port.StatsRepo
	analytics port.AnalyticsStore
}

func NewStatsService(repo port.StatsRepo, analytics port.AnalyticsStore) *StatsService {
	return &StatsService{repo: repo, analytics: analytics}
}

// Summary returns public counts. TotalPostViews comes from the analytics
// store (the analytical source of truth for views); all other counts come
// from the operational store.
func (s *StatsService) Summary(ctx context.Context) (port.StatsSummary, error) {
	summary, err := s.repo.Summary(ctx)
	if err != nil {
		return summary, err
	}
	totalViews, err := s.analytics.TotalViews(ctx)
	if err != nil {
		return summary, err
	}
	summary.TotalPostViews = totalViews
	return summary, nil
}

func (s *StatsService) ViewsTimeSeries(ctx context.Context, from, to time.Time, bucket string) ([]port.ViewsBucket, error) {
	if from.IsZero() {
		from = time.Now().AddDate(0, -1, 0) // default: 1 month back
	}
	if to.IsZero() {
		to = time.Now()
	}
	if bucket == "" {
		bucket = "day"
	}
	return s.analytics.ViewsTimeSeries(ctx, from, to, bucket)
}

func (s *StatsService) TopPosts(ctx context.Context, limit int) ([]port.TopPost, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return s.analytics.TopPosts(ctx, limit)
}
