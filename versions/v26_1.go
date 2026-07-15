package versions

import (
	"github.com/zeozeozeo/minego/internal/data/versions/v26_1/blocks"
	hitbox "github.com/zeozeozeo/minego/internal/data/versions/v26_1/hitboxes/blocks"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_1/packet_ids"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_1/packets"
	"github.com/zeozeozeo/minego/version"
)

// V26_1 is the Minecraft Java Edition 26.1 / protocol 775 pack.
var V26_1 version.Pack = v26_1{}

// V26_1Data is retained as an alias for callers that used the pre-release data
// pack while 26.1's packet adapters were being generated.
var V26_1Data = V26_1

type v26_1 struct{}

func (v26_1) Descriptor() version.Descriptor {
	return version.NewDescriptor("26.1", 775,
		version.FeatureStatus, version.FeatureConfiguration, version.FeatureSignedChat, version.FeatureRuntimeRegistry,
		version.FeatureWorld, version.FeatureInventory, version.FeatureNavigation,
		version.FeatureMining, version.FeatureBuilding,
		version.FeatureCrafting, version.FeatureElytra,
		version.FeatureInteractions, version.FeatureCombat, version.FeatureContainers, version.FeatureRiding,
	)
}
func (v26_1) Name() string    { return "26.1" }
func (v26_1) Protocol() int32 { return 775 }

func (v26_1) BlockByState(id int32) (version.Block, bool) {
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

func (v26_1) StateID(name string, props map[string]string) (int32, bool) {
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

func (v26_1) PacketID(state version.State, bound version.Bound, name version.PacketName) (int32, bool) {
	// Keep this table explicit rather than assuming a 26.2 packet number.
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

func (v26_1) Packet(state version.State, bound version.Bound, id int32) (any, bool) {
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
