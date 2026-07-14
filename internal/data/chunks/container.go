package chunks

import (
	"fmt"
	"math/bits"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// ContainerKind defines the palette strategy parameters for a paletted container.
type ContainerKind struct {
	EntryCount      int
	MinIndirectBits int
	MaxIndirectBits int
	DirectBits      int
}

// DirectBits is the global-palette bit width = ceil(log2(registry size)). It is
// only used when encoding a container that overflows to the direct palette;
// decoding reads the width from the wire (see Decode). Bump these when the
// registries grow past a power of two (26.2: 32366 block states -> 15 bits,
// 66 biomes -> 7 bits).
var (
	BlockStatesKind = ContainerKind{EntryCount: 4096, MinIndirectBits: 4, MaxIndirectBits: 8, DirectBits: 15}
	BiomesKind      = ContainerKind{EntryCount: 64, MinIndirectBits: 1, MaxIndirectBits: 3, DirectBits: 7}
)

// PalettedContainer stores entries (block state IDs or biome IDs) using palette compression.
//
// Three modes:
//   - Single-value (bitsPerEntry=0): all entries share one value
//   - Indirect (bitsPerEntry in [minIndirect, maxIndirect]): palette maps indices to IDs
//   - Direct (bitsPerEntry=directBits): no palette, data stores IDs directly
type PalettedContainer struct {
	kind         ContainerKind
	bitsPerEntry int
	palette      []int32  // nil for single-value and direct modes
	data         []uint64 // packed entries; empty for single-value
	singleValue  int32    // used when bitsPerEntry == 0
}

// NewSingleValue creates a container where all entries have the same value.
func NewSingleValue(kind ContainerKind, value int32) *PalettedContainer {
	return &PalettedContainer{
		kind:         kind,
		bitsPerEntry: 0,
		singleValue:  value,
	}
}

// sideLen returns the side length of the container (16 for blocks, 4 for biomes).
func (p *PalettedContainer) sideLen() int {
	return 1 << (bits.TrailingZeros(uint(p.kind.EntryCount)) / 3)
}

func (p *PalettedContainer) flatIndex(x, y, z int) int {
	side := p.sideLen()
	return (y*side+z)*side + x
}

// dataArrayLen returns the number of uint64s needed for the given bits per entry.
func dataArrayLen(bitsPerEntry, entryCount int) int {
	if bitsPerEntry == 0 {
		return 0
	}
	entriesPerLong := 64 / bitsPerEntry
	return (entryCount + entriesPerLong - 1) / entriesPerLong
}

// Get returns the entry at the given flat index.
func (p *PalettedContainer) Get(index int) int32 {
	if p.bitsPerEntry == 0 {
		return p.singleValue
	}
	entriesPerLong := 64 / p.bitsPerEntry
	longIdx := index / entriesPerLong
	bitIdx := (index % entriesPerLong) * p.bitsPerEntry
	mask := uint64((1 << p.bitsPerEntry) - 1)
	raw := int32((p.data[longIdx] >> bitIdx) & mask)
	if p.palette != nil {
		if int(raw) < len(p.palette) {
			return p.palette[raw]
		}
		return 0
	}
	return raw
}

// GetXYZ returns the entry at the given x, y, z coordinates.
// For blocks: 0-15 each. For biomes: 0-3 each.
func (p *PalettedContainer) GetXYZ(x, y, z int) int32 {
	return p.Get(p.flatIndex(x, y, z))
}

// Set sets the entry at the given flat index.
func (p *PalettedContainer) Set(index int, value int32) {
	if p.bitsPerEntry == 0 {
		if value == p.singleValue {
			return
		}
		p.expandFromSingleValue(value)
	}

	paletteIdx := value
	if p.palette != nil {
		// indirect mode: find or add to palette
		found := -1
		for i, v := range p.palette {
			if v == value {
				found = i
				break
			}
		}
		if found >= 0 {
			paletteIdx = int32(found)
		} else {
			maxPaletteSize := 1 << p.bitsPerEntry
			if len(p.palette) < maxPaletteSize {
				p.palette = append(p.palette, value)
				paletteIdx = int32(len(p.palette) - 1)
			} else {
				p.expandPalette()
				// after expansion, retry
				p.Set(index, value)
				return
			}
		}
	}

	entriesPerLong := 64 / p.bitsPerEntry
	longIdx := index / entriesPerLong
	bitIdx := (index % entriesPerLong) * p.bitsPerEntry
	mask := uint64((1 << p.bitsPerEntry) - 1)
	p.data[longIdx] &= ^(mask << bitIdx)
	p.data[longIdx] |= uint64(paletteIdx) << bitIdx
}

// SetXYZ sets the entry at the given x, y, z coordinates.
func (p *PalettedContainer) SetXYZ(x, y, z int, value int32) {
	p.Set(p.flatIndex(x, y, z), value)
}

// BitsPerEntry returns the current bits per entry.
func (p *PalettedContainer) BitsPerEntry() int {
	return p.bitsPerEntry
}

// expandFromSingleValue converts from single-value to indirect palette with two entries.
func (p *PalettedContainer) expandFromSingleValue(newValue int32) {
	oldValue := p.singleValue
	p.bitsPerEntry = p.kind.MinIndirectBits
	p.palette = []int32{oldValue, newValue}
	p.data = make([]uint64, dataArrayLen(p.bitsPerEntry, p.kind.EntryCount))
	// all entries are index 0 (old value), which is already zero-initialized
}

// expandPalette increases the bits per entry and rebuilds the data array.
func (p *PalettedContainer) expandPalette() {
	oldBPE := p.bitsPerEntry
	oldData := p.data
	oldPalette := p.palette

	newBPE := oldBPE + 1
	if newBPE > p.kind.MaxIndirectBits {
		// switch to direct mode
		p.bitsPerEntry = p.kind.DirectBits
		p.data = make([]uint64, dataArrayLen(p.bitsPerEntry, p.kind.EntryCount))
		for i := range p.kind.EntryCount {
			value := extractEntry(oldData, oldBPE, i)
			if oldPalette != nil && int(value) < len(oldPalette) {
				value = oldPalette[value]
			}
			setEntry(p.data, p.bitsPerEntry, i, value)
		}
		p.palette = nil
		return
	}

	p.bitsPerEntry = newBPE
	p.data = make([]uint64, dataArrayLen(newBPE, p.kind.EntryCount))
	for i := range p.kind.EntryCount {
		value := extractEntry(oldData, oldBPE, i)
		setEntry(p.data, newBPE, i, value)
	}
}

func extractEntry(data []uint64, bpe, index int) int32 {
	entriesPerLong := 64 / bpe
	longIdx := index / entriesPerLong
	bitIdx := (index % entriesPerLong) * bpe
	mask := uint64((1 << bpe) - 1)
	return int32((data[longIdx] >> bitIdx) & mask)
}

func setEntry(data []uint64, bpe, index int, value int32) {
	entriesPerLong := 64 / bpe
	longIdx := index / entriesPerLong
	bitIdx := (index % entriesPerLong) * bpe
	mask := uint64((1 << bpe) - 1)
	data[longIdx] &= ^(mask << bitIdx)
	data[longIdx] |= uint64(value) << bitIdx
}

// Decode reads a PalettedContainer from the buffer.
// Wire format: byte bitsPerEntry | palette | fixed-size long array (no length
// prefix). The long-array length is derived from the bits-per-entry byte.
func (p *PalettedContainer) Decode(buf *ns.PacketBuffer) error {
	bpe, err := buf.ReadUint8()
	if err != nil {
		return err
	}

	if bpe == 0 {
		// single-value mode
		val, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		p.bitsPerEntry = 0
		p.singleValue = int32(val)
		p.palette = nil
		p.data = nil
		return nil
	}

	if int(bpe) <= p.kind.MaxIndirectBits {
		// indirect palette
		p.bitsPerEntry = max(int(bpe), p.kind.MinIndirectBits)

		paletteLen, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		p.palette = make([]int32, paletteLen)
		for i := range p.palette {
			v, err := buf.ReadVarInt()
			if err != nil {
				return err
			}
			p.palette[i] = int32(v)
		}
	} else {
		// direct palette: the byte is the global-palette bit width the server
		// actually packed with (ceil(log2(registry size))). use it verbatim
		// rather than a hardcoded constant, which goes stale as the block/biome
		// registry grows (e.g. biomes crossing 64 -> 6 to 7 bits).
		p.bitsPerEntry = int(bpe)
		p.palette = nil
	}

	// guard against a malformed/desynced bits-per-entry so a bad stream surfaces
	// as a parse error rather than a divide-by-zero panic deeper down.
	if p.bitsPerEntry < 1 || p.bitsPerEntry > 64 {
		return fmt.Errorf("invalid bits-per-entry %d (wire byte %d)", p.bitsPerEntry, bpe)
	}

	// read fixed-size long array (size calculated, no VarInt prefix)
	dataLen := dataArrayLen(p.bitsPerEntry, p.kind.EntryCount)
	p.data = make([]uint64, dataLen)
	for i := range p.data {
		v, err := buf.ReadInt64()
		if err != nil {
			return fmt.Errorf("reading data long %d/%d: %w", i, dataLen, err)
		}
		p.data[i] = uint64(v)
	}
	return nil
}

// Encode writes a PalettedContainer to the buffer.
func (p *PalettedContainer) Encode(buf *ns.PacketBuffer) error {
	if err := buf.WriteUint8(ns.Uint8(p.bitsPerEntry)); err != nil {
		return err
	}

	if p.bitsPerEntry == 0 {
		// single-value
		return buf.WriteVarInt(ns.VarInt(p.singleValue))
	}

	if p.palette != nil {
		// indirect: write palette
		if err := buf.WriteVarInt(ns.VarInt(len(p.palette))); err != nil {
			return err
		}
		for _, v := range p.palette {
			if err := buf.WriteVarInt(ns.VarInt(v)); err != nil {
				return err
			}
		}
	}

	// write fixed-size long array
	for _, v := range p.data {
		if err := buf.WriteInt64(ns.Int64(v)); err != nil {
			return err
		}
	}
	return nil
}
