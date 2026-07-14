package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const maxOpenRegions = 32

// RegionStorage manages a directory of region files with an LRU cache.
type RegionStorage struct {
	dir string
	mu  sync.Mutex

	// LRU cache of open region files
	cache    map[int64]*regionEntry
	lruOrder []int64 // oldest first
}

type regionEntry struct {
	region *RegionFile
}

// NewRegionStorage creates a storage backed by .mca files in dir.
func NewRegionStorage(dir string) (*RegionStorage, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &RegionStorage{
		dir:   dir,
		cache: make(map[int64]*regionEntry),
	}, nil
}

func regionKey(rx, rz int32) int64 {
	return int64(rx)<<32 | int64(uint32(rz))
}

func regionCoords(chunkX, chunkZ int32) (int32, int32) {
	return chunkX >> 5, chunkZ >> 5
}

func (rs *RegionStorage) getRegion(chunkX, chunkZ int32) (*RegionFile, error) {
	rx, rz := regionCoords(chunkX, chunkZ)
	key := regionKey(rx, rz)

	rs.mu.Lock()
	defer rs.mu.Unlock()

	if entry, ok := rs.cache[key]; ok {
		rs.touchLRU(key)
		return entry.region, nil
	}

	// evict if at capacity
	for len(rs.cache) >= maxOpenRegions {
		rs.evictOldest()
	}

	path := filepath.Join(rs.dir, fmt.Sprintf("r.%d.%d.mca", rx, rz))
	region, err := OpenRegion(path)
	if err != nil {
		return nil, err
	}

	rs.cache[key] = &regionEntry{region: region}
	rs.lruOrder = append(rs.lruOrder, key)
	return region, nil
}

func (rs *RegionStorage) touchLRU(key int64) {
	for i, k := range rs.lruOrder {
		if k == key {
			rs.lruOrder = append(rs.lruOrder[:i], rs.lruOrder[i+1:]...)
			rs.lruOrder = append(rs.lruOrder, key)
			return
		}
	}
}

func (rs *RegionStorage) evictOldest() {
	if len(rs.lruOrder) == 0 {
		return
	}
	key := rs.lruOrder[0]
	rs.lruOrder = rs.lruOrder[1:]
	if entry, ok := rs.cache[key]; ok {
		entry.region.Close()
		delete(rs.cache, key)
	}
}

// ReadChunk reads a chunk from the appropriate region file. Returns nil, nil if absent.
func (rs *RegionStorage) ReadChunk(chunkX, chunkZ int32) ([]byte, error) {
	region, err := rs.getRegion(chunkX, chunkZ)
	if err != nil {
		return nil, err
	}
	return region.ReadChunk(chunkX, chunkZ)
}

// WriteChunk writes a chunk to the appropriate region file.
func (rs *RegionStorage) WriteChunk(chunkX, chunkZ int32, data []byte) error {
	region, err := rs.getRegion(chunkX, chunkZ)
	if err != nil {
		return err
	}
	return region.WriteChunk(chunkX, chunkZ, data)
}

// HasChunk checks if a chunk exists in the appropriate region file.
func (rs *RegionStorage) HasChunk(chunkX, chunkZ int32) bool {
	region, err := rs.getRegion(chunkX, chunkZ)
	if err != nil {
		return false
	}
	return region.HasChunk(chunkX, chunkZ)
}

// Close closes all open region files.
func (rs *RegionStorage) Close() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	for _, entry := range rs.cache {
		entry.region.Close()
	}
	rs.cache = make(map[int64]*regionEntry)
	rs.lruOrder = nil
	return nil
}
