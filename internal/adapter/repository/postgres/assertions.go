package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/walfa-labs/backend/internal/port"
)

// Compile-time assertions that each concrete repo satisfies its port interface.
var (
	_ port.ExperienceRepo = (*ExperienceRepo)(nil)
	_ port.ProjectRepo    = (*ProjectRepo)(nil)
	_ port.PostRepo       = (*PostRepo)(nil)
	_ port.TagRepo        = (*TagRepo)(nil)
	_ port.AssetRepo      = (*AssetRepo)(nil)
	_ port.AdminRepo      = (*AdminRepo)(nil)
	_ port.StatsRepo      = (*StatsRepo)(nil)
)

// Compile-time guard to keep the pgxpool import meaningful even if constructors
// are reorganized; it documents the shared dependency of all repos here.
var _ = (*pgxpool.Pool)(nil)
