package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// testSlotDecoder decodes known simple component types for testing.
func testSlotDecoder(buf *ns.PacketBuffer, id ns.VarInt) ([]byte, error) {
	w := ns.NewWriter()
	switch id {
	case 1: // max stack size - VarInt
		v, err := buf.ReadVarInt()
		if err != nil {
			return nil, err
		}
		w.WriteVarInt(v)
	case 3: // damage - VarInt
		v, err := buf.ReadVarInt()
		if err != nil {
			return nil, err
		}
		w.WriteVarInt(v)
	case 4: // unbreakable - Boolean
		v, err := buf.ReadBool()
		if err != nil {
			return nil, err
		}
		w.WriteBool(v)
	default:
		return nil, nil
	}
	return w.Bytes(), nil
}

// Slot wire format:
//   VarInt count (0 = empty, no further data)
//   VarInt itemID
//   VarInt addCount
//   VarInt removeCount
//   [addCount × (VarInt componentID + component data)]
//   [removeCount × VarInt componentID]

var slotTestCases = []struct {
	name   string
	raw    []byte
	count  ns.VarInt
	itemID ns.VarInt
	add    []ns.RawSlotComponent
	remove []ns.VarInt
}{
	{
		name:  "empty slot",
		raw:   []byte{0x00},
		count: 0,
	},
	{
		name:   "stone x64 no components",
		raw:    []byte{0x40, 0x01, 0x00, 0x00},
		count:  64,
		itemID: 1,
	},
	{
		name:   "diamond x1 no components",
		raw:    []byte{0x01, 0x88, 0x02, 0x00, 0x00},
		count:  1,
		itemID: 264,
	},
	{
		name:   "item with one removed component",
		raw:    []byte{0x01, 0x01, 0x00, 0x01, 0x03},
		count:  1,
		itemID: 1,
		remove: []ns.VarInt{3},
	},
	{
		name:   "item with two removed components",
		raw:    []byte{0x01, 0x01, 0x00, 0x02, 0x03, 0x0c},
		count:  1,
		itemID: 1,
		remove: []ns.VarInt{3, 12},
	},
	{
		name:   "item with damage component",
		raw:    []byte{0x01, 0x64, 0x01, 0x00, 0x03, 0x32},
		count:  1,
		itemID: 100,
		add:    []ns.RawSlotComponent{{ID: 3, Data: []byte{0x32}}},
	},
	{
		name:   "item with max stack size component",
		raw:    []byte{0x10, 0x32, 0x01, 0x00, 0x01, 0x10},
		count:  16,
		itemID: 50,
		add:    []ns.RawSlotComponent{{ID: 1, Data: []byte{0x10}}},
	},
	{
		name:   "item with multiple components",
		raw:    []byte{0x01, 0x64, 0x02, 0x01, 0x03, 0x19, 0x01, 0x01, 0x04},
		count:  1,
		itemID: 100,
		add:    []ns.RawSlotComponent{{ID: 3, Data: []byte{0x19}}, {ID: 1, Data: []byte{0x01}}},
		remove: []ns.VarInt{4},
	},
}

func TestSlot(t *testing.T) {
	for _, tc := range slotTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			buf := ns.NewReader(tc.raw)
			got, err := buf.ReadSlot(testSlotDecoder)
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got.Count != tc.count || got.ItemID != tc.itemID {
				t.Errorf("basic mismatch: got count=%d itemID=%d, want count=%d itemID=%d",
					got.Count, got.ItemID, tc.count, tc.itemID)
			}
			if !slotComponentsEqual(got.Components.Add, tc.add) {
				t.Errorf("Add components mismatch")
			}
			if !slotRemoveEqual(got.Components.Remove, tc.remove) {
				t.Errorf("Remove components mismatch")
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			slot := ns.Slot{
				Count:  tc.count,
				ItemID: tc.itemID,
				Components: ns.SlotComponents{
					Add:    tc.add,
					Remove: tc.remove,
				},
			}
			buf := ns.NewWriter()
			if err := slot.Encode(buf); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

func slotComponentsEqual(a, b []ns.RawSlotComponent) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID || !bytes.Equal(a[i].Data, b[i].Data) {
			return false
		}
	}
	return true
}

func slotRemoveEqual(a, b []ns.VarInt) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSlot_GetComponent(t *testing.T) {
	slot := ns.NewSlot(100, 1)
	slot.AddComponent(3, []byte{0x32})
	slot.AddComponent(5, []byte{0x01, 0x02})

	if comp := slot.GetComponent(3); comp == nil || comp.ID != 3 {
		t.Error("GetComponent(3) failed")
	}
	if comp := slot.GetComponent(999); comp != nil {
		t.Error("GetComponent(999) should return nil")
	}
}
