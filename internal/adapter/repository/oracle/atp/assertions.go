package atp

import (
	"database/sql"

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
	_ port.ProfileRepo    = (*ProfileRepo)(nil)
)

// Compile-time guard documenting the shared dependency of all repos here.
var _ = (*sql.DB)(nil)
