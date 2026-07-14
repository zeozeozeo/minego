package main

// JSON structures for server reports

type RegistryJSON struct {
	ProtocolID int32                        `json:"protocol_id"`
	Default    string                       `json:"default,omitempty"`
	Entries    map[string]RegistryEntryJSON `json:"entries"`
}

type RegistryEntryJSON struct {
	ProtocolID int32 `json:"protocol_id"`
}

type BlockJSON struct {
	Definition BlockDefinitionJSON `json:"definition"`
	Properties map[string][]string `json:"properties"`
	States     []BlockStateJSON    `json:"states"`
}

type BlockDefinitionJSON struct {
	Type string `json:"type"`
}

type BlockStateJSON struct {
	ID         int32             `json:"id"`
	Default    bool              `json:"default,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

type ItemJSON struct {
	Components map[string]any `json:"components"`
}

// PacketsJSON structure: phase -> bound -> packet_name -> {protocol_id}
type PacketsJSON map[string]map[string]map[string]PacketEntryJSON

type PacketEntryJSON struct {
	ProtocolID int32 `json:"protocol_id"`
}

// ComponentMetadata defines wire format info for a component.
type ComponentMetadata struct {
	// GoField is the field name in the Components struct (e.g., "MaxStackSize")
	GoField string `json:"goField,omitempty"`
	// GoType is the Go type (e.g., "int32", "*Food")
	GoType string `json:"goType,omitempty"`
	// WireType is the basic wire format type (varint, float32, identifier, empty, nbt, etc.)
	WireType string `json:"wireType"`
	// Passthrough means the component is decoded but not interpreted into a struct
	Passthrough bool `json:"passthrough,omitempty"`
	// WireFormat defines the detailed wire structure for complex components
	WireFormat []WireField `json:"wireFormat,omitempty"`
}

// WireField defines a single field in a complex component's wire format.
type WireField struct {
	Name    string `json:"name"`              // wire name for documentation
	Type    string `json:"type"`              // wire type: varint, float32, bool, nbt, etc.
	GoField string `json:"goField,omitempty"` // corresponding Go struct field
}

// ComponentMetadataFile is the structure of component_metadata.include.json.
type ComponentMetadataFile struct {
	Components map[string]ComponentMetadata `json:"components"`
}
