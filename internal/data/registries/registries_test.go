package registries_test

import (
	"testing"

	"github.com/zeozeozeo/minego/internal/data/registries"
)

func TestBlockRegistry(t *testing.T) {
	tests := []struct {
		name string
		id   int32
	}{
		{"minecraft:air", 0},
		{"minecraft:stone", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := registries.Block.Get(tt.name); got != tt.id {
				t.Errorf("Block.Get(%q) = %d, want %d", tt.name, got, tt.id)
			}
			if got := registries.Block.ByID(tt.id); got != tt.name {
				t.Errorf("Block.ByID(%d) = %q, want %q", tt.id, got, tt.name)
			}
		})
	}
}

func TestItemRegistry(t *testing.T) {
	// item registry should have entries
	if got := registries.Item.Get("minecraft:diamond"); got == -1 {
		t.Error("Item.Get(diamond) should return valid ID")
	}
	if got := registries.Item.Get("minecraft:stick"); got == -1 {
		t.Error("Item.Get(stick) should return valid ID")
	}
}

func TestEntityTypeRegistry(t *testing.T) {
	// check some well-known entity types
	if got := registries.EntityType.Get("minecraft:player"); got == -1 {
		t.Error("EntityType.Get(player) should return valid ID")
	}
	if got := registries.EntityType.Get("minecraft:zombie"); got == -1 {
		t.Error("EntityType.Get(zombie) should return valid ID")
	}
	if got := registries.EntityType.Get("minecraft:creeper"); got == -1 {
		t.Error("EntityType.Get(creeper) should return valid ID")
	}
}

func TestRegistryNotFound(t *testing.T) {
	if got := registries.Block.Get("minecraft:nonexistent"); got != -1 {
		t.Errorf("Block.Get for nonexistent = %d, want -1", got)
	}
	if got := registries.Block.ByID(-999); got != "" {
		t.Errorf("Block.ByID for invalid ID = %q, want empty string", got)
	}
}

func TestDataComponentTypeRegistry(t *testing.T) {
	// data component type registry should exist and have entries
	if got := registries.DataComponentType.Get("minecraft:damage"); got == -1 {
		t.Error("DataComponentType.Get(damage) should return valid ID")
	}
	if got := registries.DataComponentType.Get("minecraft:max_damage"); got == -1 {
		t.Error("DataComponentType.Get(max_damage) should return valid ID")
	}
	if got := registries.DataComponentType.Get("minecraft:food"); got == -1 {
		t.Error("DataComponentType.Get(food) should return valid ID")
	}
}
