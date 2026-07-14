package storage

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zeozeozeo/minego/internal/data/items"
	"github.com/zeozeozeo/minego/internal/data/registries"
	"github.com/zeozeozeo/minego/internal/protocol/nbt"
)

// PlayerData holds the data persisted for a player.
type PlayerData struct {
	X, Y, Z    float64
	Yaw, Pitch float32
	Dimension  string
	Gamemode   int32
	Inventory  []InventorySlot
	HeldSlot   int32
}

// InventorySlot represents one item in a player's inventory.
type InventorySlot struct {
	Slot  int8
	ID    string // e.g. "minecraft:stone"
	Count int32
}

// SavePlayer writes player data to <dir>/<uuid>.dat (gzip-compressed NBT).
func SavePlayer(dir, uuid string, pd *PlayerData) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	root := nbt.Compound{
		"DataVersion":      nbt.Int(DataVersion),
		"playerGameType":   nbt.Int(pd.Gamemode),
		"Dimension":        nbt.String(pd.Dimension),
		"SelectedItemSlot": nbt.Int(pd.HeldSlot),
		"Pos": nbt.List{
			ElementType: nbt.TagDouble,
			Elements:    []nbt.Tag{nbt.Double(pd.X), nbt.Double(pd.Y), nbt.Double(pd.Z)},
		},
		"Rotation": nbt.List{
			ElementType: nbt.TagFloat,
			Elements:    []nbt.Tag{nbt.Float(pd.Yaw), nbt.Float(pd.Pitch)},
		},
	}

	// inventory
	invElements := make([]nbt.Tag, len(pd.Inventory))
	for i, slot := range pd.Inventory {
		itemID := registries.Item.Get(slot.ID)
		if itemID < 0 {
			continue
		}
		comp := nbt.Compound{
			"Slot":  nbt.Byte(slot.Slot),
			"id":    nbt.String(slot.ID),
			"count": nbt.Int(slot.Count),
		}

		// encode default item components
		stack := items.NewStack(int32(itemID), slot.Count)
		compTag, err := nbt.MarshalTag(stack.Components)
		if err == nil {
			if compCompound, ok := compTag.(nbt.Compound); ok && len(compCompound) > 0 {
				comp["components"] = compCompound
			}
		}

		invElements[i] = comp
	}
	// filter out nil entries
	var filteredInv []nbt.Tag
	for _, e := range invElements {
		if e != nil {
			filteredInv = append(filteredInv, e)
		}
	}
	root["Inventory"] = nbt.List{ElementType: nbt.TagCompound, Elements: filteredInv}

	data, err := nbt.EncodeFile(root, "")
	if err != nil {
		return fmt.Errorf("encode NBT: %w", err)
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return fmt.Errorf("gzip write: %w", err)
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("gzip close: %w", err)
	}

	path := filepath.Join(dir, uuid+".dat")
	return os.WriteFile(path, buf.Bytes(), 0644)
}

// LoadPlayer reads player data from <dir>/<uuid>.dat. Returns nil, nil if file doesn't exist.
func LoadPlayer(dir, uuid string) (*PlayerData, error) {
	path := filepath.Join(dir, uuid+".dat")
	compressed, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	gr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()

	data, err := io.ReadAll(gr)
	if err != nil {
		return nil, fmt.Errorf("gzip read: %w", err)
	}

	tag, _, err := nbt.DecodeFile(data, nbt.WithMaxBytes(0))
	if err != nil {
		return nil, fmt.Errorf("decode NBT: %w", err)
	}

	root, ok := tag.(nbt.Compound)
	if !ok {
		return nil, fmt.Errorf("root is not compound")
	}

	pd := &PlayerData{
		Gamemode:  root.GetInt("playerGameType"),
		Dimension: root.GetString("Dimension"),
		HeldSlot:  root.GetInt("SelectedItemSlot"),
	}

	// position
	posList := root.GetList("Pos")
	if posList.Len() >= 3 {
		if d, ok := posList.Get(0).(nbt.Double); ok {
			pd.X = float64(d)
		}
		if d, ok := posList.Get(1).(nbt.Double); ok {
			pd.Y = float64(d)
		}
		if d, ok := posList.Get(2).(nbt.Double); ok {
			pd.Z = float64(d)
		}
	}

	// rotation
	rotList := root.GetList("Rotation")
	if rotList.Len() >= 2 {
		if f, ok := rotList.Get(0).(nbt.Float); ok {
			pd.Yaw = float32(f)
		}
		if f, ok := rotList.Get(1).(nbt.Float); ok {
			pd.Pitch = float32(f)
		}
	}

	// inventory
	invList := root.GetList("Inventory")
	for _, elem := range invList.Elements {
		slotComp, ok := elem.(nbt.Compound)
		if !ok {
			continue
		}
		pd.Inventory = append(pd.Inventory, InventorySlot{
			Slot:  slotComp.GetByte("Slot"),
			ID:    slotComp.GetString("id"),
			Count: slotComp.GetInt("count"),
		})
	}

	return pd, nil
}
