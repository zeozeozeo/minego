package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// HashedSlot wire format:
//   Bool present (false = empty, no further data)
//   VarInt itemID
//   VarInt count
//   HashedPatchMap:
//     VarInt addCount
//     [addCount × (VarInt componentTypeID + Int32 CRC32C hash)]
//     VarInt removeCount
//     [removeCount × VarInt componentTypeID]

var hashedSlotTestCases = []struct {
	name    string
	raw     []byte
	present bool
	itemID  ns.VarInt
	count   ns.VarInt
	add     []ns.HashedComponent
	remove  []ns.VarInt
}{
	{
		name: "empty",
		raw:  []byte{0x00},
	},
	{
		name:    "item with no component changes",
		raw:     []byte{0x01, 0x86, 0x01, 0x40, 0x00, 0x00},
		present: true,
		itemID:  134,
		count:   64,
	},
	{
		name:    "item with one added component hash",
		raw:     []byte{0x01, 0x24, 0x04, 0x01, 0x03, 0x12, 0x34, 0x56, 0x78, 0x00},
		present: true,
		itemID:  36,
		count:   4,
		add:     []ns.HashedComponent{{ID: 3, Hash: 0x12345678}},
	},
	{
		name:    "item with one removed component",
		raw:     []byte{0x01, 0x01, 0x01, 0x00, 0x01, 0x04},
		present: true,
		itemID:  1,
		count:   1,
		remove:  []ns.VarInt{4},
	},
	{
		name:    "item with added and removed components",
		raw:     []byte{0x01, 0x64, 0x01, 0x02, 0x03, 0xAA, 0xBB, 0xCC, 0xDD, 0x05, 0xFF, 0xEE, 0xDD, 0xCC, 0x01, 0x10},
		present: true,
		itemID:  100,
		count:   1,
		add: []ns.HashedComponent{
			{ID: 3, Hash: -0x55443323}, // 0xAABBCCDD as signed
			{ID: 5, Hash: -0x00112234}, // 0xFFEEDDCC as signed
		},
		remove: []ns.VarInt{16},
	},
}

func TestHashedSlot(t *testing.T) {
	for _, tc := range hashedSlotTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			buf := ns.NewReader(tc.raw)
			got, err := buf.ReadHashedSlot()
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got.Present != tc.present || got.ItemID != tc.itemID || got.Count != tc.count {
				t.Errorf("basic mismatch: got present=%v itemID=%d count=%d, want present=%v itemID=%d count=%d",
					got.Present, got.ItemID, got.Count, tc.present, tc.itemID, tc.count)
			}
			if !hashedComponentsEqual(got.Components.Add, tc.add) {
				t.Errorf("Add components mismatch: got %+v, want %+v", got.Components.Add, tc.add)
			}
			if !slotRemoveEqual(got.Components.Remove, tc.remove) {
				t.Errorf("Remove components mismatch: got %v, want %v", got.Components.Remove, tc.remove)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			slot := ns.HashedSlot{
				Present: tc.present,
				ItemID:  tc.itemID,
				Count:   tc.count,
				Components: ns.HashedComponents{
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

func hashedComponentsEqual(a, b []ns.HashedComponent) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID || a[i].Hash != b[i].Hash {
			return false
		}
	}
	return true
}
