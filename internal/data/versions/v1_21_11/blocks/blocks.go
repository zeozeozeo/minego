package blocks

import "github.com/zeozeozeo/minego/internal/data/versions/v1_21_11/registries"

// BlockID returns the protocol ID for a block identifier, or -1 if not found.
func BlockID(id string) int32 {
	return registries.Block.Get(id)
}

// BlockName returns the identifier for a block protocol ID, or empty string if not found.
func BlockName(protocolID int32) string {
	return registries.Block.ByID(protocolID)
}
