package items

import (
	"fmt"
	"strings"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"github.com/zeozeozeo/minego/internal/protocol/nbt"
)

// componentNames maps component IDs to their minecraft identifiers.
var componentNames = map[int32]string{
	ComponentCustomData:               "minecraft:custom_data",
	ComponentMaxStackSize:             "minecraft:max_stack_size",
	ComponentMaxDamage:                "minecraft:max_damage",
	ComponentDamage:                   "minecraft:damage",
	ComponentUnbreakable:              "minecraft:unbreakable",
	ComponentUseEffects:               "minecraft:use_effects",
	ComponentCustomName:               "minecraft:custom_name",
	ComponentMinimumAttackCharge:      "minecraft:minimum_attack_charge",
	ComponentDamageType:               "minecraft:damage_type",
	ComponentItemName:                 "minecraft:item_name",
	ComponentItemModel:                "minecraft:item_model",
	ComponentLore:                     "minecraft:lore",
	ComponentRarity:                   "minecraft:rarity",
	ComponentEnchantments:             "minecraft:enchantments",
	ComponentCanPlaceOn:               "minecraft:can_place_on",
	ComponentCanBreak:                 "minecraft:can_break",
	ComponentAttributeModifiers:       "minecraft:attribute_modifiers",
	ComponentCustomModelData:          "minecraft:custom_model_data",
	ComponentTooltipDisplay:           "minecraft:tooltip_display",
	ComponentRepairCost:               "minecraft:repair_cost",
	ComponentCreativeSlotLock:         "minecraft:creative_slot_lock",
	ComponentEnchantmentGlintOverride: "minecraft:enchantment_glint_override",
	ComponentIntangibleProjectile:     "minecraft:intangible_projectile",
	ComponentFood:                     "minecraft:food",
	ComponentConsumable:               "minecraft:consumable",
	ComponentUseRemainder:             "minecraft:use_remainder",
	ComponentUseCooldown:              "minecraft:use_cooldown",
	ComponentDamageResistant:          "minecraft:damage_resistant",
	ComponentTool:                     "minecraft:tool",
	ComponentWeapon:                   "minecraft:weapon",
	ComponentAttackRange:              "minecraft:attack_range",
	ComponentEnchantable:              "minecraft:enchantable",
	ComponentEquippable:               "minecraft:equippable",
	ComponentRepairable:               "minecraft:repairable",
	ComponentGlider:                   "minecraft:glider",
	ComponentTooltipStyle:             "minecraft:tooltip_style",
	ComponentDeathProtection:          "minecraft:death_protection",
	ComponentBlocksAttacks:            "minecraft:blocks_attacks",
	ComponentPiercingWeapon:           "minecraft:piercing_weapon",
	ComponentKineticWeapon:            "minecraft:kinetic_weapon",
	ComponentSwingAnimation:           "minecraft:swing_animation",
	ComponentStoredEnchantments:       "minecraft:stored_enchantments",
	ComponentDyedColor:                "minecraft:dyed_color",
	ComponentMapColor:                 "minecraft:map_color",
	ComponentMapId:                    "minecraft:map_id",
	ComponentMapDecorations:           "minecraft:map_decorations",
	ComponentMapPostProcessing:        "minecraft:map_post_processing",
	ComponentChargedProjectiles:       "minecraft:charged_projectiles",
	ComponentBundleContents:           "minecraft:bundle_contents",
	ComponentPotionContents:           "minecraft:potion_contents",
	ComponentPotionDurationScale:      "minecraft:potion_duration_scale",
	ComponentSuspiciousStewEffects:    "minecraft:suspicious_stew_effects",
	ComponentWritableBookContent:      "minecraft:writable_book_content",
	ComponentWrittenBookContent:       "minecraft:written_book_content",
	ComponentTrim:                     "minecraft:trim",
	ComponentDebugStickState:          "minecraft:debug_stick_state",
	ComponentEntityData:               "minecraft:entity_data",
	ComponentBucketEntityData:         "minecraft:bucket_entity_data",
	ComponentBlockEntityData:          "minecraft:block_entity_data",
	ComponentInstrument:               "minecraft:instrument",
	ComponentProvidesTrimMaterial:     "minecraft:provides_trim_material",
	ComponentOminousBottleAmplifier:   "minecraft:ominous_bottle_amplifier",
	ComponentJukeboxPlayable:          "minecraft:jukebox_playable",
	ComponentProvidesBannerPatterns:   "minecraft:provides_banner_patterns",
	ComponentRecipes:                  "minecraft:recipes",
	ComponentLodestoneTracker:         "minecraft:lodestone_tracker",
	ComponentFireworkExplosion:        "minecraft:firework_explosion",
	ComponentFireworks:                "minecraft:fireworks",
	ComponentProfile:                  "minecraft:profile",
	ComponentNoteBlockSound:           "minecraft:note_block_sound",
	ComponentBannerPatterns:           "minecraft:banner_patterns",
	ComponentBaseColor:                "minecraft:base_color",
	ComponentPotDecorations:           "minecraft:pot_decorations",
	ComponentContainer:                "minecraft:container",
	ComponentBlockState:               "minecraft:block_state",
	ComponentBees:                     "minecraft:bees",
	ComponentLock:                     "minecraft:lock",
	ComponentContainerLoot:            "minecraft:container_loot",
	ComponentBreakSound:               "minecraft:break_sound",
	ComponentAdditionalTradeCost:      "minecraft:additional_trade_cost",
	ComponentDye:                      "minecraft:dye",
	ComponentVillagerVariant:          "minecraft:villager/variant",
	ComponentWolfVariant:              "minecraft:wolf/variant",
	ComponentWolfSoundVariant:         "minecraft:wolf/sound_variant",
	ComponentWolfCollar:               "minecraft:wolf/collar",
	ComponentFoxVariant:               "minecraft:fox/variant",
	ComponentSalmonSize:               "minecraft:salmon/size",
	ComponentParrotVariant:            "minecraft:parrot/variant",
	ComponentTropicalFishPattern:      "minecraft:tropical_fish/pattern",
	ComponentTropicalFishBaseColor:    "minecraft:tropical_fish/base_color",
	ComponentTropicalFishPatternColor: "minecraft:tropical_fish/pattern_color",
	ComponentMooshroomVariant:         "minecraft:mooshroom/variant",
	ComponentRabbitVariant:            "minecraft:rabbit/variant",
	ComponentPigVariant:               "minecraft:pig/variant",
	ComponentPigSoundVariant:          "minecraft:pig/sound_variant",
	ComponentCowVariant:               "minecraft:cow/variant",
	ComponentCowSoundVariant:          "minecraft:cow/sound_variant",
	ComponentChickenVariant:           "minecraft:chicken/variant",
	ComponentChickenSoundVariant:      "minecraft:chicken/sound_variant",
	ComponentZombieNautilusVariant:    "minecraft:zombie_nautilus/variant",
	ComponentFrogVariant:              "minecraft:frog/variant",
	ComponentHorseVariant:             "minecraft:horse/variant",
	ComponentPaintingVariant:          "minecraft:painting/variant",
	ComponentLlamaVariant:             "minecraft:llama/variant",
	ComponentAxolotlVariant:           "minecraft:axolotl/variant",
	ComponentCatVariant:               "minecraft:cat/variant",
	ComponentCatSoundVariant:          "minecraft:cat/sound_variant",
	ComponentCatCollar:                "minecraft:cat/collar",
	ComponentSheepColor:               "minecraft:sheep/color",
	ComponentShulkerColor:             "minecraft:shulker/color",
}

