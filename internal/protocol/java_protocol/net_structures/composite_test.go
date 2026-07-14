package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// bitSetFromLongs creates a BitSet from raw long values (test helper).
func bitSetFromLongs(longs []int64) *ns.BitSet {
	bs := ns.NewBitSet(len(longs) * 64)
	for i, v := range longs {
		for bit := range 64 {
			if (v & (1 << bit)) != 0 {
				bs.Set(i*64 + bit)
			}
		}
	}
	return bs
}

// BitSet wire format:
//   VarInt length (number of longs)
//   Int64 × length (big-endian)

var bitSetTestCases = []struct {
	name     string
	raw      []byte
	expected []int64
}{
	{
		name:     "empty",
		raw:      []byte{0x00},
		expected: []int64{},
	},
	{
		name: "single long with bit 0",
		// length=1, long=1 (bit 0 set)
		raw:      []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		expected: []int64{1},
	},
	{
		name: "single long with bit 63",
		// length=1, long=0x8000000000000000 (bit 63 set)
		raw:      []byte{0x01, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		expected: []int64{-9223372036854775808}, // signed representation
	},
	{
		name: "two longs",
		// length=2, long1=3 (bits 0,1), long2=5 (bits 64,66)
		raw: []byte{
			0x02,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05,
		},
		expected: []int64{3, 5},
	},
}

func TestBitSet(t *testing.T) {
	for _, tc := range bitSetTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var got ns.BitSet
			if err := got.Decode(ns.NewReader(tc.raw)); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			longs := got.Longs()
			if len(longs) != len(tc.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(longs), len(tc.expected))
			}
			for i, v := range tc.expected {
				if longs[i] != v {
					t.Errorf("long[%d] mismatch: got %d, want %d", i, longs[i], v)
				}
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			bs := bitSetFromLongs(tc.expected)
			buf := ns.NewWriter()
			if err := bs.Encode(buf); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

func TestBitSet_GetSet(t *testing.T) {
	bs := ns.NewBitSet(128)

	// initially all bits should be unset
	for i := range 128 {
		if bs.Get(i) {
			t.Errorf("bit %d should be unset", i)
		}
	}

	// set some bits
	bs.Set(0)
	bs.Set(63)
	bs.Set(64)
	bs.Set(127)

	if !bs.Get(0) || !bs.Get(63) || !bs.Get(64) || !bs.Get(127) {
		t.Error("set bits should be set")
	}
	if bs.Get(1) || bs.Get(62) || bs.Get(65) {
		t.Error("unset bits should remain unset")
	}

	// clear a bit
	bs.Clear(63)
	if bs.Get(63) {
		t.Error("cleared bit should be unset")
	}
}

// FixedBitSet wire format:
//   ceil(n/8) bytes (no length prefix)

var fixedBitSetTestCases = []struct {
	name    string
	size    int
	raw     []byte
	setBits []int
}{
	{
		name:    "8 bits, none set",
		size:    8,
		raw:     []byte{0x00},
		setBits: []int{},
	},
	{
		name:    "8 bits, bit 0 set",
		size:    8,
		raw:     []byte{0x01},
		setBits: []int{0},
	},
	{
		name:    "8 bits, bits 0,7 set",
		size:    8,
		raw:     []byte{0x81},
		setBits: []int{0, 7},
	},
	{
		name:    "16 bits, bits 0,8 set",
		size:    16,
		raw:     []byte{0x01, 0x01},
		setBits: []int{0, 8},
	},
}

func TestFixedBitSet(t *testing.T) {
	for _, tc := range fixedBitSetTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			fbs := ns.NewFixedBitSet(tc.size)
			if err := fbs.Decode(ns.NewReader(tc.raw)); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			for _, bit := range tc.setBits {
				if !fbs.Get(bit) {
					t.Errorf("bit %d should be set", bit)
				}
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			fbs := ns.NewFixedBitSet(tc.size)
			for _, bit := range tc.setBits {
				fbs.Set(bit)
			}
			buf := ns.NewWriter()
			if err := fbs.Encode(buf); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

// IDSet wire format:
//   VarInt type (0 = tag, >0 = inline count + 1)
//   if type=0: Identifier (tag name)
//   if type>0: VarInt × (type-1) IDs

var idSetTestCases = []struct {
	name    string
	raw     []byte
	isTag   bool
	tagName ns.Identifier
	ids     []ns.VarInt
}{
	{
		name: "tag reference",
		// type=0, identifier="minecraft:test" (length 14)
		raw:     []byte{0x00, 0x0e, 'm', 'i', 'n', 'e', 'c', 'r', 'a', 'f', 't', ':', 't', 'e', 's', 't'},
		isTag:   true,
		tagName: "minecraft:test",
	},
	{
		name:  "empty inline",
		raw:   []byte{0x01},
		isTag: false,
		ids:   []ns.VarInt{},
	},
	{
		name:  "single inline ID",
		raw:   []byte{0x02, 0x2a},
		isTag: false,
		ids:   []ns.VarInt{42},
	},
	{
		name:  "multiple inline IDs",
		raw:   []byte{0x04, 0x01, 0x02, 0x03},
		isTag: false,
		ids:   []ns.VarInt{1, 2, 3},
	},
}

func TestIDSet(t *testing.T) {
	for _, tc := range idSetTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var got ns.IDSet
			if err := got.Decode(ns.NewReader(tc.raw)); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got.IsTag != tc.isTag {
				t.Errorf("IsTag mismatch: got %v, want %v", got.IsTag, tc.isTag)
			}
			if got.IsTag {
				if got.TagName != tc.tagName {
					t.Errorf("TagName mismatch: got %q, want %q", got.TagName, tc.tagName)
				}
			} else {
				if len(got.IDs) != len(tc.ids) {
					t.Fatalf("IDs length mismatch: got %d, want %d", len(got.IDs), len(tc.ids))
				}
				for i, id := range tc.ids {
					if got.IDs[i] != id {
						t.Errorf("ID[%d] mismatch: got %d, want %d", i, got.IDs[i], id)
					}
				}
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			var idset *ns.IDSet
			if tc.isTag {
				idset = ns.NewTagIDSet(tc.tagName)
			} else {
				idset = ns.NewInlineIDSet(tc.ids)
			}
			buf := ns.NewWriter()
			if err := idset.Encode(buf); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

// PrefixedArray wire format:
//   VarInt length
//   T × length

func TestPrefixedArray(t *testing.T) {
	testCases := []struct {
		name     string
		raw      []byte
		expected []ns.VarInt
	}{
		{
			name:     "empty",
			raw:      []byte{0x00},
			expected: []ns.VarInt{},
		},
		{
			name:     "single element",
			raw:      []byte{0x01, 0x2a},
			expected: []ns.VarInt{42},
		},
		{
			name:     "multiple elements",
			raw:      []byte{0x03, 0x01, 0x02, 0x03},
			expected: []ns.VarInt{1, 2, 3},
		},
	}

	decoder := func(buf *ns.PacketBuffer) (ns.VarInt, error) { return buf.ReadVarInt() }
	encoder := func(buf *ns.PacketBuffer, v ns.VarInt) error { return buf.WriteVarInt(v) }

	for _, tc := range testCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var arr ns.PrefixedArray[ns.VarInt]
			if err := arr.DecodeWith(ns.NewReader(tc.raw), decoder); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if len(arr) != len(tc.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(arr), len(tc.expected))
			}
			for i, v := range tc.expected {
				if arr[i] != v {
					t.Errorf("element[%d] mismatch: got %d, want %d", i, arr[i], v)
				}
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			arr := ns.PrefixedArray[ns.VarInt](tc.expected)
			buf := ns.NewWriter()
			if err := arr.EncodeWith(buf, encoder); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

// PrefixedOptional wire format:
//   Boolean present
//   T value (if present)

func TestPrefixedOptional(t *testing.T) {
	testCases := []struct {
		name     string
		raw      []byte
		expected ns.PrefixedOptional[ns.VarInt]
	}{
		{
			name:     "absent",
			raw:      []byte{0x00},
			expected: ns.None[ns.VarInt](),
		},
		{
			name:     "present",
			raw:      []byte{0x01, 0x2a},
			expected: ns.Some[ns.VarInt](42),
		},
	}

	decoder := func(buf *ns.PacketBuffer) (ns.VarInt, error) { return buf.ReadVarInt() }
	encoder := func(buf *ns.PacketBuffer, v ns.VarInt) error { return buf.WriteVarInt(v) }

	for _, tc := range testCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var opt ns.PrefixedOptional[ns.VarInt]
			if err := opt.DecodeWith(ns.NewReader(tc.raw), decoder); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if opt.Present != tc.expected.Present {
				t.Errorf("Present mismatch: got %v, want %v", opt.Present, tc.expected.Present)
			}
			if opt.Present && opt.Value != tc.expected.Value {
				t.Errorf("Value mismatch: got %d, want %d", opt.Value, tc.expected.Value)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := tc.expected.EncodeWith(buf, encoder); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

// XOrY wire format:
//   Boolean isX
//   X or Y value

func TestXOrY(t *testing.T) {
	testCases := []struct {
		name     string
		raw      []byte
		expected ns.XOrY[ns.VarInt, ns.String]
		isX      bool
		xVal     ns.VarInt
		yVal     ns.String
	}{
		{
			name:     "X value",
			raw:      []byte{0x01, 0x2a},
			expected: ns.NewX[ns.VarInt, ns.String](42),
			isX:      true,
			xVal:     42,
		},
		{
			name:     "Y value",
			raw:      []byte{0x00, 0x05, 'h', 'e', 'l', 'l', 'o'},
			expected: ns.NewY[ns.VarInt, ns.String]("hello"),
			isX:      false,
			yVal:     "hello",
		},
	}

	decodeX := func(buf *ns.PacketBuffer) (ns.VarInt, error) { return buf.ReadVarInt() }
	decodeY := func(buf *ns.PacketBuffer) (ns.String, error) { return buf.ReadString(32767) }
	encodeX := func(buf *ns.PacketBuffer, v ns.VarInt) error { return buf.WriteVarInt(v) }
	encodeY := func(buf *ns.PacketBuffer, v ns.String) error { return buf.WriteString(v) }

	for _, tc := range testCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var got ns.XOrY[ns.VarInt, ns.String]
			if err := got.DecodeWith(ns.NewReader(tc.raw), decodeX, decodeY); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got.IsX != tc.isX {
				t.Errorf("IsX mismatch: got %v, want %v", got.IsX, tc.isX)
			}
			if tc.isX && got.X != tc.xVal {
				t.Errorf("X mismatch: got %d, want %d", got.X, tc.xVal)
			}
			if !tc.isX && got.Y != tc.yVal {
				t.Errorf("Y mismatch: got %q, want %q", got.Y, tc.yVal)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := tc.expected.EncodeWith(buf, encodeX, encodeY); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

// IDOrX wire format:
//   VarInt id (0 = inline, >0 = registry id + 1)
//   T value (if id = 0)

func TestIDOrX(t *testing.T) {
	testCases := []struct {
		name     string
		raw      []byte
		expected ns.IDOrX[ns.VarInt]
	}{
		{
			name:     "registry reference",
			raw:      []byte{0x2b},
			expected: ns.NewIDRef[ns.VarInt](42),
		},
		{
			name:     "inline value",
			raw:      []byte{0x00, 0x64},
			expected: ns.NewInlineValue[ns.VarInt](100),
		},
	}

	decoder := func(buf *ns.PacketBuffer) (ns.VarInt, error) { return buf.ReadVarInt() }
	encoder := func(buf *ns.PacketBuffer, v ns.VarInt) error { return buf.WriteVarInt(v) }

	for _, tc := range testCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var got ns.IDOrX[ns.VarInt]
			if err := got.DecodeWith(ns.NewReader(tc.raw), decoder); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got.IsInline != tc.expected.IsInline {
				t.Errorf("IsInline mismatch: got %v, want %v", got.IsInline, tc.expected.IsInline)
			}
			if got.IsInline {
				if got.Value != tc.expected.Value {
					t.Errorf("Value mismatch: got %d, want %d", got.Value, tc.expected.Value)
				}
			} else {
				if got.ID != tc.expected.ID {
					t.Errorf("ID mismatch: got %d, want %d", got.ID, tc.expected.ID)
				}
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := tc.expected.EncodeWith(buf, encoder); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}
