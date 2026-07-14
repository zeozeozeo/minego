package items

// Component codec registry - provides a unified interface for encoding/decoding item components.
//
// Each component type implements ComponentCodec, which handles:
// - Wire format decoding (network bytes → raw bytes for storage)
// - Applying raw bytes to the Components struct
// - Clearing a component from the struct
// - Checking if a component differs from defaults
// - Encoding from the Components struct back to raw bytes

import (
	"fmt"
	"slices"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"github.com/zeozeozeo/minego/internal/protocol/nbt"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/registries"
)

// ComponentCodec defines the interface for encoding/decoding a component type.
type ComponentCodec interface {
	// DecodeWire reads the component from wire format and returns raw bytes for storage.
	DecodeWire(buf *ns.PacketBuffer) ([]byte, error)

	// Apply applies raw bytes to the Components struct.
	Apply(c *Components, data []byte) error

	// Clear clears this component from the struct (sets to zero value).
	Clear(c *Components)

	// Differs returns whether the component differs from defaults.
	// Returns (differs, hasValue) - hasValue is false if component should be removed.
	Differs(c, defaults *Components) (bool, bool)

	// Encode encodes the component from the struct to raw bytes.
	Encode(c *Components) ([]byte, error)
}

// componentCodecs maps component IDs to their codecs.
var componentCodecs = make(map[int32]ComponentCodec)

const maxStringLen = 32767

// RegisterCodec registers a codec for a component ID.
func RegisterCodec(id int32, codec ComponentCodec) {
	componentCodecs[id] = codec
}

// GetCodec returns the codec for a component ID, or nil if not found.
func GetCodec(id int32) ComponentCodec {
	return componentCodecs[id]
}

// decodeComponentWire decodes a component using the registry.
func decodeComponentWire(buf *ns.PacketBuffer, id ns.VarInt) ([]byte, error) {
	codec := componentCodecs[int32(id)]
	if codec == nil {
		return nil, fmt.Errorf("unknown component ID %d", id)
	}
	return codec.DecodeWire(buf)
}

// applyComponent applies a component using the registry.
func applyComponent(c *Components, id int32, data []byte) error {
	codec := componentCodecs[id]
	if codec == nil {
		return fmt.Errorf("unknown component ID %d", id)
	}
	return codec.Apply(c, data)
}

// clearComponent clears a component using the registry.
func clearComponent(c *Components, id int32) {
	codec := componentCodecs[id]
	if codec != nil {
		codec.Clear(c)
	}
}

// componentDiffers checks if a component differs using the registry.
func componentDiffers(c, defaults *Components, id int32) (bool, bool) {
	codec := componentCodecs[id]
	if codec == nil {
		return false, false
	}
	return codec.Differs(c, defaults)
}

// encodeComponent encodes a component using the registry.
func encodeComponent(c *Components, id int32) ([]byte, error) {
	codec := componentCodecs[id]
	if codec == nil {
		return nil, fmt.Errorf("cannot encode component %d", id)
	}
	return codec.Encode(c)
}

// ============================================================================
// Helper base types for common patterns
// ============================================================================

// varIntCodec handles simple VarInt components.
type varIntCodec struct {
	get func(c *Components) int32
	set func(c *Components, v int32)
}

func (codec *varIntCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	v, err := buf.ReadVarInt()
	if err != nil {
		return nil, err
	}
	w := ns.NewWriter()
	w.WriteVarInt(v)
	return w.Bytes(), nil
}

func (codec *varIntCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	v, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	codec.set(c, int32(v))
	return nil
}

func (codec *varIntCodec) Clear(c *Components) {
	codec.set(c, 0)
}

func (codec *varIntCodec) Differs(c, defaults *Components) (bool, bool) {
	cv := codec.get(c)
	dv := codec.get(defaults)
	return cv != dv, cv != 0
}

func (codec *varIntCodec) Encode(c *Components) ([]byte, error) {
	w := ns.NewWriter()
	w.WriteVarInt(ns.VarInt(codec.get(c)))
	return w.Bytes(), nil
}

// float32Codec handles simple Float32 components.
type float32Codec struct {
	get func(c *Components) float64
	set func(c *Components, v float64)
}

func (codec *float32Codec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	v, err := buf.ReadFloat32()
	if err != nil {
		return nil, err
	}
	w := ns.NewWriter()
	w.WriteFloat32(v)
	return w.Bytes(), nil
}

