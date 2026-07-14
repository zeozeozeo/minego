package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"github.com/zeozeozeo/minego/internal/protocol/nbt"
)

// BlockEntity wire format:
//   Uint8 packedXZ (x<<4 | z)
//   Int16 y
//   VarInt type
//   NBT data (network format, nameless root)

var blockEntityTestCases = []struct {
	name     string
	raw      []byte
	packedXZ ns.Uint8
	y        ns.Int16
	typ      ns.VarInt
}{
	{
		name: "origin sign",
		// packedXZ=0 (x=0,z=0), y=64, type=7 (sign), empty compound
		raw: []byte{
			0x00,       // packedXZ
			0x00, 0x40, // y=64 big-endian
			0x07,       // type=7
			0x0a, 0x00, // compound tag (nameless), end tag
		},
		packedXZ: 0,
		y:        64,
		typ:      7,
	},
	{
		name: "trapped chest",
		// packedXZ=0xFF (x=15,z=15), y=-64, type=2 (trapped chest), empty compound
		raw: []byte{
			0xff,       // packedXZ
			0xff, 0xc0, // y=-64 big-endian
			0x02,       // type=2
			0x0a, 0x00, // compound tag (nameless), end tag
		},
		packedXZ: 0xff,
		y:        -64,
		typ:      2,
	},
}

func TestBlockEntity(t *testing.T) {
	for _, tc := range blockEntityTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var got ns.BlockEntity
			if err := got.Decode(ns.NewReader(tc.raw)); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got.PackedXZ != tc.packedXZ {
				t.Errorf("PackedXZ mismatch: got %d, want %d", got.PackedXZ, tc.packedXZ)
			}
			if got.Y != tc.y {
				t.Errorf("Y mismatch: got %d, want %d", got.Y, tc.y)
			}
			if got.Type != tc.typ {
				t.Errorf("Type mismatch: got %d, want %d", got.Type, tc.typ)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			be := ns.BlockEntity{
				PackedXZ: tc.packedXZ,
				Y:        tc.y,
				Type:     tc.typ,
				Data:     nbt.Compound{},
			}
			buf := ns.NewWriter()
			if err := be.Encode(buf); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

func TestBlockEntity_XZ(t *testing.T) {
	cases := []struct {
		packed ns.Uint8
		x, z   int
	}{
		{0x00, 0, 0},
		{0x10, 1, 0},
		{0x01, 0, 1},
		{0xff, 15, 15},
		{0xa5, 10, 5},
	}

	for _, tc := range cases {
		be := ns.BlockEntity{PackedXZ: tc.packed}
		if be.X() != tc.x || be.Z() != tc.z {
			t.Errorf("packed 0x%02x: got (%d,%d), want (%d,%d)", tc.packed, be.X(), be.Z(), tc.x, tc.z)
		}

		// test SetXZ
		be2 := ns.BlockEntity{}
		be2.SetXZ(tc.x, tc.z)
		if be2.PackedXZ != tc.packed {
			t.Errorf("SetXZ(%d,%d): got 0x%02x, want 0x%02x", tc.x, tc.z, be2.PackedXZ, tc.packed)
		}
	}
}

// LightData wire format:
//   BitSet skyLightMask
//   BitSet blockLightMask
//   BitSet emptySkyLightMask
//   BitSet emptyBlockLightMask
//   VarInt skyArrayCount + (VarInt len + 2048 bytes) × count
//   VarInt blockArrayCount + (VarInt len + 2048 bytes) × count

func TestLightData_RoundTrip(t *testing.T) {
	// create minimal light data using BitSet operations
	skyMask := ns.NewBitSet(64)
	skyMask.Set(1) // section 1 has sky light
	blockMask := ns.NewBitSet(64)
	blockMask.Set(1) // section 1 has block light
	emptyMask := ns.NewBitSet(64)

	ld := ns.LightData{
		SkyLightMask:        *skyMask,
		BlockLightMask:      *blockMask,
		EmptySkyLightMask:   *emptyMask,
		EmptyBlockLightMask: *emptyMask,
		SkyLightArrays:      [][]byte{make([]byte, 2048)},
		BlockLightArrays:    [][]byte{make([]byte, 2048)},
	}

	// set some light values
	ld.SkyLightArrays[0][0] = 0xff   // full sky light at first block
	ld.BlockLightArrays[0][0] = 0x0f // some block light

	// encode
	buf := ns.NewWriter()
	if err := ld.Encode(buf); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	// decode
	var decoded ns.LightData
	if err := decoded.Decode(ns.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	// verify
	if len(decoded.SkyLightArrays) != 1 || len(decoded.BlockLightArrays) != 1 {
		t.Error("array count mismatch")
	}
	if decoded.SkyLightArrays[0][0] != 0xff {
		t.Error("sky light value mismatch")
	}
	if decoded.BlockLightArrays[0][0] != 0x0f {
		t.Error("block light value mismatch")
	}
}

// ChunkData wire format:
//   VarInt heightmap count + (VarInt key + VarInt longCount + Int64[]) entries
//   VarInt dataLen + raw bytes
//   VarInt blockEntityCount + BlockEntity × count

func TestChunkData_RoundTrip(t *testing.T) {
	cd := ns.ChunkData{
		Heightmaps: map[int32][]int64{
			4: make([]int64, 37), // MOTION_BLOCKING
		},
		Data:          []byte{0x01, 0x02, 0x03, 0x04},
		BlockEntities: []ns.BlockEntity{},
	}

	// encode
	buf := ns.NewWriter()
	if err := cd.Encode(buf); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	// decode
	var decoded ns.ChunkData
	if err := decoded.Decode(ns.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	// verify data
	if !bytes.Equal(decoded.Data, cd.Data) {
		t.Errorf("Data mismatch: got %x, want %x", decoded.Data, cd.Data)
	}
	if len(decoded.BlockEntities) != 0 {
		t.Errorf("BlockEntities count mismatch: got %d, want 0", len(decoded.BlockEntities))
	}
	if len(decoded.Heightmaps) != 1 {
		t.Errorf("Heightmaps count mismatch: got %d, want 1", len(decoded.Heightmaps))
	}
	if longs, ok := decoded.Heightmaps[4]; !ok || len(longs) != 37 {
		t.Errorf("Heightmaps MOTION_BLOCKING mismatch")
	}
}

func TestChunkData_WithBlockEntities(t *testing.T) {
	cd := ns.ChunkData{
		Heightmaps: map[int32][]int64{},
		Data:       []byte{},
		BlockEntities: []ns.BlockEntity{
			{PackedXZ: 0x00, Y: 64, Type: 7, Data: nbt.Compound{}},
			{PackedXZ: 0xff, Y: -64, Type: 2, Data: nbt.Compound{}},
		},
	}

	// encode
	buf := ns.NewWriter()
	if err := cd.Encode(buf); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	// decode
	var decoded ns.ChunkData
	if err := decoded.Decode(ns.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	// verify block entities
	if len(decoded.BlockEntities) != 2 {
		t.Fatalf("BlockEntities count: got %d, want 2", len(decoded.BlockEntities))
	}
	if decoded.BlockEntities[0].Y != 64 || decoded.BlockEntities[0].Type != 7 {
		t.Error("first block entity mismatch")
	}
	if decoded.BlockEntities[1].Y != -64 || decoded.BlockEntities[1].Type != 2 {
		t.Error("second block entity mismatch")
	}
}
