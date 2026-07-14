package items

import (
	"fmt"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// ItemStack represents a fully decoded item stack with typed components.
// It acts as middleware over net_structures.Slot, encoding components
// as a DataComponentPatch (adds/removes) relative to the item's defaults.
type ItemStack struct {
	ID         int32
	Count      int32
	Components *Components
}

// EmptyStack returns an empty item stack.
func EmptyStack() *ItemStack {
	return &ItemStack{}
}

// NewStack creates a new item stack with the given item ID and count.
// Components are initialized to the item's defaults, all marked as present.
func NewStack(itemID int32, count int32) *ItemStack {
	components := DefaultComponents(itemID).Clone()
	markAllPresent(components)
	return &ItemStack{
		ID:         itemID,
		Count:      count,
		Components: components,
	}
}

// NewStackWithComponents creates a new item stack with the given item ID, count, and components.
// Only non-zero fields in the provided Components are marked as present and
// encoded on the wire; the item's default components are not included.
func NewStackWithComponents(itemID int32, count int32, components *Components) *ItemStack {
	detectPresent(components)
	return &ItemStack{
		ID:         itemID,
		Count:      count,
		Components: components,
	}
}

// SetDefaultComponents overwrites the stack's components with the item's
// defaults and marks all of them as present.
func (s *ItemStack) SetDefaultComponents() {
	defaults := DefaultComponents(s.ID)
	if defaults != nil {
		s.Components = defaults.Clone()
	} else {
		s.Components = &Components{}
	}
	markAllPresent(s.Components)
}

// IsEmpty returns true if the stack is empty.
func (s *ItemStack) IsEmpty() bool {
	return s == nil || s.Count <= 0
}

// FromSlot creates an ItemStack from a raw net_structures.Slot.
// It applies the slot's component patch on top of the item's defaults.
// All default and patched components are marked as present.
func FromSlot(slot ns.Slot) (*ItemStack, error) {
	if slot.IsEmpty() {
		return EmptyStack(), nil
	}

	defaults := DefaultComponents(int32(slot.ItemID))
	components := defaults.Clone()
	markAllPresent(components)

	// apply added components
	for _, raw := range slot.Components.Add {
		if err := applyComponent(components, int32(raw.ID), raw.Data); err != nil {
			return nil, fmt.Errorf("component %d: %w", raw.ID, err)
		}
		components.SetPresent(int32(raw.ID))
	}

	// apply removals (set to zero values, keep present so they encode as removes)
	for _, id := range slot.Components.Remove {
		clearComponent(components, int32(id))
		components.SetPresent(int32(id))
	}

	return &ItemStack{
		ID:         int32(slot.ItemID),
		Count:      int32(slot.Count),
		Components: components,
	}, nil
}

// ToSlot converts the ItemStack back to a raw net_structures.Slot.
// Only present components that differ from the item's defaults are encoded,
// mirroring Java's DataComponentPatch serialization.
func (s *ItemStack) ToSlot() (ns.Slot, error) {
	if s.IsEmpty() {
		return ns.EmptySlot(), nil
	}

	slot := ns.NewSlot(ns.VarInt(s.ID), ns.VarInt(s.Count))
	defaults := DefaultComponents(s.ID)

	for id := int32(0); id <= MaxComponentID; id++ {
		if !s.Components.HasComponent(id) {
			continue
		}
		differs, hv := componentDiffers(s.Components, defaults, id)
		if !differs {
			continue
		}
		if hv {
			data, err := encodeComponent(s.Components, id)
			if err != nil {
				return ns.Slot{}, fmt.Errorf("encode component %d: %w", id, err)
			}
			slot.AddComponent(ns.VarInt(id), data)
		} else {
			slot.RemoveComponent(ns.VarInt(id))
		}
	}

	return slot, nil
}

// markAllPresent marks all registered component IDs as present.
func markAllPresent(c *Components) {
	for id := int32(0); id <= MaxComponentID; id++ {
		if componentCodecs[id] != nil {
			c.SetPresent(id)
		}
	}
}

// detectPresent scans a Components struct and marks any non-zero fields as present.
// Used for sparse structs where only some fields are intentionally set.
func detectPresent(c *Components) {
	zero := &Components{}
	for id := int32(0); id <= MaxComponentID; id++ {
		codec := componentCodecs[id]
		if codec == nil {
			continue
		}
		differs, _ := codec.Differs(c, zero)
		if differs {
			c.SetPresent(id)
		}
	}
}

// Decoder returns a SlotDecoder function that can be passed to Slot.Decode.
// This reads component data from the wire format.
func Decoder() ns.SlotDecoder {
	return decodeComponentWire
}

// DecoderDelimited returns a SlotDecoder that handles length-prefixed
// component data, as used by OPTIONAL_UNTRUSTED_STREAM_CODEC (e.g. creative mode slots).
func DecoderDelimited() ns.SlotDecoder {
	return func(buf *ns.PacketBuffer, id ns.VarInt) ([]byte, error) {
		// read length prefix
		length, err := buf.ReadVarInt()
		if err != nil {
			return nil, fmt.Errorf("failed to read component length: %w", err)
		}

		if length == 0 {
			// empty component (e.g. Unbreakable)
			return nil, nil
		}

		// read exactly 'length' bytes
		rawData, err := buf.ReadFixedByteArray(int(length))
		if err != nil {
			return nil, fmt.Errorf("failed to read component data: %w", err)
		}

		// decode from the raw bytes
		limitedBuf := ns.NewReader(rawData)
		return decodeComponentWire(limitedBuf, id)
	}
}

// ReadSlot is a convenience function that reads a Slot from the buffer
// and converts it to an ItemStack.
func ReadSlot(buf *ns.PacketBuffer) (*ItemStack, error) {
	slot, err := buf.ReadSlot(Decoder())
	if err != nil {
		return nil, err
	}
	return FromSlot(slot)
}

// ReadSlotDelimited reads a slot with length-prefixed component data.
// Used for packets with OPTIONAL_UNTRUSTED_STREAM_CODEC like creative mode.
func ReadSlotDelimited(buf *ns.PacketBuffer) (*ItemStack, error) {
	slot, err := buf.ReadSlot(DecoderDelimited())
	if err != nil {
		return nil, err
	}
	return FromSlot(slot)
}

// WriteSlot writes the ItemStack to the buffer as a Slot.
func (s *ItemStack) WriteSlot(buf *ns.PacketBuffer) error {
	slot, err := s.ToSlot()
	if err != nil {
		return err
	}
	return buf.WriteSlot(slot)
}

// WriteSlotDelimited writes the ItemStack with length-prefixed component data.
// Used for packets with OPTIONAL_UNTRUSTED_STREAM_CODEC like creative mode.
func (s *ItemStack) WriteSlotDelimited(buf *ns.PacketBuffer) error {
	slot, err := s.ToSlot()
	if err != nil {
		return err
	}
	return writeSlotDelimited(buf, slot)
}

// WriteRawSlotDelimited writes a raw ns.Slot with length-prefixed component data.
// Used for packets with OPTIONAL_UNTRUSTED_STREAM_CODEC like creative mode.
func WriteRawSlotDelimited(buf *ns.PacketBuffer, slot ns.Slot) error {
	return writeSlotDelimited(buf, slot)
}

// writeSlotDelimited writes a slot with length-prefixed component data.
func writeSlotDelimited(buf *ns.PacketBuffer, slot ns.Slot) error {
	// write count
	if err := buf.WriteVarInt(slot.Count); err != nil {
		return err
	}

	if slot.Count <= 0 {
		return nil
	}

	// write item ID
	if err := buf.WriteVarInt(slot.ItemID); err != nil {
		return err
	}

	// write add count and remove count
	if err := buf.WriteVarInt(ns.VarInt(len(slot.Components.Add))); err != nil {
		return err
	}
	if err := buf.WriteVarInt(ns.VarInt(len(slot.Components.Remove))); err != nil {
		return err
	}

	// write components with length prefix
	for _, comp := range slot.Components.Add {
		// write component ID
		if err := buf.WriteVarInt(comp.ID); err != nil {
			return err
		}
		// write component data length
		if err := buf.WriteVarInt(ns.VarInt(len(comp.Data))); err != nil {
			return err
		}
		// write component data
		if _, err := buf.Write(comp.Data); err != nil {
			return err
		}
	}

	// write removed component IDs (no length prefix for these)
	for _, id := range slot.Components.Remove {
		if err := buf.WriteVarInt(id); err != nil {
			return err
		}
	}

	return nil
}
