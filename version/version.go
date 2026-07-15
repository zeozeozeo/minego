// Package version defines the stable boundary between MineGo and generated
// Minecraft-version data.
package version

import "sort"

// Feature identifies a capability whose wire representation or game mechanics
// may differ between Minecraft releases.
type Feature string

const (
	FeatureStatus          Feature = "status"
	FeatureConfiguration   Feature = "configuration"
	FeatureSignedChat      Feature = "signed_chat"
	FeatureRuntimeRegistry Feature = "runtime_registry"
	FeatureWorld           Feature = "world"
	FeatureInventory       Feature = "inventory"
	FeatureNavigation      Feature = "navigation"
	FeatureMining          Feature = "mining"
	FeatureBuilding        Feature = "building"
	FeatureCrafting        Feature = "crafting"
	FeatureElytra          Feature = "elytra"
)

// Descriptor is stable, immutable metadata for a Minecraft version pack.
type Descriptor struct {
	Name     string
	Protocol int32
	Features []Feature
}

func (d Descriptor) Supports(feature Feature) bool {
	for _, f := range d.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// NewDescriptor returns a descriptor with sorted, deduplicated capabilities.
func NewDescriptor(name string, protocol int32, features ...Feature) Descriptor {
	set := make(map[Feature]struct{}, len(features))
	for _, feature := range features {
		set[feature] = struct{}{}
	}
	features = features[:0]
	for feature := range set {
		features = append(features, feature)
	}
	sort.Slice(features, func(i, j int) bool { return features[i] < features[j] })
	return Descriptor{Name: name, Protocol: protocol, Features: features}
}

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

// PacketName is the normalized identity of a wire packet. It deliberately
// names protocol semantics rather than exposing a release-specific numeric ID.
// Packs may map a name to different IDs or decline a packet entirely.
type PacketName string

const (
	PacketIntention             PacketName = "intention"
	PacketStatusRequest         PacketName = "status_request"
	PacketStatusResponse        PacketName = "status_response"
	PacketLoginHello            PacketName = "login_hello"
	PacketLoginCompression      PacketName = "login_compression"
	PacketLoginEncryption       PacketName = "login_encryption"
	PacketLoginPluginRequest    PacketName = "login_plugin_request"
	PacketLoginPluginResponse   PacketName = "login_plugin_response"
	PacketLoginSuccess          PacketName = "login_success"
	PacketLoginAcknowledged     PacketName = "login_acknowledged"
	PacketLoginDisconnect       PacketName = "login_disconnect"
	PacketLoginEncryptionAnswer PacketName = "login_encryption_answer"
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
	Descriptor() Descriptor
	Name() string
	Protocol() int32
	BlockByState(stateID int32) (Block, bool)
	StateID(name string, properties map[string]string) (int32, bool)
	// PacketID resolves a normalized packet name for this release.
	PacketID(state State, bound Bound, name PacketName) (int32, bool)
	Packet(state State, bound Bound, id int32) (any, bool)
}
