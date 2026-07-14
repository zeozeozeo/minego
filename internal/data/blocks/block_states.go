package blocks

import "sync"

// blockStateData holds the data needed to calculate state IDs.
type blockStateData struct {
	BaseID     int32
	DefaultID  int32
	Properties []blockProperty
}

type blockProperty struct {
	Name        string
	Values      []string
	Cardinality int
}

// stateRangeEntry maps a range of state IDs to a block ID for O(log n) lookup.
type stateRangeEntry struct {
	BaseID, EndID, BlockID int32
}

// cache for StateID lookups (StateProperties uses binary search which is fast enough)
var (
	stateIDCache    = make(map[uint64]int32)
	stateIDCacheMu  sync.RWMutex
	stateIDCacheMax = 4096
)

// SetCacheSize sets the maximum cache size for StateID. Set to 0 to disable caching.
func SetCacheSize(maxEntries int) {
	stateIDCacheMu.Lock()
	stateIDCacheMax = maxEntries
	if maxEntries == 0 {
		stateIDCache = make(map[uint64]int32)
	}
	stateIDCacheMu.Unlock()
}

// ClearCache clears all cached StateID lookups.
func ClearCache() {
	stateIDCacheMu.Lock()
	stateIDCache = make(map[uint64]int32)
	stateIDCacheMu.Unlock()
}

// stateIDCacheKey creates a cache key from blockID and properties
func stateIDCacheKey(blockID int, props map[string]string) uint64 {
	h := uint64(blockID)
	for k, v := range props {
		for _, c := range k {
			h = h*31 + uint64(c)
		}
		for _, c := range v {
			h = h*31 + uint64(c)
		}
	}
	return h
}

// StateID calculates the state protocol ID from a block ID and property values.
// Results are cached for performance. Returns -1 if the block ID is invalid or properties are missing/invalid.
func StateID(blockID int, props map[string]string) int32 {
	if stateIDCacheMax > 0 {
		key := stateIDCacheKey(blockID, props)
		stateIDCacheMu.RLock()
		if result, ok := stateIDCache[key]; ok {
			stateIDCacheMu.RUnlock()
			return result
		}
		stateIDCacheMu.RUnlock()
	}

	result := stateIDUncached(blockID, props)

	if stateIDCacheMax > 0 && result != -1 {
		key := stateIDCacheKey(blockID, props)
		stateIDCacheMu.Lock()
		if len(stateIDCache) >= stateIDCacheMax {
			for k := range stateIDCache {
				delete(stateIDCache, k)
				if len(stateIDCache) < stateIDCacheMax/2 {
					break
				}
			}
		}
		stateIDCache[key] = result
		stateIDCacheMu.Unlock()
	}

	return result
}

// stateIDUncached is the uncached implementation of StateID.
func stateIDUncached(blockID int, props map[string]string) int32 {
	state := blockStates[int32(blockID)]
	if state == nil {
		return -1
	}
	if len(state.Properties) == 0 {
		return state.BaseID
	}

	offset := int32(0)
	multiplier := int32(1)

	for i := len(state.Properties) - 1; i >= 0; i-- {
		p := &state.Properties[i]
		propValue, ok := props[p.Name]
		if !ok {
			return -1
		}
		valueIdx := -1
		for j, v := range p.Values {
			if v == propValue {
				valueIdx = j
				break
			}
		}
		if valueIdx == -1 {
			return -1
		}
		offset += int32(valueIdx) * multiplier
		multiplier *= int32(p.Cardinality)
	}

	return state.BaseID + offset
}

// StateProperties returns the block ID and property values for a given state ID.
// Uses binary search for O(log n) block lookup. Returns -1 and nil if the state ID is invalid.
func StateProperties(stateID int) (blockID int32, props map[string]string) {
	stateID32 := int32(stateID)
	lo, hi := 0, len(stateRanges)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		r := &stateRanges[mid]
		if stateID32 < r.BaseID {
			hi = mid - 1
		} else if stateID32 >= r.EndID {
			lo = mid + 1
		} else {
			blockID = r.BlockID
			state := blockStates[blockID]
			offset := stateID32 - state.BaseID

			props = make(map[string]string)
			for i := len(state.Properties) - 1; i >= 0; i-- {
				p := &state.Properties[i]
				valueIdx := offset % int32(p.Cardinality)
				offset /= int32(p.Cardinality)
				props[p.Name] = p.Values[valueIdx]
			}
			return blockID, props
		}
	}
	return -1, nil
}

// DefaultStateID returns the default state ID for a block, or -1 if invalid.
func DefaultStateID(blockID int32) int32 {
	state := blockStates[blockID]
	if state == nil {
		return -1
	}
	return state.DefaultID
}
