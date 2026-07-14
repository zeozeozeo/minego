package net_structures

import (
	"io"
	"math"
)

// Angle represents a rotation angle.
//
// Stored as a single byte where 256 units = 360 degrees (full rotation).
// Each unit represents 1/256 of a full turn (1.40625 degrees).
//
// Examples:
//
//	0   = 0°
//	64  = 90°
//	128 = 180°
//	192 = 270°
type Angle uint8

// Encode writes the Angle to w.
func (a Angle) Encode(w io.Writer) error {
	_, err := w.Write([]byte{byte(a)})
	return err
}

// DecodeAngle reads an Angle from r.
func DecodeAngle(r io.Reader) (Angle, error) {
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return Angle(b[0]), nil
}

// AngleFromDegrees converts degrees to an Angle.
func AngleFromDegrees(degrees float64) Angle {
	// Normalize to 0-360 range
	degrees = math.Mod(degrees, 360)
	if degrees < 0 {
		degrees += 360
	}
	return Angle(degrees * 256 / 360)
}

// Degrees converts the Angle to degrees (0-360).
func (a Angle) Degrees() float64 {
	return float64(a) * 360 / 256
}

// Radians converts the Angle to radians (0-2π).
func (a Angle) Radians() float64 {
	return float64(a) * 2 * math.Pi / 256
}
