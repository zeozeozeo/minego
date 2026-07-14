package entities_test

import (
	"testing"

	"github.com/zeozeozeo/minego/internal/data/entities"
)

func TestEntityTypeIDLookup(t *testing.T) {
	names := []string{
		"minecraft:player", "minecraft:zombie", "minecraft:creeper",
		"minecraft:skeleton", "minecraft:spider", "minecraft:pig",
		"minecraft:cow", "minecraft:sheep", "minecraft:chicken",
		"minecraft:villager",
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			id := entities.EntityTypeID(name)
			if id < 0 {
				t.Fatalf("EntityTypeID(%q) = %d, want >= 0", name, id)
			}
			if got := entities.EntityTypeName(id); got != name {
				t.Errorf("EntityTypeName(%d) = %q, want %q", id, got, name)
			}
		})
	}
}

func TestEntityTypeIDNotFound(t *testing.T) {
	if got := entities.EntityTypeID("minecraft:nonexistent_entity"); got != -1 {
		t.Errorf("EntityTypeID for nonexistent entity = %d, want -1", got)
	}
	if got := entities.EntityTypeName(-999); got != "" {
		t.Errorf("EntityTypeName for invalid ID = %q, want empty string", got)
	}
}

func TestCommonEntityTypes(t *testing.T) {
	commonTypes := []string{
		"minecraft:player", "minecraft:zombie", "minecraft:creeper",
		"minecraft:skeleton", "minecraft:ender_dragon", "minecraft:wither",
		"minecraft:item", "minecraft:experience_orb", "minecraft:arrow",
		"minecraft:fireball",
	}

	for _, name := range commonTypes {
		id := entities.EntityTypeID(name)
		if id < 0 {
			t.Errorf("EntityTypeID(%q) = %d, want >= 0", name, id)
		}
		if got := entities.EntityTypeName(id); got != name {
			t.Errorf("EntityTypeName(%d) = %q, want %q", id, got, name)
		}
	}
}

func TestMetadataSerializerConstants(t *testing.T) {
	tests := []struct {
		name string
		id   int32
	}{
		{"BYTE", entities.SerializerBYTE},
		{"INT", entities.SerializerINT},
		{"FLOAT", entities.SerializerFLOAT},
		{"STRING", entities.SerializerSTRING},
		{"BOOLEAN", entities.SerializerBOOLEAN},
		{"ROTATIONS", entities.SerializerROTATIONS},
		{"BLOCK_POS", entities.SerializerBLOCK_POS},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.id < 0 {
				t.Errorf("serializer %s should have non-negative ID", tt.name)
			}
		})
	}
}

func TestMetadataFieldIndices(t *testing.T) {
	if entities.EntityIndexFlags != 0 {
		t.Errorf("EntityIndexFlags = %d, want 0", entities.EntityIndexFlags)
	}
	if entities.EntityIndexAirSupply != 1 {
		t.Errorf("EntityIndexAirSupply = %d, want 1", entities.EntityIndexAirSupply)
	}

	if entities.PlayerIndexAdditionalHearts < 8 {
		t.Errorf("PlayerIndexAdditionalHearts = %d, should be >= 8", entities.PlayerIndexAdditionalHearts)
	}

	if entities.CreeperIndexSwellDir < 8 {
		t.Errorf("CreeperIndexSwellDir = %d, should be >= 8", entities.CreeperIndexSwellDir)
	}
}

func TestMetadataOperations(t *testing.T) {
	var m entities.Metadata

	m.Set(0, entities.SerializerBYTE, []byte{0x01})
	m.Set(1, entities.SerializerINT, []byte{0x64})

	if data := m.Get(0); data == nil || data[0] != 0x01 {
		t.Errorf("Get(0) failed")
	}
	if data := m.Get(1); data == nil || data[0] != 0x64 {
		t.Errorf("Get(1) failed")
	}
	if data := m.Get(99); data != nil {
		t.Errorf("Get(99) should return nil for missing index")
	}

	m.Set(0, entities.SerializerBYTE, []byte{0x02})
	if data := m.Get(0); data == nil || data[0] != 0x02 {
		t.Errorf("Set update failed")
	}
}
