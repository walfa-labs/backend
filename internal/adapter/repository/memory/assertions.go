package memory

import "github.com/walfa-labs/backend/internal/port"

// Compile-time assertions that each memory repository satisfies its port interface.
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
	_ port.AssetStore     = (*AssetStore)(nil)
)
