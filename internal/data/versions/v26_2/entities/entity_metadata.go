package entities

import (
	"fmt"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/items"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"github.com/zeozeozeo/minego/internal/protocol/nbt"
)

// Rotations represents 3D rotation angles (x, y, z in degrees).
type Rotations struct {
	X, Y, Z float32
}

// Position represents a block position (packed 64-bit format).
type Position struct {
	X, Y, Z int32
}

// Metadata represents raw entity metadata entries.
// Use this for passthrough or when you don't need typed access.
type Metadata []MetadataEntry

// ReadMetadata reads entity metadata from a packet buffer.
// Wire format: [Index(UByte)][Type(VarInt)][Value]...[0xFF terminator]
func ReadMetadata(buf *ns.PacketBuffer) (Metadata, error) {
	var entries Metadata

	for {
		index, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		if index == 0xFF {
			break
		}

		serializerID, err := buf.ReadVarInt()
		if err != nil {
			return nil, err
		}

		data, err := readSerializerValue(buf, int32(serializerID))
		if err != nil {
			return nil, fmt.Errorf("reading metadata index %d (serializer %d): %w", index, serializerID, err)
		}

		entries = append(entries, MetadataEntry{
			Index:      index,
			Serializer: int32(serializerID),
			Data:       data,
		})
	}

	return entries, nil
}

// WriteMetadata writes entity metadata to a packet buffer.
func WriteMetadata(buf *ns.PacketBuffer, entries Metadata) error {
	for _, entry := range entries {
		buf.WriteByte(entry.Index)
		buf.WriteVarInt(ns.VarInt(entry.Serializer))
		if _, err := buf.Write(entry.Data); err != nil {
			return err
		}
	}
	buf.WriteByte(0xFF) // terminator
	return nil
}

// Get returns the data for a specific metadata index, or nil if not present.
func (m Metadata) Get(index byte) []byte {
	for _, entry := range m {
		if entry.Index == index {
			return entry.Data
		}
	}
	return nil
}

// Set sets or adds a metadata entry.
func (m *Metadata) Set(index byte, serializerID int32, data []byte) {
	for i, entry := range *m {
		if entry.Index == index {
			(*m)[i].Serializer = serializerID
			(*m)[i].Data = data
			return
		}
	}
	*m = append(*m, MetadataEntry{
		Index:      index,
		Serializer: serializerID,
		Data:       data,
	})
}

