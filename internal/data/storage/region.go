// Package storage implements Minecraft's Anvil region file format for
// persisting chunk and player data.
package storage

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
)

const (
	sectorSize      = 4096
	headerSectors   = 2 // location table + timestamp table
	headerBytes     = headerSectors * sectorSize
	chunksPerAxis   = 32
	chunksPerFile   = chunksPerAxis * chunksPerAxis
	compressionZlib = 2
)

// RegionFile handles reading and writing chunks in one .mca region file.
// Each region file holds up to 32×32 = 1024 chunks.
type RegionFile struct {
	path       string
	file       *os.File
	mu         sync.Mutex
	locations  [chunksPerFile]uint32 // sector offset (24 bits) | sector count (8 bits)
	timestamps [chunksPerFile]uint32
	sectors    []bool // bitmap of used sectors
}

// OpenRegion opens or creates a region file at the given path.
func OpenRegion(path string) (*RegionFile, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	r := &RegionFile{path: path, file: f}

	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	if stat.Size() < headerBytes {
		// new file — write empty header
		if err := r.initHeader(); err != nil {
			f.Close()
			return nil, err
		}
	} else {
		if err := r.readHeader(); err != nil {
			f.Close()
			return nil, err
		}
	}

	return r, nil
}

func (r *RegionFile) initHeader() error {
	header := make([]byte, headerBytes)
	if _, err := r.file.WriteAt(header, 0); err != nil {
		return err
	}
	r.sectors = make([]bool, headerSectors)
	r.sectors[0] = true
	r.sectors[1] = true
	return nil
}

func (r *RegionFile) readHeader() error {
	var buf [headerBytes]byte
	if _, err := r.file.ReadAt(buf[:], 0); err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	for i := range chunksPerFile {
		r.locations[i] = binary.BigEndian.Uint32(buf[i*4:])
		r.timestamps[i] = binary.BigEndian.Uint32(buf[sectorSize+i*4:])
	}

	// build sector bitmap from file size
	stat, err := r.file.Stat()
	if err != nil {
		return err
	}
	totalSectors := int((stat.Size() + sectorSize - 1) / sectorSize)
	if totalSectors < headerSectors {
		totalSectors = headerSectors
	}
	r.sectors = make([]bool, totalSectors)
	r.sectors[0] = true
	r.sectors[1] = true

	for _, loc := range r.locations {
		if loc == 0 {
			continue
		}
		offset := int(loc >> 8)
		count := int(loc & 0xFF)
		for s := offset; s < offset+count && s < len(r.sectors); s++ {
			r.sectors[s] = true
		}
	}

	return nil
}

func chunkIndex(chunkX, chunkZ int32) int {
	return int(chunkX&31) + int(chunkZ&31)*chunksPerAxis
}

// HasChunk returns true if the region file contains data for the given chunk.
func (r *RegionFile) HasChunk(chunkX, chunkZ int32) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.locations[chunkIndex(chunkX, chunkZ)] != 0
}

// ReadChunk reads and decompresses chunk data. Returns nil, nil if chunk is absent.
func (r *RegionFile) ReadChunk(chunkX, chunkZ int32) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	loc := r.locations[chunkIndex(chunkX, chunkZ)]
	if loc == 0 {
		return nil, nil
	}

	offset := int64(loc>>8) * sectorSize
	count := int(loc & 0xFF)

	// read raw sector data
	raw := make([]byte, count*sectorSize)
	if _, err := r.file.ReadAt(raw, offset); err != nil {
		return nil, fmt.Errorf("read sectors: %w", err)
	}

	// parse chunk header: 4-byte length + 1-byte compression
	length := int(binary.BigEndian.Uint32(raw[0:4]))
	if length < 1 || length > len(raw)-4 {
		return nil, fmt.Errorf("invalid chunk length %d", length)
	}
	compression := raw[4]
	compressed := raw[5 : 4+length]

	switch compression {
	case compressionZlib:
		zr, err := zlib.NewReader(bytes.NewReader(compressed))
		if err != nil {
			return nil, fmt.Errorf("zlib init: %w", err)
		}
		defer zr.Close()
		return io.ReadAll(zr)
	case 3: // no compression
		return compressed, nil
	default:
		return nil, fmt.Errorf("unsupported compression type %d", compression)
	}
}

// WriteChunk compresses and writes chunk data to the region file.
func (r *RegionFile) WriteChunk(chunkX, chunkZ int32, data []byte) error {
	// compress with zlib
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		return fmt.Errorf("zlib write: %w", err)
	}
	if err := zw.Close(); err != nil {
		return fmt.Errorf("zlib close: %w", err)
	}
	compressed := buf.Bytes()

	// chunk payload: 4-byte length + 1-byte compression + compressed data
	payloadLen := 1 + len(compressed) // compression byte + data
	totalLen := 4 + payloadLen        // length field + payload
	neededSectors := (totalLen + sectorSize - 1) / sectorSize

	r.mu.Lock()
	defer r.mu.Unlock()

	idx := chunkIndex(chunkX, chunkZ)

	// free old sectors
	oldLoc := r.locations[idx]
	if oldLoc != 0 {
		oldOffset := int(oldLoc >> 8)
		oldCount := int(oldLoc & 0xFF)
		for s := oldOffset; s < oldOffset+oldCount && s < len(r.sectors); s++ {
			r.sectors[s] = false
		}
	}

	// allocate new sectors (first-fit)
	newOffset := r.allocateSectors(neededSectors)

	// write chunk data
	sectorData := make([]byte, neededSectors*sectorSize)
	binary.BigEndian.PutUint32(sectorData[0:4], uint32(payloadLen))
	sectorData[4] = compressionZlib
	copy(sectorData[5:], compressed)

	fileOffset := int64(newOffset) * sectorSize
	if _, err := r.file.WriteAt(sectorData, fileOffset); err != nil {
		return fmt.Errorf("write sectors: %w", err)
	}

	// update location table
	r.locations[idx] = uint32(newOffset)<<8 | uint32(neededSectors)
	r.timestamps[idx] = 0 // TODO: unix timestamp

	return r.writeHeader()
}

func (r *RegionFile) allocateSectors(count int) int {
	// find first contiguous run of `count` free sectors
	for start := headerSectors; start <= len(r.sectors)-count; start++ {
		found := true
		for s := start; s < start+count; s++ {
			if r.sectors[s] {
				found = false
				start = s // skip ahead
				break
			}
		}
		if found {
			for s := start; s < start+count; s++ {
				r.sectors[s] = true
			}
			return start
		}
	}

	// append at end of file
	start := len(r.sectors)
	for range count {
		r.sectors = append(r.sectors, true)
	}
	return start
}

func (r *RegionFile) writeHeader() error {
	var buf [headerBytes]byte
	for i := range chunksPerFile {
		binary.BigEndian.PutUint32(buf[i*4:], r.locations[i])
		binary.BigEndian.PutUint32(buf[sectorSize+i*4:], r.timestamps[i])
	}
	_, err := r.file.WriteAt(buf[:], 0)
	return err
}

// Close closes the underlying file.
func (r *RegionFile) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.file.Close()
}
