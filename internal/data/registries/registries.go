package registries

import "maps"

import "iter"

// Registry represents a Minecraft registry.
type Registry struct {
	Identifier string // e.g., "minecraft:block"
	ProtocolID int32
	entries    map[string]int32
	byID       map[int32]string
}

// Get returns the protocol ID for an entry, or -1 if not found.
func (r *Registry) Get(id string) int32 {
	if v, ok := r.entries[id]; ok {
		return v
	}
	return -1
}

// ByID returns the entry name for a protocol ID, or empty string if not found.
func (r *Registry) ByID(protocolID int32) string {
	return r.byID[protocolID]
}

// Size returns the number of entries in the registry.
func (r *Registry) Size() int {
	return len(r.entries)
}

// Entries iterates over all entries in the registry.
func (r *Registry) Entries() iter.Seq2[string, int32] {
	return func(yield func(string, int32) bool) {
		for k, v := range r.entries {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Clone returns an independent copy of the registry.
func (r *Registry) Clone() *Registry {
	entries := make(map[string]int32, len(r.entries))
	maps.Copy(entries, r.entries)
	return newRegistry(r.Identifier, r.ProtocolID, entries)
}

func newRegistry(identifier string, protocolID int32, entries map[string]int32) *Registry {
	byID := make(map[int32]string, len(entries))
	for k, v := range entries {
		byID[v] = k
	}
	return &Registry{Identifier: identifier, ProtocolID: protocolID, entries: entries, byID: byID}
}

// newRegistryFromOrdered creates a registry from an ordered list of entry identifiers.
// The index of each entry becomes its protocol ID.
func newRegistryFromOrdered(identifier string, protocolID int32, entryIDs []string) *Registry {
	entries := make(map[string]int32, len(entryIDs))
	for i, id := range entryIDs {
		entries[id] = int32(i)
	}
	return newRegistry(identifier, protocolID, entries)
}