// readSerializerValue reads a value based on serializer type.
func readSerializerValue(buf *ns.PacketBuffer, serializerID int32) ([]byte, error) {
	w := ns.NewWriter()

	wireType, ok := serializerWireTypes[serializerID]
	if !ok {
		return nil, fmt.Errorf("unknown serializer ID %d", serializerID)
	}

	switch wireType {
	case "byte":
		if err := w.CopyInt8(buf); err != nil {
			return nil, err
		}

	case "varint":
		if err := w.CopyVarInt(buf); err != nil {
			return nil, err
		}

	case "varlong":
		if err := w.CopyVarLong(buf); err != nil {
			return nil, err
		}

	case "float32":
		if err := w.CopyFloat32(buf); err != nil {
			return nil, err
		}

	case "string":
		if err := w.CopyString(buf, 32767); err != nil {
			return nil, err
		}

	case "nbt":
		if err := nbt.Copy(w.Writer(), buf.Reader(), true); err != nil {
			return nil, err
		}

	case "optional_nbt":
		present, err := buf.ReadBool()
		if err != nil {
			return nil, err
		}
		w.WriteBool(present)
		if present {
			if err := nbt.Copy(w.Writer(), buf.Reader(), true); err != nil {
				return nil, err
			}
		}

	case "slot":
		// read slot with items decoder, then re-encode to bytes
		slot, err := buf.ReadSlot(items.Decoder())
		if err != nil {
			return nil, err
		}
		if err := w.WriteSlot(slot); err != nil {
			return nil, err
		}

	case "bool":
		if err := w.CopyBool(buf); err != nil {
			return nil, err
		}

	case "rotations":
		// 3x float32 (x, y, z rotation)
		for range 3 {
			if err := w.CopyFloat32(buf); err != nil {
				return nil, err
			}
		}

	case "position":
		if err := w.CopyPosition(buf); err != nil {
			return nil, err
		}

	case "optional_position":
		present, err := buf.ReadBool()
		if err != nil {
			return nil, err
		}
		w.WriteBool(present)
		if present {
			if err := w.CopyPosition(buf); err != nil {
				return nil, err
			}
		}

	case "optional_uuid":
		present, err := buf.ReadBool()
		if err != nil {
			return nil, err
		}
		w.WriteBool(present)
		if present {
			if err := w.CopyUUID(buf); err != nil {
				return nil, err
			}
		}

	case "particle":
		if err := copyParticle(buf, w); err != nil {
			return nil, err
		}

	case "particle_list":
		count, err := buf.ReadVarInt()
		if err != nil {
			return nil, err
		}
		w.WriteVarInt(count)
		for range int(count) {
			if err := copyParticle(buf, w); err != nil {
				return nil, err
			}
		}

	case "villager_data":
		// type, profession, level (all VarInt)
		for range 3 {
			if err := w.CopyVarInt(buf); err != nil {
				return nil, err
			}
		}

	case "optional_varint":
		// varint with 0 = absent, 1+ = value+1
		if err := w.CopyVarInt(buf); err != nil {
			return nil, err
		}

	case "optional_global_pos":
		present, err := buf.ReadBool()
		if err != nil {
			return nil, err
		}
		w.WriteBool(present)
		if present {
			// dimension (identifier)
			if err := w.CopyString(buf, 32767); err != nil {
				return nil, err
			}
			// position
			if err := w.CopyPosition(buf); err != nil {
				return nil, err
			}
		}

	case "id_or_inline":
		// painting variant: either registry ID or inline data
		typeID, err := buf.ReadVarInt()
		if err != nil {
			return nil, err
		}
		w.WriteVarInt(typeID)
		if typeID == 0 {
			// inline: asset_id, width, height
			if err := w.CopyString(buf, 32767); err != nil {
				return nil, err
			}
			for range 2 {
				if err := w.CopyVarInt(buf); err != nil {
					return nil, err
				}
			}
			// optional title (NBT)
			hasTitle, err := buf.ReadBool()
			if err != nil {
				return nil, err
			}
			w.WriteBool(hasTitle)
			if hasTitle {
				if err := nbt.Copy(w.Writer(), buf.Reader(), true); err != nil {
					return nil, err
				}
			}
			// optional author (NBT)
			hasAuthor, err := buf.ReadBool()
			if err != nil {
				return nil, err
			}
			w.WriteBool(hasAuthor)
			if hasAuthor {
				if err := nbt.Copy(w.Writer(), buf.Reader(), true); err != nil {
					return nil, err
				}
			}
		}

	case "vector3f":
		// 3x float32
		for range 3 {
			if err := w.CopyFloat32(buf); err != nil {
				return nil, err
			}
		}

	case "quaternionf":
		// 4x float32
		for range 4 {
			if err := w.CopyFloat32(buf); err != nil {
				return nil, err
			}
		}

	case "resolvable_profile":
		// optional name
		hasName, err := buf.ReadBool()
		if err != nil {
			return nil, err
		}
		w.WriteBool(hasName)
		if hasName {
			if err := w.CopyString(buf, 32767); err != nil {
				return nil, err
			}
		}
		// optional uuid
		hasUUID, err := buf.ReadBool()
		if err != nil {
			return nil, err
		}
		w.WriteBool(hasUUID)
		if hasUUID {
			if err := w.CopyUUID(buf); err != nil {
				return nil, err
			}
		}
		// properties
		propCount, err := buf.ReadVarInt()
		if err != nil {
			return nil, err
		}
		w.WriteVarInt(propCount)
		for range int(propCount) {
			// name, value
			if err := w.CopyString(buf, 32767); err != nil {
				return nil, err
			}
			if err := w.CopyString(buf, 32767); err != nil {
				return nil, err
			}
			// optional signature
			hasSig, err := buf.ReadBool()
			if err != nil {
				return nil, err
			}
			w.WriteBool(hasSig)
			if hasSig {
				if err := w.CopyString(buf, 32767); err != nil {
					return nil, err
				}
			}
		}

	case "empty":
		// no data

	default:
		return nil, fmt.Errorf("unhandled wire type %q for serializer %d", wireType, serializerID)
	}

	return w.Bytes(), nil
}

// copyParticle copies a particle from buf to w.
func copyParticle(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	particleType, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(particleType)

	// particle data varies by type - for now just copy the type ID
	// full particle data parsing would require the particle registry
	// TODO: implement Particle type, maybe in pkg/data/misc/particles_gen.go or something?
	return nil
}

// SerializerName returns the name of a serializer by ID, or empty string if unknown.
func SerializerName(id int32) string {
	return serializerNames[id]
}

