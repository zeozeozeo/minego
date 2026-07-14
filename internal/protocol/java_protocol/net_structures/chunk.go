package net_structures

import (
	"encoding/binary"
	"fmt"

	"github.com/zeozeozeo/minego/internal/protocol/nbt"
)

// ChunkData represents the data portion of chunk packets.
// Heightmaps are encoded as a VarInt-keyed map of long arrays.
// Chunk sections are raw bytes. Block entities follow.
//
// Wire format:
//
//	┌─────────────────────────────────────────────────────────────────────────┐
//	│  Heightmaps (VarInt count + entries of VarInt key + VarInt len + longs) │
//	├─────────────────────────────────────────────────────────────────────────┤
//	│  Data (VarInt length + raw bytes containing chunk sections)             │
//	├─────────────────────────────────────────────────────────────────────────┤
//	│  BlockEntities (VarInt length + array of BlockEntity)                   │
//	└─────────────────────────────────────────────────────────────────────────┘
type ChunkData struct {
	// Heightmaps maps heightmap type IDs to long arrays.
	// Type IDs: 1=WORLD_SURFACE, 4=MOTION_BLOCKING, 5=MOTION_BLOCKING_NO_LEAVES.
	Heightmaps map[int32][]int64

	// Data contains packed chunk sections. Each section contains:
	// - Block count (short)
	// - Block states (paletted container)
	// - Biomes (paletted container)
	Data []byte

	// BlockEntities in this chunk.
	BlockEntities []BlockEntity
}

// BlockEntity represents a block entity within a chunk.
//
// Wire format:
//
//	┌──────────────────┬────────────────────┬─────────────────┬──────────────────┐
//	│  PackedXZ (byte) │  Y (short)         │  Type (VarInt)  │  Data (NBT)      │
//	└──────────────────┴────────────────────┴─────────────────┴──────────────────┘
//
// PackedXZ encodes relative X and Z coordinates:
//
//	packed = ((blockX & 15) << 4) | (blockZ & 15)
//	x = packed >> 4, z = packed & 15
type BlockEntity struct {
	// PackedXZ contains relative X (high nibble) and Z (low nibble) coordinates.
	PackedXZ Uint8

	// Y is the absolute Y coordinate.
	Y Int16

	// Type is the block entity type registry ID.
	Type VarInt

	// Data is the block entity's NBT data (without x, y, z fields).
	Data nbt.Tag
}

// X returns the relative X coordinate (0-15) from PackedXZ.
func (b *BlockEntity) X() int {
	return int(b.PackedXZ >> 4)
}

// Z returns the relative Z coordinate (0-15) from PackedXZ.
func (b *BlockEntity) Z() int {
	return int(b.PackedXZ & 15)
}

// SetXZ sets the PackedXZ field from relative X and Z coordinates.
func (b *BlockEntity) SetXZ(x, z int) {
	b.PackedXZ = Uint8(((x & 15) << 4) | (z & 15))
}

// Decode reads ChunkData from the buffer.
func (c *ChunkData) Decode(buf *PacketBuffer) error {
	// read heightmaps map: VarInt count, then (VarInt key, VarInt len, Int64[len]) entries
	hmCount, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read heightmap count: %w", err)
	}
	c.Heightmaps = make(map[int32][]int64, hmCount)
	for range int(hmCount) {
		key, err := buf.ReadVarInt()
		if err != nil {
			return fmt.Errorf("failed to read heightmap type: %w", err)
		}
		arrLen, err := buf.ReadVarInt()
		if err != nil {
			return fmt.Errorf("failed to read heightmap array length: %w", err)
		}
		longs := make([]int64, arrLen)
		for j := range longs {
			var b [8]byte
			if _, err := buf.Read(b[:]); err != nil {
				return fmt.Errorf("failed to read heightmap long %d: %w", j, err)
			}
			longs[j] = int64(binary.BigEndian.Uint64(b[:]))
		}
		c.Heightmaps[int32(key)] = longs
	}

	// read chunk data as byte array (max ~2MB for full chunk)
	c.Data, err = buf.ReadByteArray(2097152)
	if err != nil {
		return fmt.Errorf("failed to read chunk data: %w", err)
	}

	// read block entities
	count, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read block entity count: %w", err)
	}

	c.BlockEntities = make([]BlockEntity, count)
	for i := range c.BlockEntities {
		if err := c.BlockEntities[i].Decode(buf); err != nil {
			return fmt.Errorf("failed to read block entity %d: %w", i, err)
		}
	}

	return nil
}

