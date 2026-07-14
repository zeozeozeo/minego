package items_test

import (
	"testing"

	"github.com/zeozeozeo/minego/internal/data/items"
)

func TestItemIDLookup(t *testing.T) {
	tests := []string{
		"minecraft:diamond_sword",
		"minecraft:iron_sword",
		"minecraft:apple",
		"minecraft:golden_apple",
		"minecraft:diamond_pickaxe",
		"minecraft:iron_pickaxe",
		"minecraft:stick",
		"minecraft:diamond",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			id := items.ItemID(name)
			if id < 0 {
				t.Fatalf("ItemID(%q) = %d, want >= 0", name, id)
			}
			if got := items.ItemName(id); got != name {
				t.Errorf("ItemName(%d) = %q, want %q", id, got, name)
			}
		})
	}
}

func TestItemIDNotFound(t *testing.T) {
	if got := items.ItemID("minecraft:nonexistent_item"); got != -1 {
		t.Errorf("ItemID for nonexistent item = %d, want -1", got)
	}
	if got := items.ItemName(-999); got != "" {
		t.Errorf("ItemName for invalid ID = %q, want empty string", got)
	}
}

func TestDefaultComponents(t *testing.T) {
	appleID := items.ItemID("minecraft:apple")
	apple := items.DefaultComponents(appleID)
	if apple == nil {
		t.Fatal("DefaultComponents(apple) = nil")
	}
	if apple.Food == nil {
		t.Error("apple should have Food component")
	} else if apple.Food.Nutrition != 4 {
		t.Errorf("apple nutrition = %d, want 4", apple.Food.Nutrition)
	}

	swordID := items.ItemID("minecraft:diamond_sword")
	sword := items.DefaultComponents(swordID)
	if sword == nil {
		t.Fatal("DefaultComponents(diamond_sword) = nil")
	}
	if sword.MaxDamage != 1561 {
		t.Errorf("diamond_sword max damage = %d, want 1561", sword.MaxDamage)
	}
}

func TestComponentConstants(t *testing.T) {
	if items.ComponentDamage < 0 {
		t.Error("ComponentDamage should be non-negative")
	}
	if items.ComponentMaxDamage < 0 {
		t.Error("ComponentMaxDamage should be non-negative")
	}
	if items.ComponentFood < 0 {
		t.Error("ComponentFood should be non-negative")
	}
	if items.MaxComponentID < items.ComponentFood {
		t.Error("MaxComponentID should be >= ComponentFood")
	}
}
