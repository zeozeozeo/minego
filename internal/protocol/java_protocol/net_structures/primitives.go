package net_structures

import (
	"encoding/binary"
	"io"
	"math"
)

// Primitive type definitions for the Minecraft protocol.
// All multi-byte integers are big-endian unless otherwise specified.

// Boolean is a single byte (0x00 = false, 0x01 = true).
type Boolean bool

// Encode writes the Boolean to w.
func (v Boolean) Encode(w io.Writer) error {
	var b byte
	if v {
		b = 0x01
	}
	_, err := w.Write([]byte{b})
	return err
}

// DecodeBoolean reads a Boolean from r.
func DecodeBoolean(r io.Reader) (Boolean, error) {
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return false, err
	}
	return b[0] != 0, nil
}

// Int8 is a signed 8-bit integer (-128 to 127).
type Int8 int8

// Encode writes the Int8 to w.
func (v Int8) Encode(w io.Writer) error {
	_, err := w.Write([]byte{byte(v)})
	return err
}

// DecodeInt8 reads an Int8 from r.
func DecodeInt8(r io.Reader) (Int8, error) {
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return Int8(b[0]), nil
}

// Uint8 is an unsigned 8-bit integer (0 to 255).
type Uint8 uint8

// Encode writes the Uint8 to w.
func (v Uint8) Encode(w io.Writer) error {
	_, err := w.Write([]byte{byte(v)})
	return err
}

// DecodeUint8 reads a Uint8 from r.
func DecodeUint8(r io.Reader) (Uint8, error) {
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return Uint8(b[0]), nil
}

// Int16 is a big-endian signed 16-bit integer.
type Int16 int16

// Encode writes the Int16 to w.
func (v Int16) Encode(w io.Writer) error {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], uint16(v))
	_, err := w.Write(b[:])
	return err
}

// DecodeInt16 reads an Int16 from r.
func DecodeInt16(r io.Reader) (Int16, error) {
	var b [2]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return Int16(binary.BigEndian.Uint16(b[:])), nil
}

// Uint16 is a big-endian unsigned 16-bit integer.
type Uint16 uint16

// Encode writes the Uint16 to w.
func (v Uint16) Encode(w io.Writer) error {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], uint16(v))
	_, err := w.Write(b[:])
	return err
}

// DecodeUint16 reads a Uint16 from r.
func DecodeUint16(r io.Reader) (Uint16, error) {
	var b [2]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return Uint16(binary.BigEndian.Uint16(b[:])), nil
}

// Int32 is a big-endian signed 32-bit integer.
type Int32 int32

// Encode writes the Int32 to w.
func (v Int32) Encode(w io.Writer) error {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(v))
	_, err := w.Write(b[:])
	return err
}

// DecodeInt32 reads an Int32 from r.
func DecodeInt32(r io.Reader) (Int32, error) {
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return Int32(binary.BigEndian.Uint32(b[:])), nil
}

// Int64 is a big-endian signed 64-bit integer.
type Int64 int64

// Encode writes the Int64 to w.
func (v Int64) Encode(w io.Writer) error {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(v))
	_, err := w.Write(b[:])
	return err
}

// DecodeInt64 reads an Int64 from r.
func DecodeInt64(r io.Reader) (Int64, error) {
	var b [8]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return Int64(binary.BigEndian.Uint64(b[:])), nil
}

// Float32 is a big-endian IEEE 754 single-precision float.
type Float32 float32

// Encode writes the Float32 to w.
func (v Float32) Encode(w io.Writer) error {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], math.Float32bits(float32(v)))
	_, err := w.Write(b[:])
	return err
}

// DecodeFloat32 reads a Float32 from r.
func DecodeFloat32(r io.Reader) (Float32, error) {
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return Float32(math.Float32frombits(binary.BigEndian.Uint32(b[:]))), nil
}

// Float64 is a big-endian IEEE 754 double-precision float.
type Float64 float64

// Encode writes the Float64 to w.
func (v Float64) Encode(w io.Writer) error {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], math.Float64bits(float64(v)))
	_, err := w.Write(b[:])
	return err
}

// DecodeFloat64 reads a Float64 from r.
func DecodeFloat64(r io.Reader) (Float64, error) {
	var b [8]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return Float64(math.Float64frombits(binary.BigEndian.Uint64(b[:]))), nil
}
