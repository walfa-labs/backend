package postgres

import (
	"github.com/walfa-labs/backend/internal/port"
)

// Compile-time interface assertions. These cause a build failure immediately
// if any PostgreSQL repository stops satisfying its port contract.
var (
	_ port.ExperienceRepo = (*ExperienceRepo)(nil)
	_ port.ProjectRepo    = (*ProjectRepo)(nil)
	_ port.PostRepo       = (*PostRepo)(nil)
	_ port.TagRepo        = (*TagRepo)(nil)
	_ port.AssetRepo      = (*AssetRepo)(nil)
	_ port.AdminRepo      = (*AdminRepo)(nil)
	_ port.StatsRepo      = (*StatsRepo)(nil)
	_ port.ProfileRepo    = (*ProfileRepo)(nil)
	_ port.AnalyticsStore = (*AnalyticsStore)(nil)
)
