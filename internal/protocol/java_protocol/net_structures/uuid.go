package net_structures

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

// UUID is a 128-bit universally unique identifier.
//
// Encoded as two big-endian 64-bit integers (most significant bits first).
type UUID [16]byte

// NilUUID is the zero UUID (all zeros).
var NilUUID = UUID{}

// Encode writes the UUID to w.
func (u UUID) Encode(w io.Writer) error {
	_, err := w.Write(u[:])
	return err
}

// DecodeUUID reads a UUID from r.
func DecodeUUID(r io.Reader) (UUID, error) {
	var u UUID
	if _, err := io.ReadFull(r, u[:]); err != nil {
		return UUID{}, err
	}
	return u, nil
}

// UUIDFromBytes creates a UUID from a 16-byte slice.
func UUIDFromBytes(b []byte) (UUID, error) {
	if len(b) != 16 {
		return UUID{}, fmt.Errorf("invalid UUID byte length: %d", len(b))
	}
	var u UUID
	copy(u[:], b)
	return u, nil
}

// UUIDFromString parses a UUID from its string representation.
// Accepts formats: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" or "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
func UUIDFromString(s string) (UUID, error) {
	// Remove hyphens if present
	clean := make([]byte, 0, 32)
	for i := 0; i < len(s); i++ {
		if s[i] != '-' {
			clean = append(clean, s[i])
		}
	}

	if len(clean) != 32 {
		return UUID{}, fmt.Errorf("invalid UUID string length: %d", len(clean))
	}

	var u UUID
	_, err := hex.Decode(u[:], clean)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid UUID hex: %w", err)
	}

	return u, nil
}

// String returns the UUID in standard hyphenated format.
func (u UUID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

// MostSignificantBits returns the first 64 bits of the UUID.
func (u UUID) MostSignificantBits() int64 {
	return int64(u[0])<<56 | int64(u[1])<<48 | int64(u[2])<<40 | int64(u[3])<<32 |
		int64(u[4])<<24 | int64(u[5])<<16 | int64(u[6])<<8 | int64(u[7])
}

// LeastSignificantBits returns the last 64 bits of the UUID.
func (u UUID) LeastSignificantBits() int64 {
	return int64(u[8])<<56 | int64(u[9])<<48 | int64(u[10])<<40 | int64(u[11])<<32 |
		int64(u[12])<<24 | int64(u[13])<<16 | int64(u[14])<<8 | int64(u[15])
}

// UUIDFromInt64s creates a UUID from most and least significant bits.
func UUIDFromInt64s(msb, lsb int64) UUID {
	var u UUID
	u[0] = byte(msb >> 56)
	u[1] = byte(msb >> 48)
	u[2] = byte(msb >> 40)
	u[3] = byte(msb >> 32)
	u[4] = byte(msb >> 24)
	u[5] = byte(msb >> 16)
	u[6] = byte(msb >> 8)
	u[7] = byte(msb)
	u[8] = byte(lsb >> 56)
	u[9] = byte(lsb >> 48)
	u[10] = byte(lsb >> 40)
	u[11] = byte(lsb >> 32)
	u[12] = byte(lsb >> 24)
	u[13] = byte(lsb >> 16)
	u[14] = byte(lsb >> 8)
	u[15] = byte(lsb)
	return u
}

// IsNil returns true if this is the nil UUID (all zeros).
func (u UUID) IsNil() bool {
	return u == NilUUID
}

// ValidateUUID checks if a string is a valid UUID format.
// Accepts formats: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" (36 chars) or "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" (32 chars)
func ValidateUUID(uuid string) bool {
	if len(uuid) == 36 && strings.Count(uuid, "-") == 4 {
		return true
	}
	if len(uuid) == 32 && strings.Count(uuid, "-") == 0 {
		return true
	}
	return false
}
