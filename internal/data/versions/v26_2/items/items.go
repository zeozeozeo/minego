package items

import "github.com/zeozeozeo/minego/internal/data/versions/v26_2/registries"

// ItemID returns the protocol ID for an item identifier, or -1 if not found.
func ItemID(id string) int32 {
	return registries.Item.Get(id)
}

// ItemName returns the identifier for an item protocol ID, or empty string if not found.
func ItemName(protocolID int32) string {
	return registries.Item.ByID(protocolID)
}

// DefaultComponents returns the default components for an item, or nil if not found.
func DefaultComponents(itemID int32) *Components {
	return defaultComponents[itemID]
}
