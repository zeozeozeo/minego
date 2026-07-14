package net_structures

import (
	"fmt"
)

// Slot represents an item stack with data components.
// Components are stored as raw bytes to keep this package protocol-level only.
// Callers should use a higher-level package to parse specific component types.
//
// Wire format:
//
//	┌──────────────────┬─────────────────┬─────────────────┬─────────────────┬──────────────────────────────────┐
//	│  Count (VarInt)  │  ItemID (VarInt)│  Add (VarInt)   │  Remove (VarInt)│  Components...                   │
//	└──────────────────┴─────────────────┴─────────────────┴─────────────────┴──────────────────────────────────┘
//
// If Count <= 0, the slot is empty and no further data is read.
// ItemID is the registry ID from minecraft:item.
// Add count is the number of components to add (with data).
// Remove count is the number of component type IDs to remove.
type Slot struct {
	Count      VarInt
	ItemID     VarInt         // only if Count > 0
	Components SlotComponents // only if Count > 0
}

// SlotComponents holds the component modifications for a slot.
type SlotComponents struct {
	Add    []RawSlotComponent // components with data (ID + raw bytes)
	Remove []VarInt           // component type IDs to remove
}

// RawSlotComponent stores a component as ID + raw bytes.
// This allows passthrough without parsing component internals.
type RawSlotComponent struct {
	ID   VarInt
	Data []byte
}

// EmptySlot returns an empty slot.
func EmptySlot() Slot {
	return Slot{Count: 0}
}

// NewSlot creates a slot with the given item and count.
func NewSlot(itemID VarInt, count VarInt) Slot {
	return Slot{
		Count:  count,
		ItemID: itemID,
	}
}

// IsEmpty returns true if the slot is empty.
func (s *Slot) IsEmpty() bool {
	return s.Count <= 0
}

// SlotDecoder is a function that decodes a component's data given its ID.
// Returns the raw bytes of the component. If the component format is unknown,
// return an error - there's no way to know where a component ends without
// understanding its format.
type SlotDecoder func(buf *PacketBuffer, componentID VarInt) ([]byte, error)

// SlotEncoder is a function that encodes a component's data.
// By default, raw bytes are written as-is.
type SlotEncoder func(buf *PacketBuffer, componentID VarInt, data []byte) error

// defaultSlotEncoder writes raw bytes as-is.
func defaultSlotEncoder(buf *PacketBuffer, _ VarInt, data []byte) error {
	_, err := buf.Write(data)
	return err
}

// Encode writes the slot to the buffer.
func (s *Slot) Encode(buf *PacketBuffer) error {
	return s.EncodeWith(buf, defaultSlotEncoder)
}

// EncodeWith writes the slot using a custom encoder for components.
func (s *Slot) EncodeWith(buf *PacketBuffer, encode SlotEncoder) error {
	if err := buf.WriteVarInt(s.Count); err != nil {
		return fmt.Errorf("failed to write slot count: %w", err)
	}

	if s.Count <= 0 {
		return nil
	}

	if err := buf.WriteVarInt(s.ItemID); err != nil {
		return fmt.Errorf("failed to write slot item id: %w", err)
	}

	// write add count
	if err := buf.WriteVarInt(VarInt(len(s.Components.Add))); err != nil {
		return fmt.Errorf("failed to write slot add count: %w", err)
	}

	// write remove count
	if err := buf.WriteVarInt(VarInt(len(s.Components.Remove))); err != nil {
		return fmt.Errorf("failed to write slot remove count: %w", err)
	}

	// write added components
	for i, comp := range s.Components.Add {
		if err := buf.WriteVarInt(comp.ID); err != nil {
			return fmt.Errorf("failed to write component %d id: %w", i, err)
		}
		if err := encode(buf, comp.ID, comp.Data); err != nil {
			return fmt.Errorf("failed to write component %d data: %w", i, err)
		}
	}

	// write removed component IDs
	for i, id := range s.Components.Remove {
		if err := buf.WriteVarInt(id); err != nil {
			return fmt.Errorf("failed to write removed component %d id: %w", i, err)
		}
	}

	return nil
}

