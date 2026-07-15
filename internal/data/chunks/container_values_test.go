package chunks

import (
	"slices"
	"testing"
)

func TestPalettedContainerValues(t *testing.T) {
	p := NewSingleValue(BlockStatesKind, 7)
	if got := p.Values(); !slices.Equal(got, []int32{7}) {
		t.Fatalf("single values = %v", got)
	}
	p.Set(0, 11)
	values := p.Values()
	if !slices.Contains(values, int32(7)) || !slices.Contains(values, int32(11)) {
		t.Fatalf("indirect values = %v", values)
	}
	values[0] = 99
	if slices.Contains(p.Values(), int32(99)) {
		t.Fatal("Values exposed the internal palette")
	}
}
