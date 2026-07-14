package net_structures

import (
	"fmt"
)

// HashedSlot represents an item stack where component data is replaced with
// CRC32C hashes. Used in C2S packets (e.g. ContainerClick) where the server
// only needs to verify component identity, not read full data.
//
// Wire format:
//
//	┌───────────────────┬─────────────────┬─────────────────┬───────────────────────────────────────┐
//	│  Present (Bool)   │  ItemID (VarInt)│  Count (VarInt) │  HashedPatchMap                       │
//	└───────────────────┴─────────────────┴─────────────────┴───────────────────────────────────────┘
//
// If Present is false, the slot is empty and no further data is read.
//
// HashedPatchMap:
//
//	┌──────────────────────┬────────────────────────────────┬──────────────────────┬─────────────────────┐
//	│  AddedCount (VarInt) │  Added (VarInt ID + Int32) × N │  RemovedCount (VarInt)│  Removed (VarInt) × M│
//	└──────────────────────┴────────────────────────────────┴──────────────────────┴─────────────────────┘
type HashedSlot struct {
	Present    bool
	ItemID     VarInt
	Count      VarInt
	Components HashedComponents
}

// HashedComponents holds component hashes for a hashed slot.
type HashedComponents struct {
	Add    []HashedComponent // component type ID → CRC32C hash
	Remove []VarInt          // component type IDs to remove
}

// HashedComponent stores a component type ID and its CRC32C hash.
type HashedComponent struct {
	ID   VarInt
	Hash Int32
}

// EmptyHashedSlot returns an empty hashed slot.
func EmptyHashedSlot() HashedSlot {
	return HashedSlot{}
}

// NewHashedSlot creates a hashed slot with the given item and count.
func NewHashedSlot(itemID VarInt, count VarInt) HashedSlot {
	return HashedSlot{
		Present: true,
		ItemID:  itemID,
		Count:   count,
	}
}

// IsEmpty returns true if the hashed slot is empty.
func (s *HashedSlot) IsEmpty() bool {
	return !s.Present
}

// Decode reads a hashed slot from the buffer.
func (s *HashedSlot) Decode(buf *PacketBuffer) error {
	present, err := buf.ReadBool()
	if err != nil {
		return fmt.Errorf("failed to read hashed slot present: %w", err)
	}
	s.Present = bool(present)

	if !s.Present {
		return nil
	}

	if s.ItemID, err = buf.ReadVarInt(); err != nil {
		return fmt.Errorf("failed to read hashed slot item id: %w", err)
	}
	if s.Count, err = buf.ReadVarInt(); err != nil {
		return fmt.Errorf("failed to read hashed slot count: %w", err)
	}

	// HashedPatchMap: added components
	addCount, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read hashed slot add count: %w", err)
	}
	s.Components.Add = make([]HashedComponent, addCount)
	for i := range s.Components.Add {
		if s.Components.Add[i].ID, err = buf.ReadVarInt(); err != nil {
			return fmt.Errorf("failed to read added component %d id: %w", i, err)
		}
		if s.Components.Add[i].Hash, err = buf.ReadInt32(); err != nil {
			return fmt.Errorf("failed to read added component %d hash: %w", i, err)
		}
	}

	// HashedPatchMap: removed components
	removeCount, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read hashed slot remove count: %w", err)
	}
	s.Components.Remove = make([]VarInt, removeCount)
	for i := range s.Components.Remove {
		if s.Components.Remove[i], err = buf.ReadVarInt(); err != nil {
			return fmt.Errorf("failed to read removed component %d id: %w", i, err)
		}
	}

	return nil
}

// Encode writes the hashed slot to the buffer.
func (s *HashedSlot) Encode(buf *PacketBuffer) error {
	if err := buf.WriteBool(Boolean(s.Present)); err != nil {
		return fmt.Errorf("failed to write hashed slot present: %w", err)
	}

	if !s.Present {
		return nil
	}

	if err := buf.WriteVarInt(s.ItemID); err != nil {
		return fmt.Errorf("failed to write hashed slot item id: %w", err)
	}
	if err := buf.WriteVarInt(s.Count); err != nil {
		return fmt.Errorf("failed to write hashed slot count: %w", err)
	}

	// HashedPatchMap: added components
	if err := buf.WriteVarInt(VarInt(len(s.Components.Add))); err != nil {
		return fmt.Errorf("failed to write hashed slot add count: %w", err)
	}
	for i, comp := range s.Components.Add {
		if err := buf.WriteVarInt(comp.ID); err != nil {
			return fmt.Errorf("failed to write added component %d id: %w", i, err)
		}
		if err := buf.WriteInt32(comp.Hash); err != nil {
			return fmt.Errorf("failed to write added component %d hash: %w", i, err)
		}
	}

	// HashedPatchMap: removed components
	if err := buf.WriteVarInt(VarInt(len(s.Components.Remove))); err != nil {
		return fmt.Errorf("failed to write hashed slot remove count: %w", err)
	}
	for i, id := range s.Components.Remove {
		if err := buf.WriteVarInt(id); err != nil {
			return fmt.Errorf("failed to write removed component %d id: %w", i, err)
		}
	}

	return nil
}

// ReadHashedSlot reads a hashed slot from the buffer.
func (pb *PacketBuffer) ReadHashedSlot() (HashedSlot, error) {
	var slot HashedSlot
	err := slot.Decode(pb)
	return slot, err
}

// WriteHashedSlot writes a hashed slot to the buffer.
func (pb *PacketBuffer) WriteHashedSlot(s HashedSlot) error {
	return s.Encode(pb)
}
