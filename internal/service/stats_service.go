package service

import (
	"context"
	"time"

	"github.com/walfa-labs/backend/internal/port"
)

type StatsService struct {
	repo port.StatsRepo
}

func NewStatsService(repo port.StatsRepo) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) Summary(ctx context.Context) (port.StatsSummary, error) {
	return s.repo.Summary(ctx)
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
	return s.repo.ViewsTimeSeries(ctx, from, to, bucket)
}

func (s *StatsService) TopPosts(ctx context.Context, limit int) ([]port.TopPost, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return s.repo.TopPosts(ctx, limit)
}
