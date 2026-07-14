package items

// Component codec registration - registers complex component codecs at init time.
// Simple codecs (varint, float32, string, empty) are auto-generated in item_components_codec_gen.go.

import (
	"fmt"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

func init() {
	// struct components with dedicated codecs
	RegisterCodec(ComponentCustomName, &customNameCodec{})
	RegisterCodec(ComponentItemName, &itemNameCodec{})
	RegisterCodec(ComponentAttributeModifiers, &attributeModifiersCodec{})
	RegisterCodec(ComponentRarity, &rarityCodec{})

	// lore, enchantments, and tool codecs (full implementations)
	RegisterCodec(ComponentLore, &loreCodec{})
	RegisterCodec(ComponentEnchantments, &enchantmentsCodec{
		get: func(c *Components) map[string]int32 { return c.Enchantments },
		set: func(c *Components, v map[string]int32) { c.Enchantments = v },
	})
	RegisterCodec(ComponentStoredEnchantments, &enchantmentsCodec{
		get: func(c *Components) map[string]int32 { return c.StoredEnchantments },
		set: func(c *Components, v map[string]int32) { c.StoredEnchantments = v },
	})
	RegisterCodec(ComponentTool, &toolCodec{})

	// complex passthrough codecs - these have custom decoders
	// simple passthroughs (varint, bool, string, empty, int32, nbt, holderSet, slot, slotList)
	// are registered in item_components_codec_gen.go
	RegisterCodec(ComponentCanBreak, &passthroughCodec{decode: decodeBlockPredicatesWire})
	RegisterCodec(ComponentCanPlaceOn, &passthroughCodec{decode: decodeBlockPredicatesWire})
	RegisterCodec(ComponentCustomModelData, &passthroughCodec{decode: decodeCustomModelDataWire})
	RegisterCodec(ComponentConsumable, &passthroughCodec{decode: decodeConsumableWire})
	RegisterCodec(ComponentUseEffects, &passthroughCodec{decode: decodeUseEffectsWire})
	RegisterCodec(ComponentAttackRange, &passthroughCodec{decode: decodeAttackRangeWire})
	RegisterCodec(ComponentEquippable, &passthroughCodec{decode: decodeEquippableWire})
	RegisterCodec(ComponentDeathProtection, &passthroughCodec{decode: decodeDeathProtectionWire})
	RegisterCodec(ComponentBlocksAttacks, &passthroughCodec{decode: decodeBlocksAttacksWire})
	RegisterCodec(ComponentKineticWeapon, &passthroughCodec{decode: decodeKineticWeaponWire})
	RegisterCodec(ComponentPiercingWeapon, &passthroughCodec{decode: decodePiercingWeaponWire})
	RegisterCodec(ComponentPotionContents, &passthroughCodec{decode: decodePotionContentsWire})
	RegisterCodec(ComponentSuspiciousStewEffects, &passthroughCodec{decode: decodeSuspiciousStewWire})
	RegisterCodec(ComponentWritableBookContent, &passthroughCodec{decode: decodeWritableBookWire})
	RegisterCodec(ComponentWrittenBookContent, &passthroughCodec{decode: decodeWrittenBookWire})
	RegisterCodec(ComponentTrim, &passthroughCodec{decode: decodeTrimWire})
	RegisterCodec(ComponentRecipes, &passthroughCodec{decode: decodeRecipesWire})
	RegisterCodec(ComponentLodestoneTracker, &passthroughCodec{decode: decodeLodestoneWire})
	RegisterCodec(ComponentFireworkExplosion, &passthroughCodec{decode: decodeFireworkExplosionWire})
	RegisterCodec(ComponentProfile, &passthroughCodec{decode: decodeProfileWire})
	RegisterCodec(ComponentBannerPatterns, &passthroughCodec{decode: decodeBannerPatternsWire})
	RegisterCodec(ComponentPotDecorations, &passthroughCodec{decode: decodePotDecorationsWire})
	RegisterCodec(ComponentBlockState, &passthroughCodec{decode: decodeBlockStateWire})
	RegisterCodec(ComponentBees, &passthroughCodec{decode: decodeBeesWire})
}

// Helper functions for registering simple passthrough codecs

func registerVarIntPassthrough(id int32) {
	RegisterCodec(id, &passthroughCodec{decode: decodeVarIntWire})
}

func registerBoolPassthrough(id int32) {
	RegisterCodec(id, &passthroughCodec{decode: decodeBoolWire})
}

func registerStringPassthrough(id int32) {
	RegisterCodec(id, &passthroughCodec{decode: decodeStringWire})
}

func registerEmptyPassthrough(id int32) {
	RegisterCodec(id, &passthroughCodec{decode: decodeEmptyWire})
}

func registerInt32Passthrough(id int32) {
	RegisterCodec(id, &passthroughCodec{decode: decodeInt32Wire})
}

func registerNBTPassthrough(id int32) {
	RegisterCodec(id, &passthroughCodec{decode: decodeNBTWire})
}

func registerHolderSetPassthrough(id int32) {
	RegisterCodec(id, &passthroughCodec{decode: decodeHolderSetWire})
}

func registerSlotListPassthrough(id int32) {
	RegisterCodec(id, &passthroughCodec{decode: decodeSlotListWire})
}

func registerSlotPassthrough(id int32) {
	RegisterCodec(id, &passthroughCodec{decode: decodeSlotWire})
}

// Wire format decode functions for passthrough codecs

func decodeVarIntWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return w.CopyVarInt(buf)
}

