package registries

import "maps"

import "fmt"

var synchronizedSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(SynchronizedRegistryIDs))
	for _, id := range SynchronizedRegistryIDs {
		m[id] = struct{}{}
	}
	return m
}()

// IsSynchronized reports whether the given registry identifier is sent over the
// network during the configuration phase.
func IsSynchronized(id string) bool {
	_, ok := synchronizedSet[id]
	return ok
}

// RegistryAccess holds the complete set of registries for a connection.
// Static registries (blocks, items, etc.) point to the shared global instances.
// Synchronized registries (biomes, dimension types, etc.) are data-driven and
// populated from S2CRegistryData packets during the configuration phase.
type RegistryAccess struct {
	registries map[string]*Registry
}

// NewRegistryAccess creates a RegistryAccess initialized with vanilla defaults.
// Static registries share the global instances (cheap).
// Synchronized registries start empty and must be populated via ApplyRegistryData.
func NewRegistryAccess() *RegistryAccess {
	ra := &RegistryAccess{
		registries: make(map[string]*Registry, len(ByIdentifier)+len(SynchronizedRegistryIDs)),
	}
	// add static registries (shared, immutable)
	maps.Copy(ra.registries, ByIdentifier)
	// add empty placeholders for synchronized registries
	for _, id := range SynchronizedRegistryIDs {
		ra.registries[id] = newRegistry(id, -1, nil)
	}
	return ra
}

// Lookup returns the registry for the given identifier (e.g., "minecraft:block").
// Returns nil if the identifier is unknown.
func (ra *RegistryAccess) Lookup(id string) *Registry {
	return ra.registries[id]
}

// ApplyRegistryData replaces a synchronized registry's entries from packet data.
// entryIDs are the ordered entry identifiers from S2CRegistryData;
// their index becomes the numeric protocol ID (0, 1, 2, ...).
func (ra *RegistryAccess) ApplyRegistryData(registryID string, entryIDs []string) (*Registry, error) {
	existing := ra.registries[registryID]
	if existing == nil {
		return nil, fmt.Errorf("registries: unknown registry %q", registryID)
	}
	reg := newRegistryFromOrdered(registryID, existing.ProtocolID, entryIDs)
	ra.registries[registryID] = reg
	return reg, nil
}
