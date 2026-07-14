package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// String wire format: VarInt byte-length + UTF-8 bytes
// Identifier: same as String, format namespace:path

var stringTestCases = []struct {
	name  string
	raw   []byte
	value ns.String
}{
	{"empty", []byte{0x00}, ""},
	{"hello", []byte{0x05, 'h', 'e', 'l', 'l', 'o'}, "hello"},
	{"minecraft", []byte{0x09, 'm', 'i', 'n', 'e', 'c', 'r', 'a', 'f', 't'}, "minecraft"},
	{"UTF-8 café", []byte{0x05, 'c', 'a', 'f', 0xc3, 0xa9}, "café"},
	{"UTF-8 日本", []byte{0x06, 0xe6, 0x97, 0xa5, 0xe6, 0x9c, 0xac}, "日本"},
}

var identifierTestCases = []struct {
	name  string
	raw   []byte
	value ns.Identifier
}{
	{"stone", []byte{0x0f, 'm', 'i', 'n', 'e', 'c', 'r', 'a', 'f', 't', ':', 's', 't', 'o', 'n', 'e'}, "minecraft:stone"},
	{"custom", []byte{0x0b, 'c', 'u', 's', 't', 'o', 'm', ':', 'i', 't', 'e', 'm'}, "custom:item"},
}

func TestString(t *testing.T) {
	for _, tc := range stringTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			got, err := ns.NewReader(tc.raw).ReadString(0)
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got != tc.value {
				t.Errorf("got %q, want %q", got, tc.value)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteString(tc.value); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("got %x, want %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

func TestIdentifier(t *testing.T) {
	for _, tc := range identifierTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			got, err := ns.NewReader(tc.raw).ReadIdentifier()
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got != tc.value {
				t.Errorf("got %q, want %q", got, tc.value)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteIdentifier(tc.value); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("got %x, want %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

func TestString_MaxLength(t *testing.T) {
	data := []byte{0x05, 'h', 'e', 'l', 'l', 'o'}
	_, err := ns.NewReader(data).ReadString(3)
	if err == nil {
		t.Error("should error when exceeding max length")
	}
}

func TestIdentifier_NamespacePath(t *testing.T) {
	cases := []struct {
		id        ns.Identifier
		namespace string
		path      string
	}{
		{"minecraft:stone", "minecraft", "stone"},
		{"custom:my_item", "custom", "my_item"},
		{"stone", "minecraft", "stone"}, // default namespace
	}
	for _, tc := range cases {
		if got := tc.id.Namespace(); got != tc.namespace {
			t.Errorf("%q.Namespace() = %q, want %q", tc.id, got, tc.namespace)
		}
		if got := tc.id.Path(); got != tc.path {
			t.Errorf("%q.Path() = %q, want %q", tc.id, got, tc.path)
		}
	}
}
