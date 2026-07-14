package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// UUID wire format: 16 bytes (128-bit, big-endian)

var uuidTestCases = []struct {
	name  string
	raw   []byte
	str   string
	value ns.UUID
	msb   int64
	lsb   int64
}{
	{
		name:  "nil",
		raw:   []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		str:   "00000000-0000-0000-0000-000000000000",
		value: ns.NilUUID,
		msb:   0,
		lsb:   0,
	},
	{
		name:  "standard",
		raw:   []byte{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00},
		str:   "550e8400-e29b-41d4-a716-446655440000",
		value: ns.UUID{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00},
		msb:   0x550e8400e29b41d4,
		lsb:   -0x58e9bb99aabbffff - 1, // 0xa716446655440000 as signed
	},
	{
		name:  "all ones",
		raw:   []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		str:   "ffffffff-ffff-ffff-ffff-ffffffffffff",
		value: ns.UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		msb:   -1,
		lsb:   -1,
	},
}

func TestUUID(t *testing.T) {
	for _, tc := range uuidTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			got, err := ns.NewReader(tc.raw).ReadUUID()
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got != tc.value {
				t.Errorf("got %v, want %v", got, tc.value)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteUUID(tc.value); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("got %x, want %x", buf.Bytes(), tc.raw)
			}
		})

		t.Run(tc.name+" from string", func(t *testing.T) {
			got, err := ns.UUIDFromString(tc.str)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if got != tc.value {
				t.Errorf("got %v, want %v", got, tc.value)
			}
		})

		t.Run(tc.name+" to string", func(t *testing.T) {
			if got := tc.value.String(); got != tc.str {
				t.Errorf("got %q, want %q", got, tc.str)
			}
		})

		t.Run(tc.name+" int64s", func(t *testing.T) {
			if got := tc.value.MostSignificantBits(); got != tc.msb {
				t.Errorf("MSB got %d, want %d", got, tc.msb)
			}
			if got := tc.value.LeastSignificantBits(); got != tc.lsb {
				t.Errorf("LSB got %d, want %d", got, tc.lsb)
			}
			if got := ns.UUIDFromInt64s(tc.msb, tc.lsb); got != tc.value {
				t.Errorf("FromInt64s got %v, want %v", got, tc.value)
			}
		})
	}
}

func TestUUID_ParseErrors(t *testing.T) {
	invalid := []string{
		"550e8400",                              // too short
		"550e8400-e29b-41d4-a716-44665544000g",  // invalid hex
		"550e8400-e29b-41d4-a716-4466554400000", // too long
	}
	for _, s := range invalid {
		if _, err := ns.UUIDFromString(s); err == nil {
			t.Errorf("UUIDFromString(%q) should error", s)
		}
	}
}

func TestUUID_IsNil(t *testing.T) {
	if !ns.NilUUID.IsNil() {
		t.Error("NilUUID.IsNil() should be true")
	}
	nonNil := ns.UUID{0x01}
	if nonNil.IsNil() {
		t.Error("non-nil UUID.IsNil() should be false")
	}
}