// Encode writes ChunkData to the buffer.
func (c *ChunkData) Encode(buf *PacketBuffer) error {
	// write heightmaps map
	if err := buf.WriteVarInt(VarInt(len(c.Heightmaps))); err != nil {
		return fmt.Errorf("failed to write heightmap count: %w", err)
	}
	for key, longs := range c.Heightmaps {
		if err := buf.WriteVarInt(VarInt(key)); err != nil {
			return fmt.Errorf("failed to write heightmap type: %w", err)
		}
		if err := buf.WriteVarInt(VarInt(len(longs))); err != nil {
			return fmt.Errorf("failed to write heightmap array length: %w", err)
		}
		for _, v := range longs {
			var b [8]byte
			binary.BigEndian.PutUint64(b[:], uint64(v))
			if _, err := buf.Write(b[:]); err != nil {
				return fmt.Errorf("failed to write heightmap long: %w", err)
			}
		}
	}

	// write chunk data
	if err := buf.WriteByteArray(c.Data); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}

	// write block entities
	if err := buf.WriteVarInt(VarInt(len(c.BlockEntities))); err != nil {
		return fmt.Errorf("failed to write block entity count: %w", err)
	}
	for i := range c.BlockEntities {
		if err := c.BlockEntities[i].Encode(buf); err != nil {
			return fmt.Errorf("failed to write block entity %d: %w", i, err)
		}
	}

	return nil
}

// Decode reads a BlockEntity from the buffer.
func (b *BlockEntity) Decode(buf *PacketBuffer) error {
	var err error

	packed, err := buf.ReadUint8()
	if err != nil {
		return fmt.Errorf("failed to read packed xz: %w", err)
	}
	b.PackedXZ = packed

	b.Y, err = buf.ReadInt16()
	if err != nil {
		return fmt.Errorf("failed to read y: %w", err)
	}

	b.Type, err = buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read type: %w", err)
	}

	nbtReader := nbt.NewReaderFrom(buf.Reader())
	b.Data, _, err = nbtReader.ReadTag(true)
	if err != nil {
		return fmt.Errorf("failed to read nbt data: %w", err)
	}

	return nil
}

// Encode writes a BlockEntity to the buffer.
func (b *BlockEntity) Encode(buf *PacketBuffer) error {
	if err := buf.WriteUint8(b.PackedXZ); err != nil {
		return fmt.Errorf("failed to write packed xz: %w", err)
	}

	if err := buf.WriteInt16(b.Y); err != nil {
		return fmt.Errorf("failed to write y: %w", err)
	}

	if err := buf.WriteVarInt(b.Type); err != nil {
		return fmt.Errorf("failed to write type: %w", err)
	}

	if b.Data == nil {
		b.Data = nbt.Compound{}
	}
	nbtData, err := nbt.Encode(b.Data, "", true)
	if err != nil {
		return fmt.Errorf("failed to encode nbt data: %w", err)
	}
	if _, err := buf.Write(nbtData); err != nil {
		return fmt.Errorf("failed to write nbt data: %w", err)
	}

	return nil
}

// ReadChunkData reads ChunkData from the buffer.
func (pb *PacketBuffer) ReadChunkData() (ChunkData, error) {
	var c ChunkData
	err := c.Decode(pb)
	return c, err
}

// WriteChunkData writes ChunkData to the buffer.
func (pb *PacketBuffer) WriteChunkData(c ChunkData) error {
	return c.Encode(pb)
}

