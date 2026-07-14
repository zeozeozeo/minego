// Package version defines the stable boundary between MineGo and generated
// Minecraft-version data.
package version

// State is a protocol connection phase.
type State uint8

const (
	Handshake State = iota
	Status
	Login
	Configuration
	Play
)

// Bound is a packet direction.
type Bound uint8

const (
	Serverbound Bound = iota
	Clientbound
)

// Block is version-neutral block metadata.
type Block struct {
	Name                string
	StateID             int32
	Properties          map[string]string
	Hardness            float32
	RequiresCorrectTool bool
	Collision           []AABB
}

// AABB is an axis-aligned box in block-local coordinates.
type AABB struct{ MinX, MinY, MinZ, MaxX, MaxY, MaxZ float64 }

// Pack supplies generated data and packet constructors for one game version.
// Packet returns an opaque implementation understood by MineGo's connection
// layer; callers normally use higher-level services instead.
type Pack interface {
	Name() string
	Protocol() int32
	BlockByState(stateID int32) (Block, bool)
	StateID(name string, properties map[string]string) (int32, bool)
	Packet(state State, bound Bound, id int32) (any, bool)
}
