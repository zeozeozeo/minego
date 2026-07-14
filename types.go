package minego

import "math"

type Vec3 struct{ X, Y, Z float64 }

func (v Vec3) Distance(o Vec3) float64 {
	x, y, z := v.X-o.X, v.Y-o.Y, v.Z-o.Z
	return math.Sqrt(x*x + y*y + z*z)
}

type BlockPos struct{ X, Y, Z int }

func (v Vec3) Block() BlockPos {
	return BlockPos{int(math.Floor(v.X)), int(math.Floor(v.Y)), int(math.Floor(v.Z))}
}

type Rotation struct{ Yaw, Pitch float32 }

type DisconnectEvent struct {
	Reason string
	Err    error
}
type BlockChange struct {
	Position BlockPos
	Old, New Block
}
type EntityChange struct {
	Kind   string
	Entity Entity
}

type Block struct {
	Position            BlockPos
	Name                string
	StateID             int32
	Properties          map[string]string
	Hardness            float32
	RequiresCorrectTool bool
	Collision           []AABB
}
type AABB struct{ MinX, MinY, MinZ, MaxX, MaxY, MaxZ float64 }
