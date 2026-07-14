package net_structures

import (
	"fmt"
	"io"
)

// VarInt is a variable-length signed 32-bit integer.
//
// Uses 7 bits per byte with bit 7 as continuation flag.
// Values are encoded in little-endian order within the variable-length format.
// Maximum 5 bytes for 32-bit values.
//
// Encoding:
//
//	For each byte:
//	  - Bits 0-6: 7 bits of data
//	  - Bit 7: 1 if more bytes follow, 0 if this is the last byte
//
// Examples:
//
//	0          -> [0x00]
//	1          -> [0x01]
//	127        -> [0x7f]
//	128        -> [0x80, 0x01]
//	255        -> [0xff, 0x01]
//	2147483647 -> [0xff, 0xff, 0xff, 0xff, 0x07]
//	-1         -> [0xff, 0xff, 0xff, 0xff, 0x0f]
type VarInt int32

// Encode writes the VarInt to w.
func (v VarInt) Encode(w io.Writer) error {
	var buf [5]byte
	n := 0
	value := uint32(v)

	for {
		if (value & ^uint32(0x7F)) == 0 {
			buf[n] = byte(value)
			n++
			break
		}
		buf[n] = byte((value & 0x7F) | 0x80)
		n++
		value >>= 7
	}

	_, err := w.Write(buf[:n])
	return err
}

// ToBytes encodes the VarInt to bytes.
func (v VarInt) ToBytes() (ByteArray, error) {
	var buf [5]byte
	n := 0
	value := uint32(v)

	for {
		if (value & ^uint32(0x7F)) == 0 {
			buf[n] = byte(value)
			n++
			break
		}
		buf[n] = byte((value & 0x7F) | 0x80)
		n++
		value >>= 7
	}

	return buf[:n], nil
}

// Len returns the number of bytes needed to encode this VarInt.
func (v VarInt) Len() int {
	value := uint32(v)
	switch {
	case value < 1<<7:
		return 1
	case value < 1<<14:
		return 2
	case value < 1<<21:
		return 3
	case value < 1<<28:
		return 4
	default:
		return 5
	}
}

// DecodeVarInt reads a VarInt from r.
func DecodeVarInt(r io.Reader) (VarInt, error) {
	var value int32
	var position uint
	var b [1]byte

	for {
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, err
		}

		value |= int32(b[0]&0x7F) << position

		if (b[0] & 0x80) == 0 {
			break
		}

		position += 7
		if position >= 35 {
			return 0, fmt.Errorf("VarInt is too big")
		}
	}

	return VarInt(value), nil
}

// VarLong is a variable-length signed 64-bit integer.
//
// Same encoding as VarInt but for 64-bit values.
// Maximum 10 bytes for 64-bit values.
//
// Examples:
//
//	0                    -> [0x00]
//	9223372036854775807  -> [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f]
//	-1                   -> [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01]
type VarLong int64

// Encode writes the VarLong to w.
func (v VarLong) Encode(w io.Writer) error {
	var buf [10]byte
	n := 0
	value := uint64(v)

	for {
		if (value & ^uint64(0x7F)) == 0 {
			buf[n] = byte(value)
			n++
			break
		}
		buf[n] = byte((value & 0x7F) | 0x80)
		n++
		value >>= 7
	}

	_, err := w.Write(buf[:n])
	return err
}

// ToBytes encodes the VarLong to bytes.
func (v VarLong) ToBytes() (ByteArray, error) {
	var buf [10]byte
	n := 0
	value := uint64(v)

	for {
		if (value & ^uint64(0x7F)) == 0 {
			buf[n] = byte(value)
			n++
			break
		}
		buf[n] = byte((value & 0x7F) | 0x80)
		n++
		value >>= 7
	}

	return buf[:n], nil
}

// Len returns the number of bytes needed to encode this VarLong.
func (v VarLong) Len() int {
	value := uint64(v)
	n := 1
	for value >= 0x80 {
		value >>= 7
		n++
	}
	return n
}

// DecodeVarLong reads a VarLong from r.
func DecodeVarLong(r io.Reader) (VarLong, error) {
	var value int64
	var position uint
	var b [1]byte

	for {
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, err
		}

		value |= int64(b[0]&0x7F) << position

		if (b[0] & 0x80) == 0 {
			break
		}

		position += 7
		if position >= 70 {
			return 0, fmt.Errorf("VarLong is too big")
		}
	}

	return VarLong(value), nil
}
