package chunks

import (
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// ChunkSection represents a 16x16x16 section of blocks and a 4x4x4 grid of biomes.
type ChunkSection struct {
	BlockCount int16
	// FluidCount is the number of fluid blocks in the section (added in 26.1).
	FluidCount  int16
	BlockStates *PalettedContainer
	Biomes      *PalettedContainer
}

// NewEmptySection creates a section with all air blocks and a single-value biome (ID 0).
func NewEmptySection() *ChunkSection {
	return &ChunkSection{
		BlockCount:  0,
		FluidCount:  0,
		BlockStates: NewSingleValue(BlockStatesKind, 0),
		Biomes:      NewSingleValue(BiomesKind, 0),
	}
}

// GetBlockState returns the block state ID at local coordinates (0-15 each).
func (s *ChunkSection) GetBlockState(x, y, z int) int32 {
	return s.BlockStates.GetXYZ(x, y, z)
}

// SetBlockState sets the block state ID at local coordinates (0-15 each).
func (s *ChunkSection) SetBlockState(x, y, z int, stateID int32) {
	s.BlockStates.SetXYZ(x, y, z, stateID)
}

// GetBiome returns the biome ID at local biome coordinates (0-3 each).
func (s *ChunkSection) GetBiome(x, y, z int) int32 {
	return s.Biomes.GetXYZ(x, y, z)
}

// SetBiome sets the biome ID at local biome coordinates (0-3 each).
func (s *ChunkSection) SetBiome(x, y, z int, biomeID int32) {
	s.Biomes.SetXYZ(x, y, z, biomeID)
}

// Decode reads a ChunkSection from the buffer.
func (s *ChunkSection) Decode(buf *ns.PacketBuffer) error {
	bc, err := buf.ReadInt16()
	if err != nil {
		return err
	}
	s.BlockCount = int16(bc)

	fc, err := buf.ReadInt16()
	if err != nil {
		return err
	}
	s.FluidCount = int16(fc)

	s.BlockStates = &PalettedContainer{kind: BlockStatesKind}
	if err := s.BlockStates.Decode(buf); err != nil {
		return err
	}

	s.Biomes = &PalettedContainer{kind: BiomesKind}
	return s.Biomes.Decode(buf)
}

// Encode writes a ChunkSection to the buffer.
func (s *ChunkSection) Encode(buf *ns.PacketBuffer) error {
	if err := buf.WriteInt16(ns.Int16(s.BlockCount)); err != nil {
		return err
	}
	if err := buf.WriteInt16(ns.Int16(s.FluidCount)); err != nil {
		return err
	}
	if err := s.BlockStates.Encode(buf); err != nil {
		return err
	}
	return s.Biomes.Encode(buf)
}
