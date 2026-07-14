package net_structures

import (
	"encoding/binary"
	"fmt"
	"math"
)

// LpVec3 is a low-precision 3D vector used for entity velocity.
// Matches net.minecraft.network.LpVec3 from the Minecraft Java Edition source.
//
// Wire format:
//   - 1 byte if zero vector (0x00)
//   - Otherwise 6 bytes + optional VarInt for large scale:
//     byte 0 (lowest): bits 0-1 = scale low, bit 2 = continuation, bits 3-7 = x[0:4]
//     byte 1 (middle): x[5:12]
//     bytes 2-5 (big-endian uint32): x[13:14] | y[0:14] | z[0:14]
//
// Components are unsigned 15-bit values in [0, 32766] mapped to [-1, +1],
// then multiplied by the integer scale factor.
type LpVec3 struct {
	X, Y, Z float64
}

const (
	lpDataBits    = 15
	lpDataMask    = (1 << lpDataBits) - 1 // 32767
	lpMaxQuantize = 32766.0
	lpScaleMask   = 3
	lpContFlag    = 4
	lpAbsMin      = 3.051944088384301e-5
	lpAbsMax      = 1.7179869183e10
)

// Decode reads an LpVec3 from the buffer.
func (v *LpVec3) Decode(buf *PacketBuffer) error {
	lowest, err := buf.ReadByte()
	if err != nil {
		return fmt.Errorf("lpvec3: %w", err)
	}

	if lowest == 0 {
		v.X, v.Y, v.Z = 0, 0, 0
		return nil
	}

	middle, err := buf.ReadByte()
	if err != nil {
		return fmt.Errorf("lpvec3: %w", err)
	}

	var rest [4]byte
	if _, err := buf.Read(rest[:]); err != nil {
		return fmt.Errorf("lpvec3: %w", err)
	}
	highest := uint64(binary.BigEndian.Uint32(rest[:]))

	// reconstruct 48-bit buffer: highest(32) << 16 | middle(8) << 8 | lowest(8)
	buffer := highest<<16 | uint64(middle)<<8 | uint64(lowest)

	// scale from bottom 2 bits, with optional continuation via VarInt
	scale := uint64(lowest & lpScaleMask)
	if lowest&lpContFlag != 0 {
		extra, err := buf.ReadVarInt()
		if err != nil {
			return fmt.Errorf("lpvec3 continuation: %w", err)
		}
		scale |= uint64(uint32(extra)) << 2
	}

	// x at bits 3-17, y at bits 18-32, z at bits 33-47
	v.X = lpUnpack(buffer>>3) * float64(scale)
	v.Y = lpUnpack(buffer>>18) * float64(scale)
	v.Z = lpUnpack(buffer>>33) * float64(scale)

	return nil
}

// lpUnpack maps a 15-bit unsigned value [0, 32766] to [-1.0, +1.0].
// Matches: Math.min((double)(value & 32767L), 32766.0) * 2.0 / 32766.0 - 1.0
func lpUnpack(value uint64) float64 {
	raw := float64(value & lpDataMask)
	return math.Min(raw, lpMaxQuantize)*2.0/lpMaxQuantize - 1.0
}

// Encode writes an LpVec3 to the buffer.
func (v *LpVec3) Encode(buf *PacketBuffer) error {
	x := lpSanitize(v.X)
	y := lpSanitize(v.Y)
	z := lpSanitize(v.Z)

	maxAbs := math.Max(math.Abs(x), math.Max(math.Abs(y), math.Abs(z)))
	if maxAbs < lpAbsMin {
		return buf.WriteByte(0)
	}

	scale := int64(math.Ceil(maxAbs))
	isPartial := (scale & lpScaleMask) != scale

	var markers int64
	if isPartial {
		markers = (scale & lpScaleMask) | lpContFlag
	} else {
		markers = scale
	}

	xn := lpPack(x/float64(scale)) << 3
	yn := lpPack(y/float64(scale)) << 18
	zn := lpPack(z/float64(scale)) << 33
	buffer := uint64(markers) | uint64(xn) | uint64(yn) | uint64(zn)

	if err := buf.WriteByte(byte(buffer)); err != nil {
		return err
	}
	if err := buf.WriteByte(byte(buffer >> 8)); err != nil {
		return err
	}

	var rest [4]byte
	binary.BigEndian.PutUint32(rest[:], uint32(buffer>>16))
	if _, err := buf.Write(rest[:]); err != nil {
		return err
	}

	if isPartial {
		if err := buf.WriteVarInt(VarInt(scale >> 2)); err != nil {
			return err
		}
	}

	return nil
}

func lpSanitize(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	return max(min(value, lpAbsMax), -lpAbsMax)
}

// lpPack maps [-1.0, +1.0] to [0, 32766].
// Matches: Math.round((value * 0.5 + 0.5) * 32766.0)
func lpPack(value float64) int64 {
	return int64(math.Round((value*0.5 + 0.5) * lpMaxQuantize))
}

// ReadLpVec3 reads an LpVec3 from the buffer.
func (pb *PacketBuffer) ReadLpVec3() (LpVec3, error) {
	var v LpVec3
	err := v.Decode(pb)
	return v, err
}

// WriteLpVec3 writes an LpVec3 to the buffer.
func (pb *PacketBuffer) WriteLpVec3(v LpVec3) error {
	return v.Encode(pb)
}
