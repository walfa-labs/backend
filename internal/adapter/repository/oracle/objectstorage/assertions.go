package objectstorage

import (
	"github.com/walfa-labs/backend/internal/port"
)

// Compile-time assertion that AssetStore satisfies the port interface.
var _ port.AssetStore = (*AssetStore)(nil)
