package net_structures_test

import (
	"bytes"
	"math"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

func TestBooleanReadWrite(t *testing.T) {
	tests := []struct {
		name     string
		value    ns.Boolean
		expected []byte
	}{
		{"false", false, []byte{0x00}},
		{"true", true, []byte{0x01}},
	}

	for _, tt := range tests {
		t.Run(tt.name+" write", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteBool(tt.value); err != nil {
				t.Fatalf("WriteBool() error = %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("WriteBool() = %v, want %v", buf.Bytes(), tt.expected)
			}
		})

		t.Run(tt.name+" read", func(t *testing.T) {
			buf := ns.NewReader(tt.expected)
			got, err := buf.ReadBool()
			if err != nil {
				t.Fatalf("ReadBool() error = %v", err)
			}
			if got != tt.value {
				t.Errorf("ReadBool() = %v, want %v", got, tt.value)
			}
		})
	}

	// Non-zero values are truthy
	t.Run("non-zero is true", func(t *testing.T) {
		buf := ns.NewReader([]byte{0x42})
		got, err := buf.ReadBool()
		if err != nil {
			t.Fatalf("ReadBool() error = %v", err)
		}
		if got != true {
			t.Errorf("ReadBool(0x42) = %v, want true", got)
		}
	})
}

func TestInt8ReadWrite(t *testing.T) {
	tests := []struct {
		value    ns.Int8
		expected []byte
	}{
		{0, []byte{0x00}},
		{1, []byte{0x01}},
		{127, []byte{0x7f}},
		{-1, []byte{0xff}},
		{-128, []byte{0x80}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteInt8(tt.value); err != nil {
				t.Fatalf("WriteInt8() error = %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("WriteInt8(%d) = %v, want %v", tt.value, buf.Bytes(), tt.expected)
			}

			reader := ns.NewReader(tt.expected)
			got, err := reader.ReadInt8()
			if err != nil {
				t.Fatalf("ReadInt8() error = %v", err)
			}
			if got != tt.value {
				t.Errorf("ReadInt8() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestUint8ReadWrite(t *testing.T) {
	tests := []struct {
		value    ns.Uint8
		expected []byte
	}{
		{0, []byte{0x00}},
		{1, []byte{0x01}},
		{127, []byte{0x7f}},
		{128, []byte{0x80}},
		{255, []byte{0xff}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteUint8(tt.value); err != nil {
				t.Fatalf("WriteUint8() error = %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("WriteUint8(%d) = %v, want %v", tt.value, buf.Bytes(), tt.expected)
			}

			reader := ns.NewReader(tt.expected)
			got, err := reader.ReadUint8()
			if err != nil {
				t.Fatalf("ReadUint8() error = %v", err)
			}
			if got != tt.value {
				t.Errorf("ReadUint8() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestInt16ReadWrite(t *testing.T) {
	// Big-endian encoding
	tests := []struct {
		value    ns.Int16
		expected []byte
	}{
		{0, []byte{0x00, 0x00}},
		{1, []byte{0x00, 0x01}},
		{256, []byte{0x01, 0x00}},
		{32767, []byte{0x7f, 0xff}},
		{-1, []byte{0xff, 0xff}},
		{-32768, []byte{0x80, 0x00}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteInt16(tt.value); err != nil {
				t.Fatalf("WriteInt16() error = %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("WriteInt16(%d) = %v, want %v", tt.value, buf.Bytes(), tt.expected)
			}

			reader := ns.NewReader(tt.expected)
			got, err := reader.ReadInt16()
			if err != nil {
				t.Fatalf("ReadInt16() error = %v", err)
			}
			if got != tt.value {
				t.Errorf("ReadInt16() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestUint16ReadWrite(t *testing.T) {
	// Big-endian encoding - common for port numbers
	tests := []struct {
		name     string
		value    ns.Uint16
		expected []byte
	}{
		{"zero", 0, []byte{0x00, 0x00}},
		{"one", 1, []byte{0x00, 0x01}},
		{"256", 256, []byte{0x01, 0x00}},
		{"25565 (MC port)", 25565, []byte{0x63, 0xdd}},
		{"max", 65535, []byte{0xff, 0xff}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteUint16(tt.value); err != nil {
				t.Fatalf("WriteUint16() error = %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("WriteUint16(%d) = %v, want %v", tt.value, buf.Bytes(), tt.expected)
			}

			reader := ns.NewReader(tt.expected)
			got, err := reader.ReadUint16()
			if err != nil {
				t.Fatalf("ReadUint16() error = %v", err)
			}
			if got != tt.value {
				t.Errorf("ReadUint16() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestInt32ReadWrite(t *testing.T) {
	tests := []struct {
		value    ns.Int32
		expected []byte
	}{
		{0, []byte{0x00, 0x00, 0x00, 0x00}},
		{1, []byte{0x00, 0x00, 0x00, 0x01}},
		{256, []byte{0x00, 0x00, 0x01, 0x00}},
		{2147483647, []byte{0x7f, 0xff, 0xff, 0xff}},
		{-1, []byte{0xff, 0xff, 0xff, 0xff}},
		{-2147483648, []byte{0x80, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteInt32(tt.value); err != nil {
				t.Fatalf("WriteInt32() error = %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("WriteInt32(%d) = %v, want %v", tt.value, buf.Bytes(), tt.expected)
			}

			reader := ns.NewReader(tt.expected)
			got, err := reader.ReadInt32()
			if err != nil {
				t.Fatalf("ReadInt32() error = %v", err)
			}
			if got != tt.value {
				t.Errorf("ReadInt32() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestInt64ReadWrite(t *testing.T) {
	tests := []struct {
		value    ns.Int64
		expected []byte
	}{
		{0, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{1, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{9223372036854775807, []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{-1, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{-9223372036854775808, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteInt64(tt.value); err != nil {
				t.Fatalf("WriteInt64() error = %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("WriteInt64(%d) = %v, want %v", tt.value, buf.Bytes(), tt.expected)
			}

			reader := ns.NewReader(tt.expected)
			got, err := reader.ReadInt64()
			if err != nil {
				t.Fatalf("ReadInt64() error = %v", err)
			}
			if got != tt.value {
				t.Errorf("ReadInt64() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestFloat32ReadWrite(t *testing.T) {
	tests := []struct {
		name     string
		value    ns.Float32
		expected []byte
	}{
		{"zero", 0.0, []byte{0x00, 0x00, 0x00, 0x00}},
		{"one", 1.0, []byte{0x3f, 0x80, 0x00, 0x00}},
		{"negative one", -1.0, []byte{0xbf, 0x80, 0x00, 0x00}},
		{"pi approx", 3.14159, []byte{0x40, 0x49, 0x0f, 0xd0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteFloat32(tt.value); err != nil {
				t.Fatalf("WriteFloat32() error = %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("WriteFloat32(%v) = %v, want %v", tt.value, buf.Bytes(), tt.expected)
			}

			reader := ns.NewReader(tt.expected)
			got, err := reader.ReadFloat32()
			if err != nil {
				t.Fatalf("ReadFloat32() error = %v", err)
			}
			if math.Abs(float64(got-tt.value)) > 0.0001 {
				t.Errorf("ReadFloat32() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestFloat64ReadWrite(t *testing.T) {
	tests := []struct {
		name     string
		value    ns.Float64
		expected []byte
	}{
		{"zero", 0.0, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"one", 1.0, []byte{0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"negative one", -1.0, []byte{0xbf, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"pi", 3.141592653589793, []byte{0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2d, 0x18}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteFloat64(tt.value); err != nil {
				t.Fatalf("WriteFloat64() error = %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("WriteFloat64(%v) = %v, want %v", tt.value, buf.Bytes(), tt.expected)
			}

			reader := ns.NewReader(tt.expected)
			got, err := reader.ReadFloat64()
			if err != nil {
				t.Fatalf("ReadFloat64() error = %v", err)
			}
			if got != tt.value {
				t.Errorf("ReadFloat64() = %v, want %v", got, tt.value)
			}
		})
	}
}