// ComponentName returns the minecraft identifier for a component ID, or empty string if unknown.
func ComponentName(id int32) string {
	return componentNames[id]
}

// FormatSlotForDisplay formats a raw Slot for human-readable display.
// It shows only the components that are actually sent over the wire,
// without merging with item defaults.
func FormatSlotForDisplay(slot ns.Slot, indent string) string {
	if slot.IsEmpty() {
		return "Empty"
	}

	var sb strings.Builder
	sb.WriteString("ItemStack {\n")
	sb.WriteString(indent)
	sb.WriteString(fmt.Sprintf("  Item: %s (ID: %d)\n", ItemName(int32(slot.ItemID)), slot.ItemID))
	sb.WriteString(indent)
	sb.WriteString(fmt.Sprintf("  Count: %d\n", slot.Count))

	// Show added components (wire data only)
	if len(slot.Components.Add) > 0 {
		sb.WriteString(indent)
		sb.WriteString("  Components: {\n")
		for _, comp := range slot.Components.Add {
			sb.WriteString(indent)
			sb.WriteString("    ")
			name := ComponentName(int32(comp.ID))
			if name == "" {
				name = fmt.Sprintf("unknown(%d)", comp.ID)
			}
			sb.WriteString(name)
			sb.WriteString(": ")
			sb.WriteString(formatComponentValue(int32(comp.ID), comp.Data, indent+"    "))
			sb.WriteString("\n")
		}
		sb.WriteString(indent)
		sb.WriteString("  }\n")
	}

	// Show removed components
	if len(slot.Components.Remove) > 0 {
		sb.WriteString(indent)
		sb.WriteString("  Removed: [")
		for i, id := range slot.Components.Remove {
			if i > 0 {
				sb.WriteString(", ")
			}
			name := ComponentName(int32(id))
			if name == "" {
				name = fmt.Sprintf("unknown(%d)", id)
			}
			sb.WriteString(name)
		}
		sb.WriteString("]\n")
	}

	sb.WriteString(indent)
	sb.WriteString("}")
	return sb.String()
}