// Decode reads a slot from the buffer using a decoder that knows component sizes.
// The decoder must return the raw bytes for each component.
func (s *Slot) Decode(buf *PacketBuffer, decode SlotDecoder) error {
	count, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read slot count: %w", err)
	}
	s.Count = count

	if s.Count <= 0 {
		return nil
	}

	s.ItemID, err = buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read slot item id: %w", err)
	}

	addCount, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read slot add count: %w", err)
	}

	removeCount, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read slot remove count: %w", err)
	}

	// read added components
	s.Components.Add = make([]RawSlotComponent, addCount)
	for i := range s.Components.Add {
		compID, err := buf.ReadVarInt()
		if err != nil {
			return fmt.Errorf("failed to read component %d id: %w", i, err)
		}

		data, err := decode(buf, compID)
		if err != nil {
			return fmt.Errorf("failed to read component %d (id=%d): %w", i, compID, err)
		}
		s.Components.Add[i] = RawSlotComponent{ID: compID, Data: data}
	}

	// read removed component IDs
	s.Components.Remove = make([]VarInt, removeCount)
	for i := range s.Components.Remove {
		s.Components.Remove[i], err = buf.ReadVarInt()
		if err != nil {
			return fmt.Errorf("failed to read removed component %d id: %w", i, err)
		}
	}

	return nil
}

// ReadSlot reads a slot from the buffer using the provided decoder.
func (pb *PacketBuffer) ReadSlot(decode SlotDecoder) (Slot, error) {
	var slot Slot
	err := slot.Decode(pb, decode)
	return slot, err
}

// WriteSlot writes a slot to the buffer.
func (pb *PacketBuffer) WriteSlot(s Slot) error {
	return s.Encode(pb)
}

// WriteSlotWith writes a slot using a custom encoder.
func (pb *PacketBuffer) WriteSlotWith(s Slot, encode SlotEncoder) error {
	return s.EncodeWith(pb, encode)
}

// GetComponent returns the first component with the given ID, or nil if not found.
func (s *Slot) GetComponent(id VarInt) *RawSlotComponent {
	for i := range s.Components.Add {
		if s.Components.Add[i].ID == id {
			return &s.Components.Add[i]
		}
	}
	return nil
}

// AddComponent adds a raw component to the slot.
func (s *Slot) AddComponent(id VarInt, data []byte) {
	s.Components.Add = append(s.Components.Add, RawSlotComponent{ID: id, Data: data})
}

// RemoveComponent marks a component type for removal.
func (s *Slot) RemoveComponent(id VarInt) {
	s.Components.Remove = append(s.Components.Remove, id)
}

// CopySlot copies a slot from src to this buffer.
// This only works for empty slots or slots without component modifications.
// For slots with components, use ReadSlot with a decoder and WriteSlot.
func (pb *PacketBuffer) CopySlot(src *PacketBuffer) error {
	count, err := src.ReadVarInt()
	if err != nil {
		return err
	}
	if err := pb.WriteVarInt(count); err != nil {
		return err
	}

	if count <= 0 {
		return nil
	}

	// item ID
	if err := pb.CopyVarInt(src); err != nil {
		return err
	}

	// add count
	addCount, err := src.ReadVarInt()
	if err != nil {
		return err
	}
	if err := pb.WriteVarInt(addCount); err != nil {
		return err
	}

	// remove count
	removeCount, err := src.ReadVarInt()
	if err != nil {
		return err
	}
	if err := pb.WriteVarInt(removeCount); err != nil {
		return err
	}

	// component data requires a decoder to know sizes
	if addCount > 0 || removeCount > 0 {
		return fmt.Errorf("cannot copy slot with components without decoder; use ReadSlot/WriteSlot")
	}

	return nil
}
