package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// VarInt/VarLong wire format: 7 bits per byte, MSB = continuation flag
// Reference: https://minecraft.wiki/w/Java_Edition_protocol/Data_types#VarInt_and_VarLong

var varIntTestCases = []struct {
	name  string
	raw   []byte
	value ns.VarInt
}{
	{"zero", []byte{0x00}, 0},
	{"one", []byte{0x01}, 1},
	{"max 1-byte", []byte{0x7f}, 127},
	{"min 2-byte", []byte{0x80, 0x01}, 128},
	{"255", []byte{0xff, 0x01}, 255},
	{"25565 (MC port)", []byte{0xdd, 0xc7, 0x01}, 25565},
	{"max 3-byte", []byte{0xff, 0xff, 0x7f}, 2097151},
	{"max int32", []byte{0xff, 0xff, 0xff, 0xff, 0x07}, 2147483647},
	{"-1", []byte{0xff, 0xff, 0xff, 0xff, 0x0f}, -1},
	{"-2", []byte{0xfe, 0xff, 0xff, 0xff, 0x0f}, -2},
	{"min int32", []byte{0x80, 0x80, 0x80, 0x80, 0x08}, -2147483648},
}

var varLongTestCases = []struct {
	name  string
	raw   []byte
	value ns.VarLong
}{
	{"zero", []byte{0x00}, 0},
	{"one", []byte{0x01}, 1},
	{"max 1-byte", []byte{0x7f}, 127},
	{"128", []byte{0x80, 0x01}, 128},
	{"max int64", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}, 9223372036854775807},
	{"-1", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}, -1},
}

func TestVarInt(t *testing.T) {
	for _, tc := range varIntTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			got, err := ns.NewReader(tc.raw).ReadVarInt()
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got != tc.value {
				t.Errorf("got %d, want %d", got, tc.value)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			got, err := tc.value.ToBytes()
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(got, tc.raw) {
				t.Errorf("got %x, want %x", got, tc.raw)
			}
		})
	}
}

func TestVarLong(t *testing.T) {
	for _, tc := range varLongTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			got, err := ns.NewReader(tc.raw).ReadVarLong()
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got != tc.value {
				t.Errorf("got %d, want %d", got, tc.value)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			got, err := tc.value.ToBytes()
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(got, tc.raw) {
				t.Errorf("got %x, want %x", got, tc.raw)
			}
		})
	}
}

func TestVarInt_TooLong(t *testing.T) {
	// 6 continuation bytes - invalid
	_, err := ns.NewReader([]byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80}).ReadVarInt()
	if err == nil {
		t.Error("should error on too many bytes")
	}
}

func TestVarInt_Len(t *testing.T) {
	cases := []struct {
		value ns.VarInt
		len   int
	}{
		{0, 1}, {127, 1}, {128, 2}, {16383, 2}, {16384, 3},
		{2097151, 3}, {2097152, 4}, {268435455, 4}, {268435456, 5}, {-1, 5},
	}
	for _, tc := range cases {
		if got := tc.value.Len(); got != tc.len {
			t.Errorf("VarInt(%d).Len() = %d, want %d", tc.value, got, tc.len)
		}
	}
}
