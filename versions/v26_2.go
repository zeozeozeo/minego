// Package versions contains isolated Minecraft protocol/data packs.
package versions

import (
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/blocks"
	hitbox "github.com/zeozeozeo/minego/internal/data/versions/v26_2/hitboxes/blocks"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packet_ids"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	"github.com/zeozeozeo/minego/version"
)

// V26_2 is the Minecraft Java Edition 26.2 / protocol 776 pack.
var V26_2 version.Pack = v26_2{}

type v26_2 struct{}

func (v26_2) Descriptor() version.Descriptor {
	return version.NewDescriptor("26.2", 776,
		version.FeatureStatus, version.FeatureConfiguration, version.FeatureSignedChat,
		version.FeatureRuntimeRegistry, version.FeatureWorld, version.FeatureInventory,
		version.FeatureNavigation, version.FeatureMining, version.FeatureBuilding,
		version.FeatureCrafting, version.FeatureElytra,
		version.FeatureInteractions, version.FeatureCombat, version.FeatureContainers, version.FeatureRiding,
	)
}
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

func (v26_2) PacketID(state version.State, bound version.Bound, name version.PacketName) (int32, bool) {
	if id, ok := connectionPacketIDs(state, bound, name); ok {
		return id, true
	}
	return 0, false
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

// connectionPacketIDs contains the normalized connection boundary shared by
// the 26.1 and 26.2 adapters. Play/configuration packet tables remain owned by
// their respective adapters rather than by public packet numbers.
func connectionPacketIDs(state version.State, bound version.Bound, name version.PacketName) (int32, bool) {
	switch state {
	case version.Handshake:
		if bound == version.Serverbound && name == version.PacketIntention {
			return packet_ids.C2SIntentionID, true
		}
	case version.Status:
		if bound == version.Serverbound && name == version.PacketStatusRequest {
			return packet_ids.C2SStatusRequestID, true
		}
		if bound == version.Clientbound && name == version.PacketStatusResponse {
			return packet_ids.S2CStatusResponseID, true
		}
	case version.Login:
		if bound == version.Serverbound {
			switch name {
			case version.PacketLoginHello:
				return packet_ids.C2SHelloID, true
			case version.PacketLoginEncryptionAnswer:
				return packet_ids.C2SKeyID, true
			case version.PacketLoginPluginResponse:
				return packet_ids.C2SCustomQueryAnswerID, true
			case version.PacketLoginAcknowledged:
				return packet_ids.C2SLoginAcknowledgedID, true
			}
		}
		if bound == version.Clientbound {
			switch name {
			case version.PacketLoginDisconnect:
				return packet_ids.S2CLoginDisconnectID, true
			case version.PacketLoginEncryption:
				return packet_ids.S2CHelloID, true
			case version.PacketLoginSuccess:
				return packet_ids.S2CLoginFinishedID, true
			case version.PacketLoginCompression:
				return packet_ids.S2CLoginCompressionID, true
			case version.PacketLoginPluginRequest:
				return packet_ids.S2CCustomQueryID, true
			}
		}
	}
	return 0, false
}
