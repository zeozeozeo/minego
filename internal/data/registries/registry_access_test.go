package registries_test

import (
	"testing"

	"github.com/zeozeozeo/minego/internal/data/registries"
)

func TestNewRegistryAccess(t *testing.T) {
	ra := registries.NewRegistryAccess()

	// static registry should be accessible with correct entries
	block := ra.Lookup("minecraft:block")
	if block == nil {
		t.Fatal("Lookup(minecraft:block) returned nil")
	}
	if block.Get("minecraft:stone") != registries.Block.Get("minecraft:stone") {
		t.Error("static registry should have same entries as global")
	}

	// synchronized registry should exist but be empty (not yet populated from packets)
	biome := ra.Lookup("minecraft:worldgen/biome")
	if biome == nil {
		t.Fatal("Lookup(minecraft:worldgen/biome) returned nil")
	}
	if biome.Size() != 0 {
		t.Errorf("synchronized registry should start empty, got size %d", biome.Size())
	}
}

func TestStaticRegistryShared(t *testing.T) {
	ra := registries.NewRegistryAccess()

	// static registries should be the exact same pointer as the global
	if ra.Lookup("minecraft:block") != registries.Block {
		t.Error("static registry should share the global instance")
	}
	if ra.Lookup("minecraft:item") != registries.Item {
		t.Error("static registry should share the global instance")
	}
}

func TestSynchronizedRegistryNotShared(t *testing.T) {
	ra := registries.NewRegistryAccess()

	// synchronized registries should be independent instances, not shared globals
	biome := ra.Lookup("minecraft:worldgen/biome")
	if biome == nil {
		t.Fatal("biome registry should exist")
	}
	if biome.Identifier != "minecraft:worldgen/biome" {
		t.Errorf("biome identifier = %q, want minecraft:worldgen/biome", biome.Identifier)
	}
}

func TestApplyRegistryData(t *testing.T) {
	ra := registries.NewRegistryAccess()

	entries := []string{"custom:biome_a", "custom:biome_b", "minecraft:plains"}
	reg, err := ra.ApplyRegistryData("minecraft:worldgen/biome", entries)
	if err != nil {
		t.Fatalf("ApplyRegistryData: %v", err)
	}

	// entry order defines IDs
	if reg.Get("custom:biome_a") != 0 {
		t.Errorf("custom:biome_a = %d, want 0", reg.Get("custom:biome_a"))
	}
	if reg.Get("custom:biome_b") != 1 {
		t.Errorf("custom:biome_b = %d, want 1", reg.Get("custom:biome_b"))
	}
	if reg.Get("minecraft:plains") != 2 {
		t.Errorf("minecraft:plains = %d, want 2", reg.Get("minecraft:plains"))
	}

	// reverse lookup
	if reg.ByID(0) != "custom:biome_a" {
		t.Errorf("ByID(0) = %q, want custom:biome_a", reg.ByID(0))
	}

	if reg.Size() != 3 {
		t.Errorf("Size() = %d, want 3", reg.Size())
	}

	// Lookup should now return the new registry
	if ra.Lookup("minecraft:worldgen/biome") != reg {
		t.Error("Lookup should return the applied registry")
	}
}

func TestApplyRegistryDataUnknown(t *testing.T) {
	ra := registries.NewRegistryAccess()

	_, err := ra.ApplyRegistryData("minecraft:nonexistent", []string{"a"})
	if err == nil {
		t.Error("expected error for unknown registry")
	}
}

func TestApplyDoesNotAffectOtherAccess(t *testing.T) {
	ra1 := registries.NewRegistryAccess()
	ra2 := registries.NewRegistryAccess()

	ra1.ApplyRegistryData("minecraft:worldgen/biome", []string{"custom:only"})

	// ra2 should still have an empty biome registry
	if ra2.Lookup("minecraft:worldgen/biome").Size() != 0 {
		t.Error("applying data on one RegistryAccess should not affect another")
	}
}

func TestIsSynchronized(t *testing.T) {
	if !registries.IsSynchronized("minecraft:worldgen/biome") {
		t.Error("biome should be synchronized")
	}
	if !registries.IsSynchronized("minecraft:dimension_type") {
		t.Error("dimension_type should be synchronized")
	}
	if registries.IsSynchronized("minecraft:block") {
		t.Error("block should not be synchronized")
	}
	if registries.IsSynchronized("minecraft:item") {
		t.Error("item should not be synchronized")
	}
}

func TestLookupNil(t *testing.T) {
	ra := registries.NewRegistryAccess()
	if ra.Lookup("minecraft:nonexistent") != nil {
		t.Error("Lookup for unknown registry should return nil")
	}
}

func TestRegistryIdentifier(t *testing.T) {
	if registries.Block.Identifier != "minecraft:block" {
		t.Errorf("Block.Identifier = %q, want minecraft:block", registries.Block.Identifier)
	}
	if registries.Item.Identifier != "minecraft:item" {
		t.Errorf("Item.Identifier = %q, want minecraft:item", registries.Item.Identifier)
	}
}

func TestApplyPreservesIdentifier(t *testing.T) {
	ra := registries.NewRegistryAccess()
	reg, _ := ra.ApplyRegistryData("minecraft:dimension_type", []string{"minecraft:overworld", "minecraft:the_nether", "minecraft:the_end"})
	if reg.Identifier != "minecraft:dimension_type" {
		t.Errorf("applied registry Identifier = %q, want minecraft:dimension_type", reg.Identifier)
	}
}
