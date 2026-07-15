package versions

import (
	"github.com/zeozeozeo/minego/internal/data/versions/v1_21_11/blocks"
	hitbox "github.com/zeozeozeo/minego/internal/data/versions/v1_21_11/hitboxes/blocks"
	"github.com/zeozeozeo/minego/internal/data/versions/v1_21_11/packet_ids"
	"github.com/zeozeozeo/minego/internal/data/versions/v1_21_11/packets"
	"github.com/zeozeozeo/minego/version"
)

// V1_21_11 is the Minecraft Java Edition 1.21.11 / protocol 774 pack.
var V1_21_11 version.Pack = v1_21_11{}

type v1_21_11 struct{}

func (v1_21_11) Descriptor() version.Descriptor {
	return version.NewDescriptor("1.21.11", 774,
		version.FeatureStatus, version.FeatureConfiguration, version.FeatureSignedChat,
		version.FeatureRuntimeRegistry, version.FeatureWorld, version.FeatureInventory,
		version.FeatureNavigation, version.FeatureMining, version.FeatureBuilding,
		version.FeatureInteractions, version.FeatureCombat, version.FeatureContainers, version.FeatureRiding)
}
func (v1_21_11) Name() string    { return "1.21.11" }
func (v1_21_11) Protocol() int32 { return 774 }
func (v1_21_11) BlockByState(id int32) (version.Block, bool) {
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
func (v1_21_11) StateID(name string, props map[string]string) (int32, bool) {
	id := blocks.BlockID(name)
	if id < 0 {
		return -1, false
	}
	state := blocks.DefaultStateID(id)
	if props != nil {
		state = blocks.StateID(int(id), props)
	}
	return state, state >= 0
}
func (v1_21_11) PacketID(state version.State, bound version.Bound, name version.PacketName) (int32, bool) {
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
func (v1_21_11) Packet(state version.State, bound version.Bound, id int32) (any, bool) {
	states := [...]string{"handshake", "status", "login", "configuration", "play"}
	if int(state) >= len(states) {
		return nil, false
	}
	direction := "c2s"
	if bound == version.Clientbound {
		direction = "s2c"
	}
	factory := packets.PacketRegistries[states[state]+"_"+direction][int(id)]
	if factory == nil {
		return nil, false
	}
	return factory(), true
}