// LightData represents lighting information for a chunk.
// Contains bit masks indicating which sections have light data,
// and the actual light arrays.
//
// Wire format:
//
//	┌─────────────────────────────────────────────────────────────────────────┐
//	│  SkyLightMask (BitSet)                                                  │
//	├─────────────────────────────────────────────────────────────────────────┤
//	│  BlockLightMask (BitSet)                                                │
//	├─────────────────────────────────────────────────────────────────────────┤
//	│  EmptySkyLightMask (BitSet)                                             │
//	├─────────────────────────────────────────────────────────────────────────┤
//	│  EmptyBlockLightMask (BitSet)                                           │
//	├─────────────────────────────────────────────────────────────────────────┤
//	│  SkyLightArrays (VarInt count + arrays of 2048 bytes each)              │
//	├─────────────────────────────────────────────────────────────────────────┤
//	│  BlockLightArrays (VarInt count + arrays of 2048 bytes each)            │
//	└─────────────────────────────────────────────────────────────────────────┘
//
// Each BitSet contains bits for each section in the world + 2 (one above and below).
// Light arrays are 2048 bytes each (4096 nibbles for 16×16×16 blocks).
type LightData struct {
	// SkyLightMask indicates which sections have sky light data.
	SkyLightMask BitSet

	// BlockLightMask indicates which sections have block light data.
	BlockLightMask BitSet

	// EmptySkyLightMask indicates which sections have all-zero sky light.
	EmptySkyLightMask BitSet

	// EmptyBlockLightMask indicates which sections have all-zero block light.
	EmptyBlockLightMask BitSet

	// SkyLightArrays contains sky light data for sections with SkyLightMask bit set.
	// Each array is 2048 bytes (half a byte per block, 16×16×16 = 4096 blocks).
	SkyLightArrays [][]byte

	// BlockLightArrays contains block light data for sections with BlockLightMask bit set.
	// Each array is 2048 bytes.
	BlockLightArrays [][]byte
}

// Decode reads LightData from the buffer.
func (l *LightData) Decode(buf *PacketBuffer) error {
	if err := l.SkyLightMask.Decode(buf); err != nil {
		return fmt.Errorf("failed to read sky light mask: %w", err)
	}

	if err := l.BlockLightMask.Decode(buf); err != nil {
		return fmt.Errorf("failed to read block light mask: %w", err)
	}

	if err := l.EmptySkyLightMask.Decode(buf); err != nil {
		return fmt.Errorf("failed to read empty sky light mask: %w", err)
	}

	if err := l.EmptyBlockLightMask.Decode(buf); err != nil {
		return fmt.Errorf("failed to read empty block light mask: %w", err)
	}

	// read sky light arrays (each is 2048 bytes)
	skyCount, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read sky light array count: %w", err)
	}
	l.SkyLightArrays = make([][]byte, skyCount)
	for i := range l.SkyLightArrays {
		l.SkyLightArrays[i], err = buf.ReadByteArray(2048)
		if err != nil {
			return fmt.Errorf("failed to read sky light array %d: %w", i, err)
		}
	}

	// read block light arrays (each is 2048 bytes)
	blockCount, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read block light array count: %w", err)
	}
	l.BlockLightArrays = make([][]byte, blockCount)
	for i := range l.BlockLightArrays {
		l.BlockLightArrays[i], err = buf.ReadByteArray(2048)
		if err != nil {
			return fmt.Errorf("failed to read block light array %d: %w", i, err)
		}
	}

	return nil
}

// Encode writes LightData to the buffer.
func (l *LightData) Encode(buf *PacketBuffer) error {
	if err := l.SkyLightMask.Encode(buf); err != nil {
		return fmt.Errorf("failed to write sky light mask: %w", err)
	}

	if err := l.BlockLightMask.Encode(buf); err != nil {
		return fmt.Errorf("failed to write block light mask: %w", err)
	}

	if err := l.EmptySkyLightMask.Encode(buf); err != nil {
		return fmt.Errorf("failed to write empty sky light mask: %w", err)
	}

	if err := l.EmptyBlockLightMask.Encode(buf); err != nil {
		return fmt.Errorf("failed to write empty block light mask: %w", err)
	}

	// write sky light arrays
	if err := buf.WriteVarInt(VarInt(len(l.SkyLightArrays))); err != nil {
		return fmt.Errorf("failed to write sky light array count: %w", err)
	}
	for i, arr := range l.SkyLightArrays {
		if err := buf.WriteByteArray(arr); err != nil {
			return fmt.Errorf("failed to write sky light array %d: %w", i, err)
		}
	}

	// write block light arrays
	if err := buf.WriteVarInt(VarInt(len(l.BlockLightArrays))); err != nil {
		return fmt.Errorf("failed to write block light array count: %w", err)
	}
	for i, arr := range l.BlockLightArrays {
		if err := buf.WriteByteArray(arr); err != nil {
			return fmt.Errorf("failed to write block light array %d: %w", i, err)
		}
	}

	return nil
}

// ReadLightData reads LightData from the buffer.
func (pb *PacketBuffer) ReadLightData() (LightData, error) {
	var l LightData
	err := l.Decode(pb)
	return l, err
}

// WriteLightData writes LightData to the buffer.
func (pb *PacketBuffer) WriteLightData(l LightData) error {
	return l.Encode(pb)
}