func decodeInt32Wire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return w.CopyInt32(buf)
}

func decodeBoolWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return w.CopyBool(buf)
}

func decodeStringWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return w.CopyString(buf, maxStringLen)
}

func decodeEmptyWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return nil
}

func decodeNBTWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return copyNBT(buf, w)
}

func decodeBlockPredicatesWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := copyVarIntPrefixedList(buf, w, copyBlockPredicate); err != nil {
		return err
	}
	return w.CopyBool(buf)
}

func decodeCustomModelDataWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	copyFloat32Fn := func(buf, w *ns.PacketBuffer) error { return w.CopyFloat32(buf) }
	copyBoolFn := func(buf, w *ns.PacketBuffer) error { return w.CopyBool(buf) }
	copyStringFn := func(buf, w *ns.PacketBuffer) error { return w.CopyString(buf, maxStringLen) }
	copyInt32Fn := func(buf, w *ns.PacketBuffer) error { return w.CopyInt32(buf) }

	if err := copyVarIntPrefixedList(buf, w, copyFloat32Fn); err != nil {
		return err
	}
	if err := copyVarIntPrefixedList(buf, w, copyBoolFn); err != nil {
		return err
	}
	if err := copyVarIntPrefixedList(buf, w, copyStringFn); err != nil {
		return err
	}
	return copyVarIntPrefixedList(buf, w, copyInt32Fn)
}

func decodeConsumableWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := w.CopyFloat32(buf); err != nil { // consume seconds
		return err
	}
	if err := w.CopyVarInt(buf); err != nil { // animation
		return err
	}
	if err := copySoundEvent(buf, w); err != nil {
		return err
	}
	if err := w.CopyBool(buf); err != nil { // has particles
		return err
	}
	return copyVarIntPrefixedList(buf, w, copyConsumeEffect)
}

func decodeSlotWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return copySlot(buf, w)
}

func decodeSlotListWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return copyVarIntPrefixedList(buf, w, copySlot)
}

func decodeUseEffectsWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := w.CopyBool(buf); err != nil { // can sprint
		return err
	}
	if err := w.CopyBool(buf); err != nil { // interact vibrations
		return err
	}
	return w.CopyFloat32(buf) // speed multiplier
}

func decodeHolderSetWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return copyHolderSet(buf, w)
}

func decodeAttackRangeWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	for range 6 {
		if err := w.CopyFloat32(buf); err != nil {
			return err
		}
	}
	return nil
}

func decodeEquippableWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := w.CopyVarInt(buf); err != nil { // slot
		return err
	}
	if err := copySoundEvent(buf, w); err != nil {
		return err
	}
	if err := copyOptionalString(buf, w); err != nil { // optional asset
		return err
	}
	if err := copyOptionalIdentifier(buf, w); err != nil { // optional camera overlay
		return err
	}
	if err := copyOptionalHolderSet(buf, w); err != nil {
		return err
	}
	// dispensable, swappable, damages on hurt
	for range 3 {
		if err := w.CopyBool(buf); err != nil {
			return err
		}
	}
	// can be sheared (conditional)
	canBeSheared, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(canBeSheared)
	if canBeSheared {
		return copySoundEvent(buf, w)
	}
	return nil
}

func decodeDeathProtectionWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return copyVarIntPrefixedList(buf, w, copyConsumeEffect)
}

func decodeBlocksAttacksWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := w.CopyFloat32(buf); err != nil { // block delay seconds
		return err
	}
	if err := w.CopyFloat32(buf); err != nil { // disable cooldown scale
		return err
	}
	if err := copyVarIntPrefixedList(buf, w, copyDamageReduction); err != nil {
		return err
	}
	if err := copyItemDamageFunction(buf, w); err != nil {
		return err
	}
	if err := copyOptionalHolderSet(buf, w); err != nil { // bypassed by
		return err
	}
	if err := copyOptionalSoundEvent(buf, w); err != nil { // block sound
		return err
	}
	return copyOptionalSoundEvent(buf, w) // disable sound
}

func decodeKineticWeaponWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := w.CopyFloat32(buf); err != nil { // damage multiplier
		return err
	}
	// damage, dismount, knockback conditions
	for range 3 {
		if err := copyOptionalKineticConditions(buf, w); err != nil {
			return err
		}
	}
	if err := w.CopyFloat32(buf); err != nil { // forward movement
		return err
	}
	if err := w.CopyVarInt(buf); err != nil { // delay ticks
		return err
	}
	if err := copyOptionalSoundEvent(buf, w); err != nil { // sound
		return err
	}
	return copyOptionalSoundEvent(buf, w) // hit sound
}

func decodePiercingWeaponWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	// optional sound
	if err := copyOptionalSoundEvent(buf, w); err != nil {
		return err
	}
	// optional hit sound
	if err := copyOptionalSoundEvent(buf, w); err != nil {
		return err
	}
	return nil
}

func decodePotionContentsWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := copyOptionalVarInt(buf, w); err != nil { // potion
		return err
	}
	if err := copyOptionalInt(buf, w); err != nil { // custom color
		return err
	}
	if err := copyVarIntPrefixedList(buf, w, copyStatusEffect); err != nil { // custom effects
		return err
	}
	return copyOptionalString(buf, w) // custom name
}

func decodeSuspiciousStewWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return copyVarIntPrefixedList(buf, w, func(buf, w *ns.PacketBuffer) error {
		if err := w.CopyVarInt(buf); err != nil { // effect ID
			return err
		}
		return w.CopyVarInt(buf) // duration
	})
}

func decodeWritableBookWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return copyVarIntPrefixedList(buf, w, func(buf, w *ns.PacketBuffer) error {
		if err := w.CopyString(buf, maxStringLen); err != nil { // raw content
			return err
		}
		return copyOptionalString(buf, w) // optional filtered content
	})
}

func decodeWrittenBookWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := w.CopyString(buf, maxStringLen); err != nil { // raw title
		return err
	}
	if err := copyOptionalString(buf, w); err != nil { // filtered title
		return err
	}
	if err := w.CopyString(buf, maxStringLen); err != nil { // author
		return err
	}
	if err := w.CopyVarInt(buf); err != nil { // generation
		return err
	}
	if err := copyVarIntPrefixedList(buf, w, func(buf, w *ns.PacketBuffer) error {
		if err := copyNBT(buf, w); err != nil { // raw NBT
			return err
		}
		return copyOptionalNBT(buf, w) // optional filtered NBT
	}); err != nil {
		return err
	}
	return w.CopyBool(buf) // resolved
}

func decodeTrimWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	// material
	if err := copyTrimMaterial(buf, w); err != nil {
		return err
	}
	// pattern
	if err := copyTrimPattern(buf, w); err != nil {
		return err
	}
	return nil
}

func decodeRecipesWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return copyVarIntPrefixedList(buf, w, func(buf, w *ns.PacketBuffer) error {
		return w.CopyString(buf, maxStringLen)
	})
}

func decodeLodestoneWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := copyOptionalGlobalPos(buf, w); err != nil {
		return err
	}
	return w.CopyBool(buf) // tracked
}

func decodeFireworkExplosionWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return copyFireworkExplosion(buf, w)
}

func decodeProfileWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := copyOptionalString(buf, w); err != nil { // name
		return err
	}
	if err := copyOptionalUUID(buf, w); err != nil {
		return err
	}
	return copyVarIntPrefixedList(buf, w, copyGameProfileProperty)
}

func decodeBannerPatternsWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return copyVarIntPrefixedList(buf, w, copyBannerPattern)
}

func decodePotDecorationsWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	for range 4 {
		// optional item ID
		if err := copyOptionalVarInt(buf, w); err != nil {
			return err
		}
	}
	return nil
}

func decodeBlockStateWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return copyVarIntPrefixedList(buf, w, func(buf, w *ns.PacketBuffer) error {
		if err := w.CopyString(buf, maxStringLen); err != nil { // key
			return err
		}
		return w.CopyString(buf, maxStringLen) // value
	})
}

func decodeBeesWire(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return copyVarIntPrefixedList(buf, w, func(buf, w *ns.PacketBuffer) error {
		if err := copyNBT(buf, w); err != nil { // entity data
			return err
		}
		if err := w.CopyVarInt(buf); err != nil { // ticks in hive
			return err
		}
		return w.CopyVarInt(buf) // min ticks in hive
	})
}

// Copy helper functions

// copyVarIntPrefixedList copies a VarInt count followed by that many elements using the provided copy function.
func copyVarIntPrefixedList(buf *ns.PacketBuffer, w *ns.PacketBuffer, copyFn func(*ns.PacketBuffer, *ns.PacketBuffer) error) error {
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(count)
	for range int(count) {
		if err := copyFn(buf, w); err != nil {
			return err
		}
	}
	return nil
}

func copySlot(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(count)

	if count <= 0 {
		return nil
	}

	itemID, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(itemID)

	addCount, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(addCount)

	removeCount, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(removeCount)

	for range int(addCount) {
		compID, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		w.WriteVarInt(compID)

		codec := componentCodecs[int32(compID)]
		if codec == nil {
			return fmt.Errorf("unknown component %d in slot", compID)
		}
		data, err := codec.DecodeWire(buf)
		if err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
	}

	for range int(removeCount) {
		compID, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		w.WriteVarInt(compID)
	}

	return nil
}

func copyOptionalVarInt(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	present, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(present)
	if present {
		return w.CopyVarInt(buf)
	}
	return nil
}

func copyOptionalInt(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	present, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(present)
	if present {
		return w.CopyInt32(buf)
	}
	return nil
}

func copyOptionalString(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	present, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(present)
	if present {
		return w.CopyString(buf, maxStringLen)
	}
	return nil
}

func copyOptionalIdentifier(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	return copyOptionalString(buf, w)
}

func copyOptionalNBT(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	present, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(present)
	if present {
		return copyNBT(buf, w)
	}
	return nil
}

func copyOptionalUUID(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	present, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(present)
	if present {
		return w.CopyUUID(buf)
	}
	return nil
}

func copyHolderSet(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	typeID, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(typeID)

	if typeID == 0 {
		return w.CopyString(buf, maxStringLen)
	}
	for range int(typeID) - 1 {
		if err := w.CopyVarInt(buf); err != nil {
			return err
		}
	}
	return nil
}

func copyOptionalHolderSet(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	present, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(present)
	if present {
		return copyHolderSet(buf, w)
	}
	return nil
}

func copySoundEvent(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	typeID, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(typeID)

	if typeID == 0 {
		if err := w.CopyString(buf, maxStringLen); err != nil {
			return err
		}
		return copyOptionalVarInt(buf, w)
	}
	return nil
}

func copyOptionalSoundEvent(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	present, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(present)
	if present {
		return copySoundEvent(buf, w)
	}
	return nil
}

func copyConsumeEffect(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	typeID, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(typeID)

	switch typeID {
	case 0: // apply_effects
		if err := copyVarIntPrefixedList(buf, w, copyStatusEffect); err != nil {
			return err
		}
		return w.CopyFloat32(buf) // probability
	case 1: // remove_effects
		return copyHolderSet(buf, w)
	case 2: // clear_all_effects
		return nil
	case 3: // teleport_randomly
		return w.CopyFloat32(buf) // diameter
	case 4: // play_sound
		return copySoundEvent(buf, w)
	}
	return nil
}

func copyStatusEffect(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := w.CopyVarInt(buf); err != nil {
		return err
	}
	return copyStatusEffectDetails(buf, w)
}

func copyStatusEffectDetails(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	// amplifier, duration
	if err := w.CopyVarInt(buf); err != nil {
		return err
	}
	if err := w.CopyVarInt(buf); err != nil {
		return err
	}
	// ambient, show particles, show icon
	for range 3 {
		if err := w.CopyBool(buf); err != nil {
			return err
		}
	}
	// hidden effect (recursive optional)
	hasHidden, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(hasHidden)
	if hasHidden {
		return copyStatusEffectDetails(buf, w)
	}
	return nil
}