func (codec *float32Codec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	v, err := buf.ReadFloat32()
	if err != nil {
		return err
	}
	codec.set(c, float64(v))
	return nil
}

func (codec *float32Codec) Clear(c *Components) {
	codec.set(c, 0)
}

func (codec *float32Codec) Differs(c, defaults *Components) (bool, bool) {
	cv := codec.get(c)
	dv := codec.get(defaults)
	return cv != dv, cv != 0
}

func (codec *float32Codec) Encode(c *Components) ([]byte, error) {
	w := ns.NewWriter()
	w.WriteFloat32(ns.Float32(codec.get(c)))
	return w.Bytes(), nil
}

// stringCodec handles simple string/identifier components.
type stringCodec struct {
	get func(c *Components) string
	set func(c *Components, v string)
}

func (codec *stringCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	v, err := buf.ReadString(maxStringLen)
	if err != nil {
		return nil, err
	}
	w := ns.NewWriter()
	w.WriteString(v)
	return w.Bytes(), nil
}

func (codec *stringCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	v, err := buf.ReadString(maxStringLen)
	if err != nil {
		return err
	}
	codec.set(c, string(v))
	return nil
}

func (codec *stringCodec) Clear(c *Components) {
	codec.set(c, "")
}

func (codec *stringCodec) Differs(c, defaults *Components) (bool, bool) {
	cv := codec.get(c)
	dv := codec.get(defaults)
	return cv != dv, cv != ""
}

func (codec *stringCodec) Encode(c *Components) ([]byte, error) {
	w := ns.NewWriter()
	w.WriteString(ns.String(codec.get(c)))
	return w.Bytes(), nil
}

// emptyMarkerCodec handles empty marker components (bool flags with no wire data).
type emptyMarkerCodec struct {
	get func(c *Components) bool
	set func(c *Components, v bool)
}

func (codec *emptyMarkerCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	return nil, nil // no data
}

func (codec *emptyMarkerCodec) Apply(c *Components, data []byte) error {
	codec.set(c, true)
	return nil
}

func (codec *emptyMarkerCodec) Clear(c *Components) {
	codec.set(c, false)
}

func (codec *emptyMarkerCodec) Differs(c, defaults *Components) (bool, bool) {
	cv := codec.get(c)
	dv := codec.get(defaults)
	return cv != dv, cv
}

func (codec *emptyMarkerCodec) Encode(c *Components) ([]byte, error) {
	return nil, nil // no data
}

// ============================================================================
// Component-specific codecs
// ============================================================================

// customNameCodec handles CustomName (NBT text component).
type customNameCodec struct{}