// SerializerWireType returns the wire type of a serializer by ID, or empty string if unknown.
func SerializerWireType(id int32) string {
	return serializerWireTypes[id]
}

// FormatMetadataValue returns a human-readable string representation of a metadata value.
func FormatMetadataValue(serializerID int32, data []byte) string {
	return FormatMetadataValueIndented(serializerID, data, "")
}

// FormatMetadataValueIndented returns a human-readable string representation of a metadata value with indentation.
func FormatMetadataValueIndented(serializerID int32, data []byte, indent string) string {
	if len(data) == 0 {
		return "<empty>"
	}

	wireType := serializerWireTypes[serializerID]
	buf := ns.NewReader(data)

	switch wireType {
	case "byte":
		if v, err := buf.ReadInt8(); err == nil {
			return fmt.Sprintf("%d", v)
		}

	case "varint":
		if v, err := buf.ReadVarInt(); err == nil {
			return fmt.Sprintf("%d", v)
		}

	case "varlong":
		if v, err := buf.ReadVarLong(); err == nil {
			return fmt.Sprintf("%d", v)
		}

	case "float32":
		if v, err := buf.ReadFloat32(); err == nil {
			return fmt.Sprintf("%g", v)
		}

	case "string":
		if v, err := buf.ReadString(32767); err == nil {
			return fmt.Sprintf("%q", v)
		}

	case "bool":
		if v, err := buf.ReadBool(); err == nil {
			return fmt.Sprintf("%t", v)
		}

	case "rotations":
		x, err1 := buf.ReadFloat32()
		y, err2 := buf.ReadFloat32()
		z, err3 := buf.ReadFloat32()
		if err1 == nil && err2 == nil && err3 == nil {
			return fmt.Sprintf("Rotations{X: %g, Y: %g, Z: %g}", x, y, z)
		}

	case "position":
		if v, err := buf.ReadPosition(); err == nil {
			return fmt.Sprintf("Position{X: %d, Y: %d, Z: %d}", v.X, v.Y, v.Z)
		}

	case "optional_position":
		if present, err := buf.ReadBool(); err == nil {
			if !present {
				return "None"
			}
			if v, err := buf.ReadPosition(); err == nil {
				return fmt.Sprintf("Position{X: %d, Y: %d, Z: %d}", v.X, v.Y, v.Z)
			}
		}

	case "optional_uuid":
		if present, err := buf.ReadBool(); err == nil {
			if !present {
				return "None"
			}
			if v, err := buf.ReadUUID(); err == nil {
				return fmt.Sprintf("%x-%x-%x-%x-%x", v[0:4], v[4:6], v[6:8], v[8:10], v[10:16])
			}
		}

	case "optional_varint":
		if v, err := buf.ReadVarInt(); err == nil {
			if v == 0 {
				return "None"
			}
			return fmt.Sprintf("%d", v-1)
		}

	case "villager_data":
		vType, err1 := buf.ReadVarInt()
		profession, err2 := buf.ReadVarInt()
		level, err3 := buf.ReadVarInt()
		if err1 == nil && err2 == nil && err3 == nil {
			return fmt.Sprintf("VillagerData{Type: %d, Profession: %d, Level: %d}", vType, profession, level)
		}

	case "vector3f":
		x, err1 := buf.ReadFloat32()
		y, err2 := buf.ReadFloat32()
		z, err3 := buf.ReadFloat32()
		if err1 == nil && err2 == nil && err3 == nil {
			return fmt.Sprintf("Vec3{X: %g, Y: %g, Z: %g}", x, y, z)
		}

	case "quaternionf":
		x, err1 := buf.ReadFloat32()
		y, err2 := buf.ReadFloat32()
		z, err3 := buf.ReadFloat32()
		w, err4 := buf.ReadFloat32()
		if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
			return fmt.Sprintf("Quaternion{X: %g, Y: %g, Z: %g, W: %g}", x, y, z, w)
		}

	case "slot":
		// decode slot and format it showing only wire components
		slot, err := buf.ReadSlot(items.Decoder())
		if err != nil {
			return fmt.Sprintf("0x%x (decode error: %v)", data, err)
		}
		return items.FormatSlotForDisplay(slot, indent)

	case "nbt", "optional_nbt", "particle", "particle_list",
		"optional_global_pos", "id_or_inline", "resolvable_profile":
		// complex types - show hex for now
		return fmt.Sprintf("0x%x", data)

	case "empty":
		return "<empty>"
	}

	// fallback: show hex
	return fmt.Sprintf("0x%x", data)
}
