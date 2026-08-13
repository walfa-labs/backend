package adw

import "github.com/walfa-labs/backend/internal/port"

// Compile-time assertion that the concrete store satisfies its port interface.
var _ port.AnalyticsStore = (*AnalyticsStore)(nil)