// formatComponentValue formats a component's raw data for display.
func formatComponentValue(id int32, data []byte, indent string) string {
	if len(data) == 0 {
		return "true" // empty marker component (like unbreakable)
	}

	buf := ns.NewReader(data)

	switch id {
	case ComponentMaxStackSize, ComponentMaxDamage, ComponentDamage,
		ComponentRepairCost, ComponentOminousBottleAmplifier, ComponentAttackRange:
		// VarInt
		if v, err := buf.ReadVarInt(); err == nil {
			return fmt.Sprintf("%d", v)
		}

	case ComponentMinimumAttackCharge, ComponentPotionDurationScale:
		// Float32
		if v, err := buf.ReadFloat32(); err == nil {
			return fmt.Sprintf("%g", v)
		}

	case ComponentItemModel, ComponentDamageType, ComponentInstrument,
		ComponentProvidesTrimMaterial, ComponentJukeboxPlayable,
		ComponentProvidesBannerPatterns, ComponentBreakSound, ComponentTooltipStyle:
		// String/Identifier
		if v, err := buf.ReadString(32767); err == nil {
			return fmt.Sprintf("%q", v)
		}

	case ComponentUnbreakable, ComponentGlider, ComponentCreativeSlotLock,
		ComponentIntangibleProjectile, ComponentEnchantmentGlintOverride:
		// Empty marker - already handled above
		return "true"

	case ComponentRarity:
		// VarInt enum
		if v, err := buf.ReadVarInt(); err == nil {
			rarities := []string{"common", "uncommon", "rare", "epic"}
			if int(v) < len(rarities) {
				return rarities[v]
			}
			return fmt.Sprintf("%d", v)
		}

	case ComponentFood:
		// nutrition (VarInt) + saturation (Float32)
		nutrition, err1 := buf.ReadVarInt()
		saturation, err2 := buf.ReadFloat32()
		if err1 == nil && err2 == nil {
			return fmt.Sprintf("{nutrition: %d, saturation: %g}", nutrition, saturation)
		}

	case ComponentWeapon:
		// item_damage (VarInt) + disable_blocking (Float32)
		damage, err1 := buf.ReadVarInt()
		blocking, err2 := buf.ReadFloat32()
		if err1 == nil && err2 == nil {
			return fmt.Sprintf("{damage: %d, disable_blocking: %gs}", damage, blocking)
		}

	case ComponentEnchantable:
		// value (VarInt)
		if v, err := buf.ReadVarInt(); err == nil {
			return fmt.Sprintf("{value: %d}", v)
		}

	case ComponentUseCooldown:
		// seconds (Float32) + optional group (String)
		seconds, err := buf.ReadFloat32()
		if err == nil {
			hasGroup, _ := buf.ReadBool()
			if hasGroup {
				if group, err := buf.ReadString(32767); err == nil {
					return fmt.Sprintf("{seconds: %g, group: %q}", seconds, group)
				}
			}
			return fmt.Sprintf("{seconds: %g}", seconds)
		}

	case ComponentFireworks:
		// flight_duration (VarInt) + explosions
		duration, err := buf.ReadVarInt()
		if err == nil {
			explosions, _ := buf.ReadVarInt()
			return fmt.Sprintf("{flight_duration: %d, explosions: %d}", duration, explosions)
		}

	case ComponentAttributeModifiers:
		// count (VarInt) + modifiers
		count, err := buf.ReadVarInt()
		if err == nil {
			return fmt.Sprintf("[%d modifiers]", count)
		}

	case ComponentEnchantments, ComponentStoredEnchantments:
		// count (VarInt) + enchantments
		count, err := buf.ReadVarInt()
		if err == nil {
			return fmt.Sprintf("[%d enchantments]", count)
		}

	case ComponentTooltipDisplay:
		// hide_tooltip (Bool) + hidden_components (VarInt array)
		hideTooltip, err := buf.ReadBool()
		if err == nil {
			hiddenCount, _ := buf.ReadVarInt()
			return fmt.Sprintf("{hide_tooltip: %t, hidden: %d components}", hideTooltip, hiddenCount)
		}

	case ComponentLore:
		// count (VarInt) + lines (NBT strings)
		count, err := buf.ReadVarInt()
		if err == nil {
			return fmt.Sprintf("[%d lines]", count)
		}

	case ComponentMapId:
		// map_id (VarInt)
		if v, err := buf.ReadVarInt(); err == nil {
			return fmt.Sprintf("%d", v)
		}

	case ComponentMapColor, ComponentDyedColor:
		// color (Int32)
		if v, err := buf.ReadInt32(); err == nil {
			return fmt.Sprintf("#%06x", v&0xFFFFFF)
		}

	case ComponentCustomName, ComponentItemName:
		// NBT text component
		reader := nbt.NewReaderFrom(buf.Reader())
		tag, _, err := reader.ReadTag(true) // network format
		if err == nil {
			var tc ns.TextComponent
			if err := tc.UnmarshalNBT(tag); err == nil {
				return fmt.Sprintf("%q", tc.String())
			}
		}
	}

	// Fallback: show hex for unhandled components
	if len(data) > 32 {
		return fmt.Sprintf("0x%x... (%d bytes)", data[:32], len(data))
	}
	return fmt.Sprintf("0x%x", data)
}
