// Package blocks provides block collision shape lookups by state ID.
package blocks

import "github.com/zeozeozeo/minego/internal/data/versions/v26_1/hitboxes"

// CollisionShape returns the collision AABBs for a block state ID.
// Returns nil for invalid state IDs or blocks with no collision (air, etc.).
func CollisionShape(stateID int32) []hitboxes.AABB {
	if stateID < 0 || int(stateID) >= len(shapeByState) {
		return nil
	}
	return shapes[shapeByState[stateID]]
}

// HasCollision reports whether the given block state has any collision geometry.
func HasCollision(stateID int32) bool {
	if stateID < 0 || int(stateID) >= len(shapeByState) {
		return false
	}
	return shapeByState[stateID] != 0
}

// IsFullBlock reports whether the given block state is a full 1x1x1 collision block.
func IsFullBlock(stateID int32) bool {
	if stateID < 0 || int(stateID) >= len(shapeByState) {
		return false
	}
	return shapeByState[stateID] == fullBlockShapeIdx
}