func copyToolRule(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := copyHolderSet(buf, w); err != nil {
		return err
	}
	// optional speed
	hasSpeed, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(hasSpeed)
	if hasSpeed {
		if err := w.CopyFloat32(buf); err != nil {
			return err
		}
	}
	// optional correct for drops
	hasCorrect, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(hasCorrect)
	if hasCorrect {
		return w.CopyBool(buf)
	}
	return nil
}

func copyBlockPredicate(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	// optional blocks
	if err := copyOptionalHolderSet(buf, w); err != nil {
		return err
	}
	// optional properties
	propCount, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(propCount)
	for range int(propCount) {
		if err := copyPropertyMatcher(buf, w); err != nil {
			return err
		}
	}
	// optional NBT
	if err := copyOptionalNBT(buf, w); err != nil {
		return err
	}
	return nil
}

func copyPropertyMatcher(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := w.CopyString(buf, maxStringLen); err != nil { // property name
		return err
	}
	isExact, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(isExact)
	if isExact {
		return w.CopyString(buf, maxStringLen) // value
	}
	// range: min, max
	if err := copyOptionalString(buf, w); err != nil {
		return err
	}
	return copyOptionalString(buf, w)
}

func copyDamageReduction(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := w.CopyFloat32(buf); err != nil { // horizontal angle
		return err
	}
	if err := copyHolderSet(buf, w); err != nil {
		return err
	}
	// base, factor, horizontal limit
	for range 3 {
		if err := w.CopyFloat32(buf); err != nil {
			return err
		}
	}
	return nil
}

func copyItemDamageFunction(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	// threshold, base, factor
	for range 3 {
		if err := w.CopyFloat32(buf); err != nil {
			return err
		}
	}
	return nil
}

func copyOptionalKineticConditions(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	present, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(present)
	if present {
		if err := w.CopyVarInt(buf); err != nil { // max duration ticks
			return err
		}
		if err := w.CopyFloat32(buf); err != nil { // min speed
			return err
		}
		return w.CopyFloat32(buf) // min relative speed
	}
	return nil
}

func copyTrimMaterial(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	typeID, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(typeID)

	if typeID == 0 {
		if err := w.CopyString(buf, maxStringLen); err != nil { // asset name
			return err
		}
		if err := w.CopyVarInt(buf); err != nil { // ingredient
			return err
		}
		if err := w.CopyFloat32(buf); err != nil { // item model index
			return err
		}
		if err := copyVarIntPrefixedList(buf, w, func(buf, w *ns.PacketBuffer) error {
			if err := w.CopyVarInt(buf); err != nil { // armor type
				return err
			}
			return w.CopyString(buf, maxStringLen) // asset name
		}); err != nil {
			return err
		}
		return copyNBT(buf, w)
	}
	return nil
}

func copyTrimPattern(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	typeID, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(typeID)

	if typeID == 0 {
		if err := w.CopyString(buf, maxStringLen); err != nil { // asset ID
			return err
		}
		if err := w.CopyVarInt(buf); err != nil { // template item
			return err
		}
		if err := copyNBT(buf, w); err != nil {
			return err
		}
		return w.CopyBool(buf) // decal
	}
	return nil
}

func copyBannerPattern(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	typeID, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	w.WriteVarInt(typeID)

	if typeID == 0 {
		if err := w.CopyString(buf, maxStringLen); err != nil { // asset ID
			return err
		}
		if err := w.CopyString(buf, maxStringLen); err != nil { // translation key
			return err
		}
	}

	return w.CopyVarInt(buf) // color
}

func copyGameProfileProperty(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	if err := w.CopyString(buf, maxStringLen); err != nil { // name
		return err
	}
	if err := w.CopyString(buf, maxStringLen); err != nil { // value
		return err
	}
	return copyOptionalString(buf, w) // signature
}

func copyOptionalGlobalPos(buf *ns.PacketBuffer, w *ns.PacketBuffer) error {
	present, err := buf.ReadBool()
	if err != nil {
		return err
	}
	w.WriteBool(present)
	if present {
		if err := w.CopyString(buf, maxStringLen); err != nil { // dimension
			return err
		}
		return w.CopyPosition(buf) // position
	}
	return nil
}
