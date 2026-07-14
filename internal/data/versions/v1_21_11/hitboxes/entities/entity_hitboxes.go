// Package entities provides entity hitbox dimension lookups by entity type name.
package entities

// Dimensions returns the width, height, and eye height for an entity type.
// Returns zeros if the entity type is not found.
// Eye height of -1 means the entity uses the default: height * 0.85.
func Dimensions(name string) (width, height, eyeHeight float32) {
	idx, ok := dimensionsByName[name]
	if !ok {
		return 0, 0, 0
	}
	d := &dimensions[idx]
	return d.Width, d.Height, d.EyeHeight
}

// EyeHeight returns the effective eye height for an entity type,
// applying the default formula (height * 0.85) when no explicit value is set.
// Returns 0 if the entity type is not found.
func EyeHeight(name string) float32 {
	idx, ok := dimensionsByName[name]
	if !ok {
		return 0
	}
	d := &dimensions[idx]
	if d.EyeHeight < 0 {
		return d.Height * 0.85
	}
	return d.EyeHeight
}
