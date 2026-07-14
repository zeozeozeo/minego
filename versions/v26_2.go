// Package versions contains isolated Minecraft protocol/data packs.
package versions

import (
	"github.com/zeozeozeo/minego/internal/data/blocks"
	hitbox "github.com/zeozeozeo/minego/internal/data/hitboxes/blocks"
	"github.com/zeozeozeo/minego/internal/data/packets"
	"github.com/zeozeozeo/minego/version"
)

// V26_2 is the Minecraft Java Edition 26.2 / protocol 776 pack.
var V26_2 version.Pack = v26_2{}

type v26_2 struct{}

func (v26_2) Name() string    { return "26.2" }
func (v26_2) Protocol() int32 { return 776 }

func (v26_2) BlockByState(id int32) (version.Block, bool) {
	blockID, props := blocks.StateProperties(int(id))
	if blockID < 0 {
		return version.Block{}, false
	}
	name := blocks.BlockName(blockID)
	shape := hitbox.CollisionShape(id)
	boxes := make([]version.AABB, len(shape))
	for i, b := range shape {
		boxes[i] = version.AABB{MinX: b[0], MinY: b[1], MinZ: b[2], MaxX: b[3], MaxY: b[4], MaxZ: b[5]}
	}
	return version.Block{Name: name, StateID: id, Properties: props, Hardness: blocks.BlockHardness(name), RequiresCorrectTool: blocks.BlockRequiresCorrectTool(name), Collision: boxes}, true
}

func (v26_2) StateID(name string, props map[string]string) (int32, bool) {
	id := blocks.BlockID(name)
	if id < 0 {
		return -1, false
	}
	var state int32
	if props == nil {
		state = blocks.DefaultStateID(id)
	} else {
		state = blocks.StateID(int(id), props)
	}
	return state, state >= 0
}

func (v26_2) Packet(state version.State, bound version.Bound, id int32) (any, bool) {
	states := [...]string{"handshake", "status", "login", "configuration", "play"}
	if int(state) >= len(states) {
		return nil, false
	}
	dir := "c2s"
	if bound == version.Clientbound {
		dir = "s2c"
	}
	factory := packets.PacketRegistries[states[state]+"_"+dir][int(id)]
	if factory == nil {
		return nil, false
	}
	return factory(), true
}
