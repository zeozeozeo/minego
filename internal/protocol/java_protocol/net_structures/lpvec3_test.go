package net_structures_test

import (
	"bytes"
	"math"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// LpVec3 wire format:
//   - If all components < 3.05e-5: single 0x00 byte
//   - Otherwise: 6 bytes encoding scale(3) + X(15) + Y(15) + Z(15)
//
// First 2 bytes are little-endian, last 4 bytes are big-endian.

var lpVec3TestCases = []struct {
	name     string
	raw      []byte
	expected ns.LpVec3
	epsilon  float64
}{
	{
		name:     "zero vector",
		raw:      []byte{0x00},
		expected: ns.LpVec3{X: 0, Y: 0, Z: 0},
		epsilon:  0,
	},
}

func TestLpVec3(t *testing.T) {
	for _, tc := range lpVec3TestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			got, err := ns.NewReader(tc.raw).ReadLpVec3()
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if !lpVec3Equal(got, tc.expected, tc.epsilon) {
				t.Errorf("decode mismatch:\n  got:  %+v\n  want: %+v", got, tc.expected)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteLpVec3(tc.expected); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

func TestLpVec3_RoundTrip(t *testing.T) {
	cases := []struct {
		name    string
		vec     ns.LpVec3
		epsilon float64
	}{
		{"zero", ns.LpVec3{0, 0, 0}, 0},
		{"small positive", ns.LpVec3{0.001, 0.002, 0.003}, 0.0001},
		{"small negative", ns.LpVec3{-0.001, -0.002, -0.003}, 0.0001},
		{"unit x", ns.LpVec3{1.0, 0, 0}, 0.001},
		{"unit y", ns.LpVec3{0, 1.0, 0}, 0.001},
		{"unit z", ns.LpVec3{0, 0, 1.0}, 0.001},
		{"mixed", ns.LpVec3{0.5, -0.25, 0.125}, 0.001},
		{"typical velocity", ns.LpVec3{0.0784, -0.0784, 0}, 0.001},
		{"knockback north", ns.LpVec3{0, 0.4, -0.4}, 0.001},
		{"knockback south", ns.LpVec3{0, 0.4, 0.4}, 0.001},
		{"knockback east", ns.LpVec3{0.4, 0.4, 0}, 0.001},
		{"sprint knockback", ns.LpVec3{0, 0.4, -0.8}, 0.001},
		{"large velocity", ns.LpVec3{3.5, -2.1, 0.7}, 0.01},
		{"very large", ns.LpVec3{10.0, -10.0, 5.0}, 0.01},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteLpVec3(tc.vec); err != nil {
				t.Fatalf("encode error: %v", err)
			}

			got, err := ns.NewReader(buf.Bytes()).ReadLpVec3()
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if !lpVec3Equal(got, tc.vec, tc.epsilon) {
				t.Errorf("round-trip mismatch:\n  original: %+v\n  decoded:  %+v\n  raw: %x",
					tc.vec, got, buf.Bytes())
			}
		})
	}
}

func TestLpVec3_ZeroThreshold(t *testing.T) {
	threshold := 3.0517578125e-5

	// below threshold - should encode as single byte
	below := ns.LpVec3{X: threshold / 2, Y: threshold / 2, Z: threshold / 2}
	buf := ns.NewWriter()
	buf.WriteLpVec3(below)
	if len(buf.Bytes()) != 1 || buf.Bytes()[0] != 0x00 {
		t.Errorf("below threshold should encode as [0x00], got %x", buf.Bytes())
	}

	// at/above threshold - should encode as 6 bytes
	above := ns.LpVec3{X: threshold * 2, Y: 0, Z: 0}
	buf2 := ns.NewWriter()
	buf2.WriteLpVec3(above)
	if len(buf2.Bytes()) != 6 {
		t.Errorf("above threshold should encode as 6 bytes, got %d", len(buf2.Bytes()))
	}
}

func lpVec3Equal(a, b ns.LpVec3, epsilon float64) bool {
	return math.Abs(a.X-b.X) <= epsilon &&
		math.Abs(a.Y-b.Y) <= epsilon &&
		math.Abs(a.Z-b.Z) <= epsilon
}
