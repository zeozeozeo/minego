package chunks

import (
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

const (
	// SectionCount is the number of vertical sections in a chunk column (1.21.x: Y -64 to 319).
	SectionCount = 24
	// MinY is the minimum world Y coordinate.
	MinY = -64
	// MaxY is the maximum world Y coordinate (exclusive).
	MaxY = MinY + SectionCount*16
)

// ChunkColumn represents a vertical column of chunk sections with associated data.
type ChunkColumn struct {
	X, Z          int32
	Sections      [SectionCount]*ChunkSection
	Heightmaps    map[int32][]int64
	BlockEntities []ns.BlockEntity
	Light         *ns.LightData
}

// ParseChunkColumn decodes a ChunkColumn from protocol-level types.
func ParseChunkColumn(x, z int32, data ns.ChunkData, light *ns.LightData) (*ChunkColumn, error) {
	col := &ChunkColumn{
		X:             x,
		Z:             z,
		Heightmaps:    data.Heightmaps,
		BlockEntities: data.BlockEntities,
		Light:         light,
	}

	buf := ns.NewReader(data.Data)
	for i := range SectionCount {
		sec := &ChunkSection{}
		if err := sec.Decode(buf); err != nil {
			return nil, err
		}
		col.Sections[i] = sec
	}
	return col, nil
}

// GetBlockState returns the block state ID at absolute world coordinates.
// Returns 0 (air) if the coordinates are out of range or the section is nil.
func (c *ChunkColumn) GetBlockState(x, y, z int) int32 {
	idx := SectionIndex(y)
	if idx < 0 {
		return 0
	}
	sec := c.Sections[idx]
	if sec == nil {
		return 0
	}
	lx, ly, lz := LocalCoords(x, y, z)
	return sec.GetBlockState(lx, ly, lz)
}

// SetBlockState sets the block state ID at absolute world coordinates.
// Creates an empty section if the target section is nil.
func (c *ChunkColumn) SetBlockState(x, y, z int, stateID int32) {
	idx := SectionIndex(y)
	if idx < 0 {
		return
	}
	if c.Sections[idx] == nil {
		c.Sections[idx] = NewEmptySection()
	}
	lx, ly, lz := LocalCoords(x, y, z)
	c.Sections[idx].SetBlockState(lx, ly, lz, stateID)
}

// EncodeSections encodes all chunk sections back to raw bytes (the ChunkData.Data portion).
func (c *ChunkColumn) EncodeSections() ([]byte, error) {
	buf := ns.NewWriter()
	for _, sec := range c.Sections {
		if sec == nil {
			sec = NewEmptySection()
		}
		if err := sec.Encode(buf); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// SectionIndex returns the section array index for the given world Y coordinate.
// Returns -1 if Y is out of range.
func SectionIndex(worldY int) int {
	idx := (worldY - MinY) >> 4
	if idx < 0 || idx >= SectionCount {
		return -1
	}
	return idx
}

// LocalCoords extracts the local 0-15 coordinates from world coordinates.
func LocalCoords(worldX, worldY, worldZ int) (int, int, int) {
	return worldX & 0xF, worldY & 0xF, worldZ & 0xF
}

// ChunkPos returns the chunk coordinates for the given world X, Z.
func ChunkPos(worldX, worldZ int) (int32, int32) {
	return int32(worldX >> 4), int32(worldZ >> 4)
}
