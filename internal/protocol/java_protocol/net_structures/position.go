package net_structures

import "io"

// Position represents a block position in the world.
//
// Encoded as a 64-bit integer with the following bit layout:
//   - X: 26 bits (signed, bits 38-63)
//   - Z: 26 bits (signed, bits 12-37)
//   - Y: 12 bits (signed, bits 0-11)
//
// This allows coordinates:
//   - X, Z: -33554432 to 33554431
//   - Y: -2048 to 2047
type Position struct {
	X, Y, Z int
}

// NewPosition creates a new Position.
func NewPosition(x, y, z int) Position {
	return Position{X: x, Y: y, Z: z}
}

// Encode writes the Position to w as a packed 64-bit integer.
func (p Position) Encode(w io.Writer) error {
	return Int64(p.Pack()).Encode(w)
}

// DecodePosition reads a Position from r.
func DecodePosition(r io.Reader) (Position, error) {
	val, err := DecodeInt64(r)
	if err != nil {
		return Position{}, err
	}
	return UnpackPosition(int64(val)), nil
}

// Pack encodes the position into a 64-bit integer.
func (p Position) Pack() int64 {
	return ((int64(p.X) & 0x3FFFFFF) << 38) |
		((int64(p.Z) & 0x3FFFFFF) << 12) |
		(int64(p.Y) & 0xFFF)
}

// UnpackPosition decodes a 64-bit integer into a Position.
func UnpackPosition(val int64) Position {
	x := int(val >> 38)
	z := int(val << 26 >> 38)
	y := int(val << 52 >> 52)

	// Sign extension for X (26 bits)
	if x >= 1<<25 {
		x -= 1 << 26
	}
	// Sign extension for Z (26 bits)
	if z >= 1<<25 {
		z -= 1 << 26
	}
	// Sign extension for Y (12 bits)
	if y >= 1<<11 {
		y -= 1 << 12
	}

	return Position{X: x, Y: y, Z: z}
}

// GlobalPos represents a position in a specific dimension.
// Used for things like death locations.
//
// Wire format:
//
//	┌─────────────────────────┬─────────────────────────┐
//	│  Dimension (Identifier)  │  Position (Int64)       │
//	└─────────────────────────┴─────────────────────────┘
type GlobalPos struct {
	Dimension Identifier
	Pos       Position
}

// Encode writes the GlobalPos to w.
func (g GlobalPos) Encode(w io.Writer) error {
	if err := g.Dimension.Encode(w); err != nil {
		return err
	}
	return g.Pos.Encode(w)
}

// DecodeGlobalPos reads a GlobalPos from r.
func DecodeGlobalPos(r io.Reader) (GlobalPos, error) {
	dim, err := DecodeIdentifier(r)
	if err != nil {
		return GlobalPos{}, err
	}
	pos, err := DecodePosition(r)
	if err != nil {
		return GlobalPos{}, err
	}
	return GlobalPos{Dimension: dim, Pos: pos}, nil
}
