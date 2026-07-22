package s3

import "github.com/walfa-labs/backend/internal/port"

// Compile-time assertion that AssetStore satisfies port.AssetStore.
var _ port.AssetStore = (*AssetStore)(nil)
