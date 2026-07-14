package blocks_test

import (
	"testing"

	"github.com/zeozeozeo/minego/internal/data/blocks"
)

func TestBlockIDLookup(t *testing.T) {
	names := []string{
		"minecraft:air",
		"minecraft:stone",
		"minecraft:dirt",
		"minecraft:oak_planks",
		"minecraft:diamond_block",
		"minecraft:iron_block",
		"minecraft:gold_block",
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			id := blocks.BlockID(name)
			if id < 0 {
				t.Fatalf("BlockID(%q) = %d, want >= 0", name, id)
			}
			if got := blocks.BlockName(id); got != name {
				t.Errorf("BlockName(%d) = %q, want %q", id, got, name)
			}
		})
	}
}

func TestBlockIDNotFound(t *testing.T) {
	if got := blocks.BlockID("minecraft:nonexistent_block"); got != -1 {
		t.Errorf("BlockID for nonexistent block = %d, want -1", got)
	}
	if got := blocks.BlockName(-999); got != "" {
		t.Errorf("BlockName for invalid ID = %q, want empty string", got)
	}
}
