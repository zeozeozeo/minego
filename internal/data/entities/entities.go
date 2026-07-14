package entities

import "github.com/zeozeozeo/minego/internal/data/registries"

// EntityTypeID returns the protocol ID for an entity type identifier, or -1 if not found.
func EntityTypeID(name string) int32 {
	return registries.EntityType.Get(name)
}

// EntityTypeName returns the identifier for an entity type protocol ID, or empty string if not found.
func EntityTypeName(protocolID int32) string {
	return registries.EntityType.ByID(protocolID)
}

// EntityCategory returns the mob category for an entity type (e.g. "monster", "creature", "misc").
func EntityCategory(name string) string {
	return entityCategory[name]
}

// misc entities that extend LivingEntity and can be attacked
var attackableMisc = map[string]bool{
	"minecraft:player":       true,
	"minecraft:armor_stand":  true,
	"minecraft:mannequin":    true,
	"minecraft:copper_golem": true,
	"minecraft:iron_golem":   true,
	"minecraft:snow_golem":   true,
}

// IsAttackable returns true if the entity type is a LivingEntity (has health and can be attacked).
func IsAttackable(name string) bool {
	cat := entityCategory[name]
	if cat == "" {
		return false
	}
	if cat != "misc" {
		return true
	}
	return attackableMisc[name]
}
