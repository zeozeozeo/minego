package net_structures_test

import (
	"bytes"
	"math"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// Angle wire format: 1 byte (1/256 of full rotation)
// 0 = 0°, 64 = 90°, 128 = 180°, 192 = 270°

var angleTestCases = []struct {
	name    string
	raw     []byte
	value   ns.Angle
	degrees float64
	radians float64
}{
	{"0°", []byte{0x00}, 0, 0, 0},
	{"90°", []byte{0x40}, 64, 90, math.Pi / 2},
	{"180°", []byte{0x80}, 128, 180, math.Pi},
	{"270°", []byte{0xc0}, 192, 270, 3 * math.Pi / 2},
	{"255", []byte{0xff}, 255, 358.59375, 0}, // ~359°
}

func TestAngle(t *testing.T) {
	for _, tc := range angleTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			got, err := ns.NewReader(tc.raw).ReadAngle()
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got != tc.value {
				t.Errorf("got %d, want %d", got, tc.value)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteAngle(tc.value); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("got %x, want %x", buf.Bytes(), tc.raw)
			}
		})

		t.Run(tc.name+" degrees", func(t *testing.T) {
			if got := tc.value.Degrees(); math.Abs(got-tc.degrees) > 0.0001 {
				t.Errorf("Degrees() = %v, want %v", got, tc.degrees)
			}
		})

		if tc.radians != 0 {
			t.Run(tc.name+" radians", func(t *testing.T) {
				if got := tc.value.Radians(); math.Abs(got-tc.radians) > 0.0001 {
					t.Errorf("Radians() = %v, want %v", got, tc.radians)
				}
			})
		}
	}
}

func TestAngle_FromDegrees(t *testing.T) {
	cases := []struct {
		degrees float64
		angle   ns.Angle
	}{
		{0, 0}, {90, 64}, {180, 128}, {270, 192},
		{360, 0},   // wraps
		{-90, 192}, // wraps to 270°
		{45, 32},
	}
	for _, tc := range cases {
		if got := ns.AngleFromDegrees(tc.degrees); got != tc.angle {
			t.Errorf("AngleFromDegrees(%v) = %d, want %d", tc.degrees, got, tc.angle)
		}
	}
}