func (codec *customNameCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	if err := copyNBT(buf, w); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (codec *customNameCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	name, err := decodeItemName(buf)
	if err != nil {
		return err
	}
	c.CustomName = name
	return nil
}

func (codec *customNameCodec) Clear(c *Components) {
	c.CustomName = nil
}

func (codec *customNameCodec) Differs(c, defaults *Components) (bool, bool) {
	cHas := c.CustomName != nil
	dHas := defaults.CustomName != nil
	if cHas != dHas {
		return true, cHas
	}
	if cHas && dHas {
		return *c.CustomName != *defaults.CustomName, true
	}
	return false, false
}

func (codec *customNameCodec) Encode(c *Components) ([]byte, error) {
	if c.CustomName == nil {
		return nil, nil
	}
	w := ns.NewWriter()
	if err := encodeItemName(w, c.CustomName); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// itemNameCodec handles ItemName (NBT text component).
type itemNameCodec struct{}

func (codec *itemNameCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	if err := copyNBT(buf, w); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (codec *itemNameCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	name, err := decodeItemName(buf)
	if err != nil {
		return err
	}
	c.ItemName = name
	return nil
}

func (codec *itemNameCodec) Clear(c *Components) {
	c.ItemName = nil
}

func (codec *itemNameCodec) Differs(c, defaults *Components) (bool, bool) {
	cHas := c.ItemName != nil
	dHas := defaults.ItemName != nil
	if cHas != dHas {
		return true, cHas
	}
	if cHas && dHas {
		return *c.ItemName != *defaults.ItemName, true
	}
	return false, false
}

func (codec *itemNameCodec) Encode(c *Components) ([]byte, error) {
	if c.ItemName == nil {
		return nil, nil
	}
	w := ns.NewWriter()
	if err := encodeItemName(w, c.ItemName); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// attributeModifiersCodec handles AttributeModifiers.
type attributeModifiersCodec struct{}

func (codec *attributeModifiersCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	count, err := buf.ReadVarInt()
	if err != nil {
		return nil, err
	}
	w.WriteVarInt(count)
	for range int(count) {
		if err := copyAttributeModifier(buf, w); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

func (codec *attributeModifiersCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	modifiers := make([]AttributeModifier, 0, count)
	for range int(count) {
		mod, err := decodeAttributeModifier(buf)
		if err != nil {
			return err
		}
		modifiers = append(modifiers, mod)
	}
	c.AttributeModifiers = modifiers
	return nil
}

func (codec *attributeModifiersCodec) Clear(c *Components) {
	c.AttributeModifiers = nil
}

func (codec *attributeModifiersCodec) Differs(c, defaults *Components) (bool, bool) {
	cHas := len(c.AttributeModifiers) > 0
	dHas := len(defaults.AttributeModifiers) > 0
	if cHas != dHas {
		return true, cHas
	}
	if cHas && dHas {
		return !slices.Equal(c.AttributeModifiers, defaults.AttributeModifiers), true
	}
	return false, false
}

func (codec *attributeModifiersCodec) Encode(c *Components) ([]byte, error) {
	w := ns.NewWriter()
	w.WriteVarInt(ns.VarInt(len(c.AttributeModifiers)))
	for _, mod := range c.AttributeModifiers {
		if err := encodeAttributeModifier(w, mod); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// rarityCodec handles Rarity (VarInt enum mapped to string).
type rarityCodec struct{}

var rarityNames = []string{"common", "uncommon", "rare", "epic"}
var rarityIDs = map[string]int32{"common": 0, "uncommon": 1, "rare": 2, "epic": 3}

func (codec *rarityCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	v, err := buf.ReadVarInt()
	if err != nil {
		return nil, err
	}
	w.WriteVarInt(v)
	return w.Bytes(), nil
}

func (codec *rarityCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	v, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	if int(v) < len(rarityNames) {
		c.Rarity = rarityNames[v]
	}
	return nil
}

func (codec *rarityCodec) Clear(c *Components) {
	c.Rarity = ""
}

func (codec *rarityCodec) Differs(c, defaults *Components) (bool, bool) {
	return c.Rarity != defaults.Rarity, c.Rarity != ""
}

func (codec *rarityCodec) Encode(c *Components) ([]byte, error) {
	w := ns.NewWriter()
	w.WriteVarInt(ns.VarInt(rarityIDs[c.Rarity]))
	return w.Bytes(), nil
}

// ============================================================================
// Lore codec
// ============================================================================

// loreCodec handles Lore component (list of NBT text components).
type loreCodec struct{}

func (codec *loreCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	count, err := buf.ReadVarInt()
	if err != nil {
		return nil, err
	}
	w.WriteVarInt(count)
	for range int(count) {
		if err := copyNBT(buf, w); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

func (codec *loreCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	lore := make([]string, 0, count)
	for range int(count) {
		name, err := decodeItemName(buf)
		if err != nil {
			return err
		}
		if name.Translate != "" {
			lore = append(lore, name.Translate)
		} else {
			lore = append(lore, name.Text)
		}
	}
	c.Lore = lore
	return nil
}

func (codec *loreCodec) Clear(c *Components) {
	c.Lore = nil
}

func (codec *loreCodec) Differs(c, defaults *Components) (bool, bool) {
	cHas := len(c.Lore) > 0
	dHas := len(defaults.Lore) > 0
	if cHas != dHas {
		return true, cHas
	}
	if cHas && dHas {
		return !slices.Equal(c.Lore, defaults.Lore), true
	}
	return false, false
}

func (codec *loreCodec) Encode(c *Components) ([]byte, error) {
	w := ns.NewWriter()
	w.WriteVarInt(ns.VarInt(len(c.Lore)))
	for _, line := range c.Lore {
		if err := encodeItemName(w, &ItemNameComponent{Text: line}); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// ============================================================================
// Enchantments codec
// ============================================================================

// enchantmentsCodec handles Enchantments and StoredEnchantments components.
type enchantmentsCodec struct {
	get func(c *Components) map[string]int32
	set func(c *Components, v map[string]int32)
}

func (codec *enchantmentsCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	count, err := buf.ReadVarInt()
	if err != nil {
		return nil, err
	}
	w.WriteVarInt(count)
	for range int(count) {
		// enchantment ID
		if err := w.CopyVarInt(buf); err != nil {
			return nil, err
		}
		// level
		if err := w.CopyVarInt(buf); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

func (codec *enchantmentsCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	enchants := make(map[string]int32, count)
	for range int(count) {
		enchID, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		level, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		// store as "id:<num>" for now since enchantments are data-driven
		enchants[fmt.Sprintf("id:%d", enchID)] = int32(level)
	}
	codec.set(c, enchants)
	return nil
}

func (codec *enchantmentsCodec) Clear(c *Components) {
	codec.set(c, nil)
}

func (codec *enchantmentsCodec) Differs(c, defaults *Components) (bool, bool) {
	cv := codec.get(c)
	dv := codec.get(defaults)
	cHas := len(cv) > 0
	dHas := len(dv) > 0
	if cHas != dHas {
		return true, cHas
	}
	if cHas && dHas {
		// compare maps
		if len(cv) != len(dv) {
			return true, true
		}
		for k, v := range cv {
			if dv[k] != v {
				return true, true
			}
		}
		return false, true
	}
	return false, false
}

func (codec *enchantmentsCodec) Encode(c *Components) ([]byte, error) {
	m := codec.get(c)
	w := ns.NewWriter()
	w.WriteVarInt(ns.VarInt(len(m)))
	for name, level := range m {
		// parse "id:<num>" format or treat as raw ID
		var enchID int32
		if _, err := fmt.Sscanf(name, "id:%d", &enchID); err != nil {
			return nil, fmt.Errorf("invalid enchantment format: %s (expected id:<num>)", name)
		}
		w.WriteVarInt(ns.VarInt(enchID))
		w.WriteVarInt(ns.VarInt(level))
	}
	return w.Bytes(), nil
}

// ============================================================================
// Tool codec
// ============================================================================

// toolCodec handles Tool component.
type toolCodec struct{}

func (codec *toolCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	// rules
	count, err := buf.ReadVarInt()
	if err != nil {
		return nil, err
	}
	w.WriteVarInt(count)
	for range int(count) {
		if err := copyToolRule(buf, w); err != nil {
			return nil, err
		}
	}
	// default mining speed
	if err := w.CopyFloat32(buf); err != nil {
		return nil, err
	}
	// damage per block
	if err := w.CopyVarInt(buf); err != nil {
		return nil, err
	}
	// can destroy blocks in creative
	if err := w.CopyBool(buf); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (codec *toolCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	tool := &Tool{
		Rules: make([]ToolRule, 0, count),
	}
	for range int(count) {
		rule, err := decodeToolRule(buf)
		if err != nil {
			return err
		}
		tool.Rules = append(tool.Rules, rule)
	}
	defaultSpeed, err := buf.ReadFloat32()
	if err != nil {
		return err
	}
	_ = defaultSpeed // stored in rules, not in struct
	damagePerBlock, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	tool.DamagePerBlock = int32(damagePerBlock)
	canDestroy, err := buf.ReadBool()
	if err != nil {
		return err
	}
	tool.CanDestroyBlocksInCreative = bool(canDestroy)
	c.Tool = tool
	return nil
}

func (codec *toolCodec) Clear(c *Components) {
	c.Tool = nil
}

func (codec *toolCodec) Differs(c, defaults *Components) (bool, bool) {
	cHas := c.Tool != nil
	dHas := defaults.Tool != nil
	if cHas != dHas {
		return true, cHas
	}
	if cHas && dHas {
		// simple comparison - check if rules differ
		if len(c.Tool.Rules) != len(defaults.Tool.Rules) {
			return true, true
		}
		return c.Tool.DamagePerBlock != defaults.Tool.DamagePerBlock ||
			c.Tool.CanDestroyBlocksInCreative != defaults.Tool.CanDestroyBlocksInCreative, true
	}
	return false, false
}

func (codec *toolCodec) Encode(c *Components) ([]byte, error) {
	if c.Tool == nil {
		return nil, nil
	}
	w := ns.NewWriter()
	w.WriteVarInt(ns.VarInt(len(c.Tool.Rules)))
	for _, rule := range c.Tool.Rules {
		if err := encodeToolRule(w, rule); err != nil {
			return nil, err
		}
	}
	w.WriteFloat32(1.0) // default mining speed
	w.WriteVarInt(ns.VarInt(c.Tool.DamagePerBlock))
	w.WriteBool(ns.Boolean(c.Tool.CanDestroyBlocksInCreative))
	return w.Bytes(), nil
}

// decodeToolRule reads a tool rule from the buffer.
func decodeToolRule(buf *ns.PacketBuffer) (ToolRule, error) {
	var rule ToolRule
	// blocks (holder set)
	typeID, err := buf.ReadVarInt()
	if err != nil {
		return rule, err
	}
	if typeID == 0 {
		// tag reference
		tag, err := buf.ReadString(maxStringLen)
		if err != nil {
			return rule, err
		}
		rule.Blocks = string(tag)
	} else {
		// list of block IDs - skip for now
		for range int(typeID) - 1 {
			if _, err := buf.ReadVarInt(); err != nil {
				return rule, err
			}
		}
	}
	// optional speed
	hasSpeed, err := buf.ReadBool()
	if err != nil {
		return rule, err
	}
	if hasSpeed {
		speed, err := buf.ReadFloat32()
		if err != nil {
			return rule, err
		}
		rule.Speed = float64(speed)
	}
	// optional correct for drops
	hasCorrect, err := buf.ReadBool()
	if err != nil {
		return rule, err
	}
	if hasCorrect {
		correct, err := buf.ReadBool()
		if err != nil {
			return rule, err
		}
		rule.CorrectForDrops = bool(correct)
	}
	return rule, nil
}

// encodeToolRule writes a tool rule to the buffer.
func encodeToolRule(w *ns.PacketBuffer, rule ToolRule) error {
	// blocks as tag reference
	w.WriteVarInt(0)
	w.WriteString(ns.String(rule.Blocks))
	// speed
	if rule.Speed > 0 {
		w.WriteBool(true)
		w.WriteFloat32(ns.Float32(rule.Speed))
	} else {
		w.WriteBool(false)
	}
	// correct for drops
	w.WriteBool(true)
	w.WriteBool(ns.Boolean(rule.CorrectForDrops))
	return nil
}

// ============================================================================
// Passthrough codec for components we don't fully decode yet
// ============================================================================

// passthroughCodec stores raw bytes without interpreting them.
// Used for components that aren't fully implemented yet.
type passthroughCodec struct {
	decode func(buf *ns.PacketBuffer, w *ns.PacketBuffer) error
}

func (codec *passthroughCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	if err := codec.decode(buf, w); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (codec *passthroughCodec) Apply(c *Components, data []byte) error {
	return nil // passthrough - don't apply to struct
}

func (codec *passthroughCodec) Clear(c *Components) {
	// passthrough - nothing to clear
}

func (codec *passthroughCodec) Differs(c, defaults *Components) (bool, bool) {
	return false, false // passthrough - never differs
}

func (codec *passthroughCodec) Encode(c *Components) ([]byte, error) {
	return nil, fmt.Errorf("passthrough codec cannot encode")
}

// ============================================================================
// Helper functions used by codecs
// ============================================================================

// copyNBT copies an NBT tag from reader to writer.
func copyNBT(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	reader := nbt.NewReaderFrom(buf.Reader())
	tag, _, err := reader.ReadTag(true) // network format
	if err != nil {
		return err
	}

	writer := nbt.NewWriterTo(w.Writer())
	return writer.WriteTag(tag, "", true) // network format
}

// decodeItemName reads an NBT text component and returns an ItemNameComponent.
func decodeItemName(buf *ns.PacketBuffer) (*ItemNameComponent, error) {
	reader := nbt.NewReaderFrom(buf.Reader())
	tag, _, err := reader.ReadTag(true) // network format
	if err != nil {
		return nil, err
	}

	// text component can be a string (literal) or compound (with translate/etc)
	switch v := tag.(type) {
	case nbt.String:
		return &ItemNameComponent{Text: string(v)}, nil
	case nbt.Compound:
		name := &ItemNameComponent{}
		if t, ok := v["text"].(nbt.String); ok {
			name.Text = string(t)
		}
		if t, ok := v["translate"].(nbt.String); ok {
			name.Translate = string(t)
		}
		return name, nil
	default:
		return &ItemNameComponent{}, nil
	}
}

// encodeItemName writes an ItemNameComponent as NBT text component.
func encodeItemName(w *ns.PacketBuffer, name *ItemNameComponent) error {
	writer := nbt.NewWriterTo(w.Writer())
	var tag nbt.Tag
	if name.Translate != "" {
		tag = nbt.Compound{"translate": nbt.String(name.Translate)}
	} else {
		tag = nbt.String(name.Text)
	}
	return writer.WriteTag(tag, "", true) // network format
}

// decodeAttributeModifier reads an attribute modifier entry.
func decodeAttributeModifier(buf *ns.PacketBuffer) (AttributeModifier, error) {
	var mod AttributeModifier

	// attribute ID (registry reference)
	attrID, err := buf.ReadVarInt()
	if err != nil {
		return mod, err
	}
	mod.Type = registries.Attribute.ByID(int32(attrID))

	// modifier ID (Identifier string)
	modID, err := buf.ReadString(maxStringLen)
	if err != nil {
		return mod, err
	}
	mod.ID = string(modID)

	// amount (Double)
	amount, err := buf.ReadFloat64()
	if err != nil {
		return mod, err
	}
	mod.Amount = float64(amount)

	// operation (VarInt)
	operation, err := buf.ReadVarInt()
	if err != nil {
		return mod, err
	}
	operations := []string{"add_value", "add_multiplied_base", "add_multiplied_total"}
	if int(operation) < len(operations) {
		mod.Operation = operations[operation]
	}

	// slot (VarInt - equipment slot group)
	slot, err := buf.ReadVarInt()
	if err != nil {
		return mod, err
	}
	slots := []string{"any", "hand", "mainhand", "offhand", "armor", "feet", "legs", "chest", "head", "body"}
	if int(slot) < len(slots) {
		mod.Slot = slots[slot]
	}

	// display type (VarInt)
	displayType, err := buf.ReadVarInt()
	if err != nil {
		return mod, err
	}
	if displayType == 2 {
		// OVERRIDE includes a Component (NBT) - skip for now
		reader := nbt.NewReaderFrom(buf.Reader())
		_, _, _ = reader.ReadTag(true)
	}

	return mod, nil
}

// encodeAttributeModifier writes an attribute modifier entry.
func encodeAttributeModifier(w *ns.PacketBuffer, mod AttributeModifier) error {
	// attribute type (registry ID)
	attrID := registries.Attribute.Get(mod.Type)
	if attrID < 0 {
		return fmt.Errorf("unknown attribute type: %s", mod.Type)
	}
	w.WriteVarInt(ns.VarInt(attrID))

	// modifier ID (Identifier string)
	w.WriteString(ns.String(mod.ID))

	// amount (Double)
	w.WriteFloat64(ns.Float64(mod.Amount))

	// operation (VarInt)
	operations := map[string]int32{"add_value": 0, "add_multiplied_base": 1, "add_multiplied_total": 2}
	w.WriteVarInt(ns.VarInt(operations[mod.Operation]))

	// slot (VarInt - equipment slot group)
	slots := map[string]int32{"any": 0, "hand": 1, "mainhand": 2, "offhand": 3, "armor": 4, "feet": 5, "legs": 6, "chest": 7, "head": 8, "body": 9}
	w.WriteVarInt(ns.VarInt(slots[mod.Slot]))

	// display type (VarInt) - 0=DEFAULT, 1:WHEN_NOT_DEFAULT, 2=OVERRIDE
	w.WriteVarInt(0)

	return nil
}

// copyAttributeModifier copies an attribute modifier from reader to writer.
func copyAttributeModifier(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	// attribute type
	attrType, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(attrType)

	// modifier ID
	modID, err := buf.ReadString(maxStringLen)
	if err != nil {
		return err
	}
	w.WriteString(modID)

	// amount
	amount, err := buf.ReadFloat64()
	if err != nil {
		return err
	}
	w.WriteFloat64(amount)

	// operation
	operation, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(operation)

	// slot
	slot, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(slot)

	// display type
	displayType, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(displayType)
	if displayType == 2 {
		// OVERRIDE includes NBT
		if err := copyNBT(buf, w); err != nil {
			return err
		}
	}

	return nil
}

// copyFireworkExplosion copies a firework explosion from reader to writer.
func copyFireworkExplosion(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	// shape
	shape, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(shape)

	// colors
	colorCount, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(colorCount)
	for range int(colorCount) {
		color, err := buf.ReadInt32()
		if err != nil {
			return err
		}
		w.WriteInt32(color)
	}

	// fade colors
	fadeCount, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(fadeCount)
	for range int(fadeCount) {
		color, err := buf.ReadInt32()
		if err != nil {
			return err
		}
		w.WriteInt32(color)
	}

	// trail
	trail, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(trail)

	// twinkle
	twinkle, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(twinkle)

	return nil
}
