package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// Position wire format (64-bit integer, big-endian):
//   bits 0-11:  Y (signed 12-bit)
//   bits 12-37: Z (signed 26-bit)
//   bits 38-63: X (signed 26-bit)
//
// Reference: https://minecraft.wiki/w/Java_Edition_protocol/Data_types#Position

var positionTestCases = []struct {
	name     string
	raw      []byte
	expected ns.Position
}{
	{
		name:     "origin",
		raw:      []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		expected: ns.Position{X: 0, Y: 0, Z: 0},
	},
	{
		name: "y=64 spawn height",
		// Y=64 in bits 0-11 = 0x40
		raw:      []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40},
		expected: ns.Position{X: 0, Y: 64, Z: 0},
	},
	{
		name: "wiki example",
		// position (18357644, 831, -20882616)
		raw:      []byte{0x46, 0x07, 0x63, 0x2c, 0x15, 0xb4, 0x83, 0x3f},
		expected: ns.Position{X: 18357644, Y: 831, Z: -20882616},
	},
	{
		name: "negative coordinates",
		// X=-1, Y=-1, Z=-1 = 0xFFFFFFFFFFFFFFFF
		raw:      []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		expected: ns.Position{X: -1, Y: -1, Z: -1},
	},
	{
		name: "max positive",
		// X=33554431 (max 26-bit signed), Y=2047 (max 12-bit signed), Z=33554431
		raw:      []byte{0x7f, 0xff, 0xff, 0xdf, 0xff, 0xff, 0xf7, 0xff},
		expected: ns.Position{X: 33554431, Y: 2047, Z: 33554431},
	},
}

func TestPosition(t *testing.T) {
	for _, tc := range positionTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			buf := ns.NewReader(tc.raw)
			got, err := buf.ReadPosition()
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got != tc.expected {
				t.Errorf("decode mismatch:\n  got:  %+v\n  want: %+v", got, tc.expected)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WritePosition(tc.expected); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

func TestPosition_PackUnpack(t *testing.T) {
	// additional round-trip tests for edge cases
	positions := []ns.Position{
		{X: 100, Y: 64, Z: -200},
		{X: -100, Y: 64, Z: 200},
		{X: -33554432, Y: -2048, Z: -33554432}, // min values
	}

	for _, pos := range positions {
		packed := pos.Pack()
		decoded := ns.UnpackPosition(packed)
		if decoded != pos {
			t.Errorf("round-trip failed: %+v -> 0x%x -> %+v", pos, packed, decoded)
		}
	}
}
