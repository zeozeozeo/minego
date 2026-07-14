// Code generated for Minecraft 1.21.11 (Protocol 774); DO NOT EDIT.

package items

import (
	"slices"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// Auto-generated codec registrations.

func init() {
	// Simple VarInt codecs
	RegisterCodec(ComponentDamage, &varIntCodec{
		get: func(c *Components) int32 { return c.Damage },
		set: func(c *Components, v int32) { c.Damage = v },
	})
	RegisterCodec(ComponentMapColor, &varIntCodec{
		get: func(c *Components) int32 { return c.MapColor },
		set: func(c *Components, v int32) { c.MapColor = v },
	})
	RegisterCodec(ComponentMaxDamage, &varIntCodec{
		get: func(c *Components) int32 { return c.MaxDamage },
		set: func(c *Components, v int32) { c.MaxDamage = v },
	})
	RegisterCodec(ComponentMaxStackSize, &varIntCodec{
		get: func(c *Components) int32 { return c.MaxStackSize },
		set: func(c *Components, v int32) { c.MaxStackSize = v },
	})
	RegisterCodec(ComponentOminousBottleAmplifier, &varIntCodec{
		get: func(c *Components) int32 { return c.OminousBottleAmplifier },
		set: func(c *Components, v int32) { c.OminousBottleAmplifier = v },
	})
	RegisterCodec(ComponentRepairCost, &varIntCodec{
		get: func(c *Components) int32 { return c.RepairCost },
		set: func(c *Components, v int32) { c.RepairCost = v },
	})

	// Simple Float32 codecs
	RegisterCodec(ComponentMinimumAttackCharge, &float32Codec{
		get: func(c *Components) float64 { return c.MinimumAttackCharge },
		set: func(c *Components, v float64) { c.MinimumAttackCharge = v },
	})
	RegisterCodec(ComponentPotionDurationScale, &float32Codec{
		get: func(c *Components) float64 { return c.PotionDurationScale },
		set: func(c *Components, v float64) { c.PotionDurationScale = v },
	})

	// Simple String/Identifier codecs
	RegisterCodec(ComponentBreakSound, &stringCodec{
		get: func(c *Components) string { return c.BreakSound },
		set: func(c *Components, v string) { c.BreakSound = v },
	})
	RegisterCodec(ComponentInstrument, &stringCodec{
		get: func(c *Components) string { return c.Instrument },
		set: func(c *Components, v string) { c.Instrument = v },
	})
	RegisterCodec(ComponentItemModel, &stringCodec{
		get: func(c *Components) string { return c.ItemModel },
		set: func(c *Components, v string) { c.ItemModel = v },
	})
	RegisterCodec(ComponentJukeboxPlayable, &stringCodec{
		get: func(c *Components) string { return c.JukeboxPlayable },
		set: func(c *Components, v string) { c.JukeboxPlayable = v },
	})

	// Empty marker codecs (bool flags)
	RegisterCodec(ComponentGlider, &emptyMarkerCodec{
		get: func(c *Components) bool { return c.Glider },
		set: func(c *Components, v bool) { c.Glider = v },
	})
	RegisterCodec(ComponentUnbreakable, &emptyMarkerCodec{
		get: func(c *Components) bool { return c.Unbreakable },
		set: func(c *Components, v bool) { c.Unbreakable = v },
	})

	// Struct codecs
	RegisterCodec(ComponentEnchantable, genEnchantableCodec{})
	RegisterCodec(ComponentFireworks, genFireworksCodec{})
	RegisterCodec(ComponentFood, genFoodCodec{})
	RegisterCodec(ComponentTooltipDisplay, genTooltipDisplayCodec{})
	RegisterCodec(ComponentUseCooldown, genUseCooldownCodec{})
	RegisterCodec(ComponentWeapon, genWeaponCodec{})

	// VarInt passthrough
	for _, id := range []int32{
		ComponentBaseColor,
		ComponentDamageType,
		ComponentMapId,
		ComponentMapPostProcessing,
		ComponentSwingAnimation,
	} {
		registerVarIntPassthrough(id)
	}

	// Bool passthrough
	for _, id := range []int32{
		ComponentEnchantmentGlintOverride,
	} {
		registerBoolPassthrough(id)
	}

	// String passthrough
	for _, id := range []int32{
		ComponentNoteBlockSound,
		ComponentTooltipStyle,
	} {
		registerStringPassthrough(id)
	}

	// Empty passthrough
	for _, id := range []int32{
		ComponentCreativeSlotLock,
		ComponentIntangibleProjectile,
	} {
		registerEmptyPassthrough(id)
	}

	// Int32 passthrough
	for _, id := range []int32{
		ComponentDyedColor,
	} {
		registerInt32Passthrough(id)
	}

	// NBT passthrough
	for _, id := range []int32{
		ComponentBlockEntityData,
		ComponentBucketEntityData,
		ComponentContainerLoot,
		ComponentCustomData,
		ComponentDebugStickState,
		ComponentEntityData,
		ComponentLock,
		ComponentMapDecorations,
	} {
		registerNBTPassthrough(id)
	}

	// HolderSet passthrough
	for _, id := range []int32{
		ComponentDamageResistant,
		ComponentProvidesBannerPatterns,
		ComponentRepairable,
	} {
		registerHolderSetPassthrough(id)
	}

	// SlotList passthrough
	for _, id := range []int32{
		ComponentBundleContents,
		ComponentChargedProjectiles,
		ComponentContainer,
	} {
		registerSlotListPassthrough(id)
	}

	// Slot passthrough
	for _, id := range []int32{
		ComponentUseRemainder,
	} {
		registerSlotPassthrough(id)
	}

	// Entity variant (VarInt) passthrough
	for _, id := range []int32{
		ComponentAxolotlVariant,
		ComponentCatCollar,
		ComponentCatVariant,
		ComponentChickenVariant,
		ComponentCowVariant,
		ComponentFoxVariant,
		ComponentFrogVariant,
		ComponentHorseVariant,
		ComponentLlamaVariant,
		ComponentMooshroomVariant,
		ComponentPaintingVariant,
		ComponentParrotVariant,
		ComponentPigVariant,
		ComponentRabbitVariant,
		ComponentSalmonSize,
		ComponentSheepColor,
		ComponentShulkerColor,
		ComponentTropicalFishBaseColor,
		ComponentTropicalFishPattern,
		ComponentTropicalFishPatternColor,
		ComponentVillagerVariant,
		ComponentWolfCollar,
		ComponentWolfSoundVariant,
		ComponentWolfVariant,
		ComponentZombieNautilusVariant,
	} {
		registerVarIntPassthrough(id)
	}

}

type genEnchantableCodec struct{}

func (genEnchantableCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	if err := w.CopyVarInt(buf); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (genEnchantableCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	s := &Enchantable{}
	{
		v, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		s.Value = int32(v)
	}
	c.Enchantable = s
	return nil
}

func (genEnchantableCodec) Clear(c *Components) { c.Enchantable = nil }

func (genEnchantableCodec) Differs(c, defaults *Components) (bool, bool) {
	cHas := c.Enchantable != nil
	dHas := defaults.Enchantable != nil
	if cHas != dHas {
		return true, cHas
	}
	if cHas && dHas {
		return *c.Enchantable != *defaults.Enchantable, true
	}
	return false, false
}

func (genEnchantableCodec) Encode(c *Components) ([]byte, error) {
	if c.Enchantable == nil {
		return nil, nil
	}
	w := ns.NewWriter()
	w.WriteVarInt(ns.VarInt(c.Enchantable.Value))
	return w.Bytes(), nil
}

type genFireworksCodec struct{}

func (genFireworksCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	if err := w.CopyVarInt(buf); err != nil {
		return nil, err
	}
	{
		count, err := buf.ReadVarInt()
		if err != nil {
			return nil, err
		}
		w.WriteVarInt(count)
		for range int(count) {
			if err := copyFireworkExplosion(buf, w); err != nil {
				return nil, err
			}
		}
	}
	return w.Bytes(), nil
}

func (genFireworksCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	s := &Fireworks{}
	{
		v, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		s.FlightDuration = int32(v)
	}
	c.Fireworks = s
	return nil
}

func (genFireworksCodec) Clear(c *Components) { c.Fireworks = nil }

func (genFireworksCodec) Differs(c, defaults *Components) (bool, bool) {
	cHas := c.Fireworks != nil
	dHas := defaults.Fireworks != nil
	if cHas != dHas {
		return true, cHas
	}
	if cHas && dHas {
		return *c.Fireworks != *defaults.Fireworks, true
	}
	return false, false
}

func (genFireworksCodec) Encode(c *Components) ([]byte, error) {
	if c.Fireworks == nil {
		return nil, nil
	}
	w := ns.NewWriter()
	w.WriteVarInt(ns.VarInt(c.Fireworks.FlightDuration))
	w.WriteVarInt(0)
	return w.Bytes(), nil
}

type genFoodCodec struct{}

func (genFoodCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	if err := w.CopyVarInt(buf); err != nil {
		return nil, err
	}
	if err := w.CopyFloat32(buf); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (genFoodCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	s := &Food{}
	{
		v, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		s.Nutrition = int32(v)
	}
	{
		v, err := buf.ReadFloat32()
		if err != nil {
			return err
		}
		s.Saturation = float64(v)
	}
	c.Food = s
	return nil
}

func (genFoodCodec) Clear(c *Components) { c.Food = nil }

func (genFoodCodec) Differs(c, defaults *Components) (bool, bool) {
	cHas := c.Food != nil
	dHas := defaults.Food != nil
	if cHas != dHas {
		return true, cHas
	}
	if cHas && dHas {
		return *c.Food != *defaults.Food, true
	}
	return false, false
}

func (genFoodCodec) Encode(c *Components) ([]byte, error) {
	if c.Food == nil {
		return nil, nil
	}
	w := ns.NewWriter()
	w.WriteVarInt(ns.VarInt(c.Food.Nutrition))
	w.WriteFloat32(ns.Float32(c.Food.Saturation))
	return w.Bytes(), nil
}

type genTooltipDisplayCodec struct{}

func (genTooltipDisplayCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	if err := w.CopyBool(buf); err != nil {
		return nil, err
	}
	{
		count, err := buf.ReadVarInt()
		if err != nil {
			return nil, err
		}
		w.WriteVarInt(count)
		for range int(count) {
			if err := w.CopyVarInt(buf); err != nil {
				return nil, err
			}
		}
	}
	return w.Bytes(), nil
}

func (genTooltipDisplayCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	s := &TooltipDisplay{}
	{
		v, err := buf.ReadBool()
		if err != nil {
			return err
		}
		s.HideTooltip = bool(v)
	}
	{
		count, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		arr := make([]int32, 0, count)
		for range int(count) {
			v, err := buf.ReadVarInt()
			if err != nil {
				return err
			}
			arr = append(arr, int32(v))
		}
		s.HiddenComponents = arr
	}
	c.TooltipDisplay = s
	return nil
}

func (genTooltipDisplayCodec) Clear(c *Components) { c.TooltipDisplay = nil }

func (genTooltipDisplayCodec) Differs(c, defaults *Components) (bool, bool) {
	cHas := c.TooltipDisplay != nil
	dHas := defaults.TooltipDisplay != nil
	if cHas != dHas {
		return true, cHas
	}
	if cHas && dHas {
		if c.TooltipDisplay.HideTooltip != defaults.TooltipDisplay.HideTooltip {
			return true, true
		}
		if !slices.Equal(c.TooltipDisplay.HiddenComponents, defaults.TooltipDisplay.HiddenComponents) {
			return true, true
		}
		return false, true
	}
	return false, false
}

func (genTooltipDisplayCodec) Encode(c *Components) ([]byte, error) {
	if c.TooltipDisplay == nil {
		return nil, nil
	}
	w := ns.NewWriter()
	w.WriteBool(ns.Boolean(c.TooltipDisplay.HideTooltip))
	w.WriteVarInt(ns.VarInt(len(c.TooltipDisplay.HiddenComponents)))
	for _, v := range c.TooltipDisplay.HiddenComponents {
		w.WriteVarInt(ns.VarInt(v))
	}
	return w.Bytes(), nil
}

type genUseCooldownCodec struct{}

func (genUseCooldownCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	if err := w.CopyFloat32(buf); err != nil {
		return nil, err
	}
	{
		present, err := buf.ReadBool()
		if err != nil {
			return nil, err
		}
		w.WriteBool(present)
		if present {
			if err := w.CopyString(buf, maxStringLen); err != nil {
				return nil, err
			}
		}
	}
	return w.Bytes(), nil
}

func (genUseCooldownCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	s := &UseCooldown{}
	{
		v, err := buf.ReadFloat32()
		if err != nil {
			return err
		}
		s.Seconds = float64(v)
	}
	c.UseCooldown = s
	return nil
}

func (genUseCooldownCodec) Clear(c *Components) { c.UseCooldown = nil }

func (genUseCooldownCodec) Differs(c, defaults *Components) (bool, bool) {
	cHas := c.UseCooldown != nil
	dHas := defaults.UseCooldown != nil
	if cHas != dHas {
		return true, cHas
	}
	if cHas && dHas {
		return *c.UseCooldown != *defaults.UseCooldown, true
	}
	return false, false
}

func (genUseCooldownCodec) Encode(c *Components) ([]byte, error) {
	if c.UseCooldown == nil {
		return nil, nil
	}
	w := ns.NewWriter()
	w.WriteFloat32(ns.Float32(c.UseCooldown.Seconds))
	w.WriteBool(false)
	return w.Bytes(), nil
}

type genWeaponCodec struct{}

func (genWeaponCodec) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {
	w := ns.NewWriter()
	if err := w.CopyVarInt(buf); err != nil {
		return nil, err
	}
	if err := w.CopyFloat32(buf); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (genWeaponCodec) Apply(c *Components, data []byte) error {
	buf := ns.NewReader(data)
	s := &Weapon{}
	{
		v, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		s.ItemDamagePerAttack = int32(v)
	}
	{
		v, err := buf.ReadFloat32()
		if err != nil {
			return err
		}
		s.DisableBlockingForSeconds = float64(v)
	}
	c.Weapon = s
	return nil
}

func (genWeaponCodec) Clear(c *Components) { c.Weapon = nil }

func (genWeaponCodec) Differs(c, defaults *Components) (bool, bool) {
	cHas := c.Weapon != nil
	dHas := defaults.Weapon != nil
	if cHas != dHas {
		return true, cHas
	}
	if cHas && dHas {
		return *c.Weapon != *defaults.Weapon, true
	}
	return false, false
}

func (genWeaponCodec) Encode(c *Components) ([]byte, error) {
	if c.Weapon == nil {
		return nil, nil
	}
	w := ns.NewWriter()
	w.WriteVarInt(ns.VarInt(c.Weapon.ItemDamagePerAttack))
	w.WriteFloat32(ns.Float32(c.Weapon.DisableBlockingForSeconds))
	return w.Bytes(), nil
}
