package main

import (
	"fmt"
	"sort"
	"strings"
)

func generateItems(items map[string]ItemJSON, registries map[string]RegistryJSON, outPath string) {
	itemRegistry := registries["minecraft:item"]

	var sb strings.Builder
	sb.WriteString(generatedFileHeader("items"))

	// generate default components data
	sb.WriteString("// defaultComponents maps item IDs to their default components.\n")
	sb.WriteString("var defaultComponents = map[int32]*Components{\n")

	for _, itemName := range sortedKeys(items) {
		item := items[itemName]
		itemID := itemRegistry.Entries[itemName].ProtocolID

		sb.WriteString(fmt.Sprintf("\t%d: { // %s\n", itemID, itemName))
		generateComponentsLiteral(&sb, item.Components, "\t\t")
		sb.WriteString("\t},\n")
	}
	sb.WriteString("}\n")

	writeFile(outPath, sb.String())
}

func generateComponentsLiteral(sb *strings.Builder, components map[string]any, indent string) {
	for _, key := range sortedKeys(components) {
		value := components[key]
		goField := componentKeyToGoField(key)
		if goField == "" {
			continue
		}

		switch key {
		case "minecraft:max_stack_size":
			if v, ok := value.(float64); ok {
				sb.WriteString(fmt.Sprintf("%sMaxStackSize: %d,\n", indent, int32(v)))
			}
		case "minecraft:damage":
			if v, ok := value.(float64); ok {
				sb.WriteString(fmt.Sprintf("%sDamage: %d,\n", indent, int32(v)))
			}
		case "minecraft:max_damage":
			if v, ok := value.(float64); ok {
				sb.WriteString(fmt.Sprintf("%sMaxDamage: %d,\n", indent, int32(v)))
			}
		case "minecraft:repair_cost":
			if v, ok := value.(float64); ok {
				sb.WriteString(fmt.Sprintf("%sRepairCost: %d,\n", indent, int32(v)))
			}
		case "minecraft:rarity":
			if v, ok := value.(string); ok {
				sb.WriteString(fmt.Sprintf("%sRarity: %q,\n", indent, v))
			}
		case "minecraft:break_sound":
			if v, ok := value.(string); ok {
				sb.WriteString(fmt.Sprintf("%sBreakSound: %q,\n", indent, v))
			}
		case "minecraft:item_model":
			if v, ok := value.(string); ok {
				sb.WriteString(fmt.Sprintf("%sItemModel: %q,\n", indent, v))
			}
		case "minecraft:instrument":
			if v, ok := value.(string); ok {
				sb.WriteString(fmt.Sprintf("%sInstrument: %q,\n", indent, v))
			}
		case "minecraft:jukebox_playable":
			if v, ok := value.(string); ok {
				sb.WriteString(fmt.Sprintf("%sJukeboxPlayable: %q,\n", indent, v))
			}
		case "minecraft:provides_banner_patterns":
			if v, ok := value.(string); ok {
				sb.WriteString(fmt.Sprintf("%sProvidesBannerPatterns: %q,\n", indent, v))
			}
		case "minecraft:provides_trim_material":
			if v, ok := value.(string); ok {
				sb.WriteString(fmt.Sprintf("%sProvidesTrimMaterial: %q,\n", indent, v))
			}
		case "minecraft:damage_type":
			if v, ok := value.(string); ok {
				sb.WriteString(fmt.Sprintf("%sDamageType: %q,\n", indent, v))
			}
		case "minecraft:food":
			if m, ok := value.(map[string]any); ok {
				sb.WriteString(fmt.Sprintf("%sFood: &Food{\n", indent))
				if n, ok := m["nutrition"].(float64); ok {
					sb.WriteString(fmt.Sprintf("%s\tNutrition: %d,\n", indent, int32(n)))
				}
				if s, ok := m["saturation"].(float64); ok {
					sb.WriteString(fmt.Sprintf("%s\tSaturation: %v,\n", indent, s))
				}
				sb.WriteString(fmt.Sprintf("%s},\n", indent))
			}
		case "minecraft:tool":
			if m, ok := value.(map[string]any); ok {
				sb.WriteString(fmt.Sprintf("%sTool: &Tool{\n", indent))
				if rules, ok := m["rules"].([]any); ok && len(rules) > 0 {
					sb.WriteString(fmt.Sprintf("%s\tRules: []ToolRule{\n", indent))
					for _, r := range rules {
						if rule, ok := r.(map[string]any); ok {
							sb.WriteString(fmt.Sprintf("%s\t\t{", indent))
							if b, ok := rule["blocks"].(string); ok {
								sb.WriteString(fmt.Sprintf("Blocks: %q, ", b))
							}
							if s, ok := rule["speed"].(float64); ok {
								sb.WriteString(fmt.Sprintf("Speed: %v, ", s))
							}
							if c, ok := rule["correct_for_drops"].(bool); ok {
								sb.WriteString(fmt.Sprintf("CorrectForDrops: %v", c))
							}
							sb.WriteString("},\n")
						}
					}
					sb.WriteString(fmt.Sprintf("%s\t},\n", indent))
				}
				sb.WriteString(fmt.Sprintf("%s},\n", indent))
			}
		case "minecraft:weapon":
			if m, ok := value.(map[string]any); ok {
				sb.WriteString(fmt.Sprintf("%sWeapon: &Weapon{\n", indent))
				if d, ok := m["disable_blocking_for_seconds"].(float64); ok {
					sb.WriteString(fmt.Sprintf("%s\tDisableBlockingForSeconds: %v,\n", indent, d))
				}
				if i, ok := m["item_damage_per_attack"].(float64); ok {
					sb.WriteString(fmt.Sprintf("%s\tItemDamagePerAttack: %d,\n", indent, int32(i)))
				}
				sb.WriteString(fmt.Sprintf("%s},\n", indent))
			}
		case "minecraft:enchantable":
			if m, ok := value.(map[string]any); ok {
				if v, ok := m["value"].(float64); ok {
					sb.WriteString(fmt.Sprintf("%sEnchantable: &Enchantable{Value: %d},\n", indent, int32(v)))
				}
			}
		case "minecraft:repairable":
			if m, ok := value.(map[string]any); ok {
				if items, ok := m["items"].(string); ok {
					sb.WriteString(fmt.Sprintf("%sRepairable: &Repairable{Items: %q},\n", indent, items))
				}
			}
		case "minecraft:item_name":
			if m, ok := value.(map[string]any); ok {
				if t, ok := m["translate"].(string); ok {
					sb.WriteString(fmt.Sprintf("%sItemName: &ItemNameComponent{Translate: %q},\n", indent, t))
				}
			}
		case "minecraft:fireworks":
			if m, ok := value.(map[string]any); ok {
				if fd, ok := m["flight_duration"].(float64); ok {
					sb.WriteString(fmt.Sprintf("%sFireworks: &Fireworks{FlightDuration: %d},\n", indent, int32(fd)))
				}
			}
		case "minecraft:use_cooldown":
			if m, ok := value.(map[string]any); ok {
				if s, ok := m["seconds"].(float64); ok {
					sb.WriteString(fmt.Sprintf("%sUseCooldown: &UseCooldown{Seconds: %v},\n", indent, s))
				}
			}
		case "minecraft:use_remainder":
			if m, ok := value.(map[string]any); ok {
				sb.WriteString(fmt.Sprintf("%sUseRemainder: &UseRemainder{\n", indent))
				if c, ok := m["count"].(float64); ok {
					sb.WriteString(fmt.Sprintf("%s\tCount: %d,\n", indent, int32(c)))
				}
				if id, ok := m["id"].(string); ok {
					sb.WriteString(fmt.Sprintf("%s\tID: %q,\n", indent, id))
				}
				sb.WriteString(fmt.Sprintf("%s},\n", indent))
			}
		case "minecraft:damage_resistant":
			if m, ok := value.(map[string]any); ok {
				if t, ok := m["types"].(string); ok {
					sb.WriteString(fmt.Sprintf("%sDamageResistant: &DamageResistant{Types: %q},\n", indent, t))
				}
			}
		case "minecraft:map_color":
			if v, ok := value.(float64); ok {
				sb.WriteString(fmt.Sprintf("%sMapColor: %d,\n", indent, int32(v)))
			}
		case "minecraft:ominous_bottle_amplifier":
			if v, ok := value.(float64); ok {
				sb.WriteString(fmt.Sprintf("%sOminousBottleAmplifier: %d,\n", indent, int32(v)))
			}
		case "minecraft:potion_duration_scale":
			if v, ok := value.(float64); ok {
				sb.WriteString(fmt.Sprintf("%sPotionDurationScale: %v,\n", indent, v))
			}
		case "minecraft:minimum_attack_charge":
			if v, ok := value.(float64); ok {
				sb.WriteString(fmt.Sprintf("%sMinimumAttackCharge: %v,\n", indent, v))
			}
		case "minecraft:glider":
			// marker component
			sb.WriteString(fmt.Sprintf("%sGlider: true,\n", indent))
		case "minecraft:attribute_modifiers":
			if arr, ok := value.([]any); ok && len(arr) > 0 {
				sb.WriteString(fmt.Sprintf("%sAttributeModifiers: []AttributeModifier{\n", indent))
				for _, entry := range arr {
					if m, ok := entry.(map[string]any); ok {
						sb.WriteString(fmt.Sprintf("%s\t{", indent))
						if t, ok := m["type"].(string); ok {
							sb.WriteString(fmt.Sprintf("Type: %q, ", t))
						}
						if a, ok := m["amount"].(float64); ok {
							sb.WriteString(fmt.Sprintf("Amount: %v, ", a))
						}
						if id, ok := m["id"].(string); ok {
							sb.WriteString(fmt.Sprintf("ID: %q, ", id))
						}
						if op, ok := m["operation"].(string); ok {
							sb.WriteString(fmt.Sprintf("Operation: %q, ", op))
						}
						if s, ok := m["slot"].(string); ok {
							sb.WriteString(fmt.Sprintf("Slot: %q", s))
						}
						sb.WriteString("},\n")
					}
				}
				sb.WriteString(fmt.Sprintf("%s},\n", indent))
			}
		}
	}
}

func generateComponentTypes(registries map[string]RegistryJSON, outPath string) {
	componentRegistry := registries["minecraft:data_component_type"]

	var sb strings.Builder
	sb.WriteString(generatedFileHeader("items"))

	// component type constants
	sb.WriteString("// Component type protocol IDs from minecraft:data_component_type registry.\n")
	sb.WriteString("// https://minecraft.wiki/w/Java_Edition_protocol/Slot_data#Structured_components\n")
	sb.WriteString("const (\n")

	// find max ID for the constant
	var maxID int32
	for _, name := range sortedKeys(componentRegistry.Entries) {
		entry := componentRegistry.Entries[name]
		goName := "Component" + toGoVarName(name)
		// use untyped constants so they work with VarInt comparisons
		sb.WriteString(fmt.Sprintf("\t%s = %d\n", goName, entry.ProtocolID))
		if entry.ProtocolID > maxID {
			maxID = entry.ProtocolID
		}
	}
	sb.WriteString(fmt.Sprintf("\n\tMaxComponentID = %d\n", maxID))
	sb.WriteString(")\n")

	writeFile(outPath, sb.String())
}

func generateComponentCodecs(registries map[string]RegistryJSON, metadataPath, outPath string) {
	componentRegistry := registries["minecraft:data_component_type"]
	metadata := loadJSON[ComponentMetadataFile](metadataPath)

	// collection phase

	type simpleCodec struct {
		constName string
		goField   string
		goType    string
		wireType  string
	}
	var varIntCodecs, float32Codecs, stringCodecs, emptyCodecs []simpleCodec

	var varIntPassthrough, boolPassthrough, stringPassthrough, emptyPassthrough []string
	var int32Passthrough, nbtPassthrough, holderSetPassthrough, slotListPassthrough, slotPassthrough []string
	var entityVariants []string

	type structCodecInfo struct {
		constName string
		goField   string
		typeName  string
		fields    []WireField
	}
	var structCodecs []structCodecInfo

	for name, meta := range metadata.Components {
		constName := "Component" + toGoVarName(name)
		if _, ok := componentRegistry.Entries[name]; !ok {
			continue
		}

		if meta.Passthrough {
			switch meta.WireType {
			case "varint":
				varIntPassthrough = append(varIntPassthrough, constName)
			case "bool":
				boolPassthrough = append(boolPassthrough, constName)
			case "identifier":
				stringPassthrough = append(stringPassthrough, constName)
			case "empty":
				emptyPassthrough = append(emptyPassthrough, constName)
			case "int32":
				int32Passthrough = append(int32Passthrough, constName)
			case "nbt":
				nbtPassthrough = append(nbtPassthrough, constName)
			case "holderSet":
				holderSetPassthrough = append(holderSetPassthrough, constName)
			case "slotList":
				slotListPassthrough = append(slotListPassthrough, constName)
			case "slot":
				slotPassthrough = append(slotPassthrough, constName)
			}
		} else if meta.WireType == "struct" && len(meta.WireFormat) > 0 && meta.GoField != "" {
			typeName := strings.TrimPrefix(meta.GoType, "*")
			structCodecs = append(structCodecs, structCodecInfo{
				constName: constName,
				goField:   meta.GoField,
				typeName:  typeName,
				fields:    meta.WireFormat,
			})
		} else if meta.GoField != "" {
			sc := simpleCodec{
				constName: constName,
				goField:   meta.GoField,
				goType:    meta.GoType,
				wireType:  meta.WireType,
			}
			switch meta.WireType {
			case "varint":
				varIntCodecs = append(varIntCodecs, sc)
			case "float32":
				float32Codecs = append(float32Codecs, sc)
			case "identifier":
				stringCodecs = append(stringCodecs, sc)
			case "empty":
				emptyCodecs = append(emptyCodecs, sc)
			}
		}
	}

	// auto-derive entity variant components from registry
	for name := range componentRegistry.Entries {
		stripped := strings.TrimPrefix(name, "minecraft:")
		if strings.Contains(stripped, "/") {
			constName := "Component" + toGoVarName(name)
			entityVariants = append(entityVariants, constName)
		}
	}

	// sort for deterministic output
	sort.Slice(varIntCodecs, func(i, j int) bool { return varIntCodecs[i].constName < varIntCodecs[j].constName })
	sort.Slice(float32Codecs, func(i, j int) bool { return float32Codecs[i].constName < float32Codecs[j].constName })
	sort.Slice(stringCodecs, func(i, j int) bool { return stringCodecs[i].constName < stringCodecs[j].constName })
	sort.Slice(emptyCodecs, func(i, j int) bool { return emptyCodecs[i].constName < emptyCodecs[j].constName })
	sort.Slice(structCodecs, func(i, j int) bool { return structCodecs[i].constName < structCodecs[j].constName })
	sort.Strings(varIntPassthrough)
	sort.Strings(boolPassthrough)
	sort.Strings(stringPassthrough)
	sort.Strings(emptyPassthrough)
	sort.Strings(int32Passthrough)
	sort.Strings(nbtPassthrough)
	sort.Strings(holderSetPassthrough)
	sort.Strings(slotListPassthrough)
	sort.Strings(slotPassthrough)
	sort.Strings(entityVariants)

	// output phase

	var sb strings.Builder
	sb.WriteString(generatedFileHeader("items"))

	// imports needed for struct codecs
	if len(structCodecs) > 0 {
		needsSlices := false
		for _, sc := range structCodecs {
			for _, f := range sc.fields {
				if f.GoField != "" && f.Type == "varintArray" {
					needsSlices = true
				}
			}
		}
		sb.WriteString("import (\n")
		if needsSlices {
			sb.WriteString("\t\"slices\"\n\n")
		}
		sb.WriteString("\tns \"github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures\"\n")
		sb.WriteString(")\n\n")
	}

	sb.WriteString("// Auto-generated codec registrations.\n\n")
	sb.WriteString("func init() {\n")

	// simple codec registrations
	if len(varIntCodecs) > 0 {
		sb.WriteString("\t// Simple VarInt codecs\n")
		for _, sc := range varIntCodecs {
			fmt.Fprintf(&sb, "\tRegisterCodec(%s, &varIntCodec{\n", sc.constName)
			fmt.Fprintf(&sb, "\t\tget: func(c *Components) int32 { return c.%s },\n", sc.goField)
			fmt.Fprintf(&sb, "\t\tset: func(c *Components, v int32) { c.%s = v },\n", sc.goField)
			sb.WriteString("\t})\n")
		}
		sb.WriteString("\n")
	}

	if len(float32Codecs) > 0 {
		sb.WriteString("\t// Simple Float32 codecs\n")
		for _, sc := range float32Codecs {
			fmt.Fprintf(&sb, "\tRegisterCodec(%s, &float32Codec{\n", sc.constName)
			fmt.Fprintf(&sb, "\t\tget: func(c *Components) float64 { return c.%s },\n", sc.goField)
			fmt.Fprintf(&sb, "\t\tset: func(c *Components, v float64) { c.%s = v },\n", sc.goField)
			sb.WriteString("\t})\n")
		}
		sb.WriteString("\n")
	}

	if len(stringCodecs) > 0 {
		sb.WriteString("\t// Simple String/Identifier codecs\n")
		for _, sc := range stringCodecs {
			fmt.Fprintf(&sb, "\tRegisterCodec(%s, &stringCodec{\n", sc.constName)
			fmt.Fprintf(&sb, "\t\tget: func(c *Components) string { return c.%s },\n", sc.goField)
			fmt.Fprintf(&sb, "\t\tset: func(c *Components, v string) { c.%s = v },\n", sc.goField)
			sb.WriteString("\t})\n")
		}
		sb.WriteString("\n")
	}

	if len(emptyCodecs) > 0 {
		sb.WriteString("\t// Empty marker codecs (bool flags)\n")
		for _, sc := range emptyCodecs {
			fmt.Fprintf(&sb, "\tRegisterCodec(%s, &emptyMarkerCodec{\n", sc.constName)
			fmt.Fprintf(&sb, "\t\tget: func(c *Components) bool { return c.%s },\n", sc.goField)
			fmt.Fprintf(&sb, "\t\tset: func(c *Components, v bool) { c.%s = v },\n", sc.goField)
			sb.WriteString("\t})\n")
		}
		sb.WriteString("\n")
	}

	// struct codec registrations
	if len(structCodecs) > 0 {
		sb.WriteString("\t// Struct codecs\n")
		for _, sc := range structCodecs {
			fmt.Fprintf(&sb, "\tRegisterCodec(%s, gen%sCodec{})\n", sc.constName, sc.typeName)
		}
		sb.WriteString("\n")
	}

	// passthrough registrations
	writePassthroughList(&sb, "VarInt passthrough", varIntPassthrough, "registerVarIntPassthrough")
	writePassthroughList(&sb, "Bool passthrough", boolPassthrough, "registerBoolPassthrough")
	writePassthroughList(&sb, "String passthrough", stringPassthrough, "registerStringPassthrough")
	writePassthroughList(&sb, "Empty passthrough", emptyPassthrough, "registerEmptyPassthrough")
	writePassthroughList(&sb, "Int32 passthrough", int32Passthrough, "registerInt32Passthrough")
	writePassthroughList(&sb, "NBT passthrough", nbtPassthrough, "registerNBTPassthrough")
	writePassthroughList(&sb, "HolderSet passthrough", holderSetPassthrough, "registerHolderSetPassthrough")
	writePassthroughList(&sb, "SlotList passthrough", slotListPassthrough, "registerSlotListPassthrough")
	writePassthroughList(&sb, "Slot passthrough", slotPassthrough, "registerSlotPassthrough")
	writePassthroughList(&sb, "Entity variant (VarInt) passthrough", entityVariants, "registerVarIntPassthrough")

	sb.WriteString("}\n\n")

	// struct codec implementations
	for _, sc := range structCodecs {
		generateStructCodec(&sb, sc.goField, sc.typeName, sc.fields)
	}

	writeFile(outPath, sb.String())
}

func writePassthroughList(sb *strings.Builder, comment string, ids []string, registerFn string) {
	if len(ids) == 0 {
		return
	}
	fmt.Fprintf(sb, "\t// %s\n", comment)
	sb.WriteString("\tfor _, id := range []int32{\n")
	for _, id := range ids {
		fmt.Fprintf(sb, "\t\t%s,\n", id)
	}
	sb.WriteString("\t} {\n")
	fmt.Fprintf(sb, "\t\t%s(id)\n", registerFn)
	sb.WriteString("\t}\n\n")
}

func componentKeyToGoField(key string) string {
	switch key {
	case "minecraft:max_stack_size":
		return "MaxStackSize"
	case "minecraft:damage":
		return "Damage"
	case "minecraft:max_damage":
		return "MaxDamage"
	case "minecraft:repair_cost":
		return "RepairCost"
	case "minecraft:rarity":
		return "Rarity"
	case "minecraft:break_sound":
		return "BreakSound"
	case "minecraft:item_model":
		return "ItemModel"
	case "minecraft:food":
		return "Food"
	case "minecraft:tool":
		return "Tool"
	case "minecraft:weapon":
		return "Weapon"
	case "minecraft:enchantable":
		return "Enchantable"
	case "minecraft:repairable":
		return "Repairable"
	case "minecraft:item_name":
		return "ItemName"
	case "minecraft:instrument":
		return "Instrument"
	case "minecraft:jukebox_playable":
		return "JukeboxPlayable"
	case "minecraft:provides_banner_patterns":
		return "ProvidesBannerPatterns"
	case "minecraft:provides_trim_material":
		return "ProvidesTrimMaterial"
	case "minecraft:damage_type":
		return "DamageType"
	case "minecraft:fireworks":
		return "Fireworks"
	case "minecraft:use_cooldown":
		return "UseCooldown"
	case "minecraft:use_remainder":
		return "UseRemainder"
	case "minecraft:damage_resistant":
		return "DamageResistant"
	case "minecraft:map_color":
		return "MapColor"
	case "minecraft:ominous_bottle_amplifier":
		return "OminousBottleAmplifier"
	case "minecraft:potion_duration_scale":
		return "PotionDurationScale"
	case "minecraft:minimum_attack_charge":
		return "MinimumAttackCharge"
	case "minecraft:glider":
		return "Glider"
	case "minecraft:attribute_modifiers":
		return "AttributeModifiers"
	// skip these for now as they're empty or complex
	case "minecraft:enchantments",
		"minecraft:lore",
		"minecraft:swing_animation",
		"minecraft:tooltip_display",
		"minecraft:use_effects",
		"minecraft:consumable",
		"minecraft:container",
		"minecraft:stored_enchantments",
		"minecraft:potion_contents",
		"minecraft:bundle_contents",
		"minecraft:charged_projectiles",
		"minecraft:debug_stick_state",
		"minecraft:entity_data",
		"minecraft:bucket_entity_data",
		"minecraft:block_entity_data",
		"minecraft:block_state",
		"minecraft:bees",
		"minecraft:lock",
		"minecraft:container_loot",
		"minecraft:pot_decorations",
		"minecraft:writable_book_content",
		"minecraft:written_book_content",
		"minecraft:trim",
		"minecraft:suspicious_stew_effects",
		"minecraft:banner_patterns",
		"minecraft:base_color",
		"minecraft:profile",
		"minecraft:note_block_sound",
		"minecraft:lodestone_tracker",
		"minecraft:firework_explosion",
		"minecraft:map_decorations",
		"minecraft:map_id",
		"minecraft:map_post_processing",
		"minecraft:recipes",
		"minecraft:dyed_color",
		"minecraft:creative_slot_lock",
		"minecraft:intangible_projectile",
		"minecraft:custom_data",
		"minecraft:custom_model_data",
		"minecraft:custom_name",
		"minecraft:enchantment_glint_override",
		"minecraft:death_protection",
		"minecraft:blocks_attacks",
		"minecraft:kinetic_weapon",
		"minecraft:piercing_weapon",
		"minecraft:attack_range",
		"minecraft:equippable",
		"minecraft:unbreakable",
		"minecraft:can_break",
		"minecraft:can_place_on":
		return ""
	default:
		return ""
	}
}

// generateStructCodec generates a full ComponentCodec implementation for a pointer-to-struct component.
func generateStructCodec(sb *strings.Builder, goField, typeName string, fields []WireField) {
	codec := "gen" + typeName + "Codec"

	fmt.Fprintf(sb, "type %s struct{}\n\n", codec)

	// DecodeWire
	fmt.Fprintf(sb, "func (%s) DecodeWire(buf *ns.PacketBuffer) ([]byte, error) {\n", codec)
	sb.WriteString("\tw := ns.NewWriter()\n")
	for _, f := range fields {
		writeDecodeWireField(sb, f)
	}
	sb.WriteString("\treturn w.Bytes(), nil\n}\n\n")

	// Apply
	fmt.Fprintf(sb, "func (%s) Apply(c *Components, data []byte) error {\n", codec)
	sb.WriteString("\tbuf := ns.NewReader(data)\n")
	fmt.Fprintf(sb, "\ts := &%s{}\n", typeName)
	for _, f := range fields {
		if f.GoField == "" {
			break // trailing unmapped fields are skipped
		}
		writeApplyField(sb, f)
	}
	fmt.Fprintf(sb, "\tc.%s = s\n", goField)
	sb.WriteString("\treturn nil\n}\n\n")

	// Clear
	fmt.Fprintf(sb, "func (%s) Clear(c *Components) { c.%s = nil }\n\n", codec, goField)

	// Differs
	hasSlice := false
	for _, f := range fields {
		if f.GoField != "" && f.Type == "varintArray" {
			hasSlice = true
			break
		}
	}
	fmt.Fprintf(sb, "func (%s) Differs(c, defaults *Components) (bool, bool) {\n", codec)
	fmt.Fprintf(sb, "\tcHas := c.%s != nil\n\tdHas := defaults.%s != nil\n", goField, goField)
	sb.WriteString("\tif cHas != dHas {\n\t\treturn true, cHas\n\t}\n")
	sb.WriteString("\tif cHas && dHas {\n")
	if hasSlice {
		for _, f := range fields {
			if f.GoField == "" {
				continue
			}
			if f.Type == "varintArray" {
				fmt.Fprintf(sb, "\t\tif !slices.Equal(c.%s.%s, defaults.%s.%s) {\n\t\t\treturn true, true\n\t\t}\n",
					goField, f.GoField, goField, f.GoField)
			} else {
				fmt.Fprintf(sb, "\t\tif c.%s.%s != defaults.%s.%s {\n\t\t\treturn true, true\n\t\t}\n",
					goField, f.GoField, goField, f.GoField)
			}
		}
		sb.WriteString("\t\treturn false, true\n")
	} else {
		fmt.Fprintf(sb, "\t\treturn *c.%s != *defaults.%s, true\n", goField, goField)
	}
	sb.WriteString("\t}\n\treturn false, false\n}\n\n")

	// Encode
	fmt.Fprintf(sb, "func (%s) Encode(c *Components) ([]byte, error) {\n", codec)
	fmt.Fprintf(sb, "\tif c.%s == nil {\n\t\treturn nil, nil\n\t}\n", goField)
	sb.WriteString("\tw := ns.NewWriter()\n")
	for _, f := range fields {
		writeEncodeField(sb, goField, f)
	}
	sb.WriteString("\treturn w.Bytes(), nil\n}\n\n")
}

func writeDecodeWireField(sb *strings.Builder, f WireField) {
	switch f.Type {
	case "varint":
		sb.WriteString("\tif err := w.CopyVarInt(buf); err != nil {\n\t\treturn nil, err\n\t}\n")
	case "float32":
		sb.WriteString("\tif err := w.CopyFloat32(buf); err != nil {\n\t\treturn nil, err\n\t}\n")
	case "bool":
		sb.WriteString("\tif err := w.CopyBool(buf); err != nil {\n\t\treturn nil, err\n\t}\n")
	case "varintArray":
		sb.WriteString("\t{\n")
		sb.WriteString("\t\tcount, err := buf.ReadVarInt()\n")
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn nil, err\n\t\t}\n")
		sb.WriteString("\t\tw.WriteVarInt(count)\n")
		sb.WriteString("\t\tfor range int(count) {\n")
		sb.WriteString("\t\t\tif err := w.CopyVarInt(buf); err != nil {\n\t\t\t\treturn nil, err\n\t\t\t}\n")
		sb.WriteString("\t\t}\n\t}\n")
	case "optionalIdentifier":
		sb.WriteString("\t{\n")
		sb.WriteString("\t\tpresent, err := buf.ReadBool()\n")
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn nil, err\n\t\t}\n")
		sb.WriteString("\t\tw.WriteBool(present)\n")
		sb.WriteString("\t\tif present {\n")
		sb.WriteString("\t\t\tif err := w.CopyString(buf, maxStringLen); err != nil {\n\t\t\t\treturn nil, err\n\t\t\t}\n")
		sb.WriteString("\t\t}\n\t}\n")
	case "fireworkExplosionList":
		sb.WriteString("\t{\n")
		sb.WriteString("\t\tcount, err := buf.ReadVarInt()\n")
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn nil, err\n\t\t}\n")
		sb.WriteString("\t\tw.WriteVarInt(count)\n")
		sb.WriteString("\t\tfor range int(count) {\n")
		sb.WriteString("\t\t\tif err := copyFireworkExplosion(buf, w); err != nil {\n\t\t\t\treturn nil, err\n\t\t\t}\n")
		sb.WriteString("\t\t}\n\t}\n")
	}
}

func writeApplyField(sb *strings.Builder, f WireField) {
	switch f.Type {
	case "varint":
		sb.WriteString("\t{\n\t\tv, err := buf.ReadVarInt()\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		fmt.Fprintf(sb, "\t\ts.%s = int32(v)\n\t}\n", f.GoField)
	case "float32":
		sb.WriteString("\t{\n\t\tv, err := buf.ReadFloat32()\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		fmt.Fprintf(sb, "\t\ts.%s = float64(v)\n\t}\n", f.GoField)
	case "bool":
		sb.WriteString("\t{\n\t\tv, err := buf.ReadBool()\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		fmt.Fprintf(sb, "\t\ts.%s = bool(v)\n\t}\n", f.GoField)
	case "varintArray":
		sb.WriteString("\t{\n\t\tcount, err := buf.ReadVarInt()\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		sb.WriteString("\t\tarr := make([]int32, 0, count)\n")
		sb.WriteString("\t\tfor range int(count) {\n")
		sb.WriteString("\t\t\tv, err := buf.ReadVarInt()\n\t\t\tif err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
		sb.WriteString("\t\t\tarr = append(arr, int32(v))\n\t\t}\n")
		fmt.Fprintf(sb, "\t\ts.%s = arr\n\t}\n", f.GoField)
	}
}

func writeEncodeField(sb *strings.Builder, goField string, f WireField) {
	if f.GoField != "" {
		switch f.Type {
		case "varint":
			fmt.Fprintf(sb, "\tw.WriteVarInt(ns.VarInt(c.%s.%s))\n", goField, f.GoField)
		case "float32":
			fmt.Fprintf(sb, "\tw.WriteFloat32(ns.Float32(c.%s.%s))\n", goField, f.GoField)
		case "bool":
			fmt.Fprintf(sb, "\tw.WriteBool(ns.Boolean(c.%s.%s))\n", goField, f.GoField)
		case "varintArray":
			fmt.Fprintf(sb, "\tw.WriteVarInt(ns.VarInt(len(c.%s.%s)))\n", goField, f.GoField)
			fmt.Fprintf(sb, "\tfor _, v := range c.%s.%s {\n\t\tw.WriteVarInt(ns.VarInt(v))\n\t}\n", goField, f.GoField)
		}
	} else {
		// unmapped field: write zero/default value
		switch f.Type {
		case "varint":
			sb.WriteString("\tw.WriteVarInt(0)\n")
		case "float32":
			sb.WriteString("\tw.WriteFloat32(0)\n")
		case "bool":
			sb.WriteString("\tw.WriteBool(false)\n")
		case "optionalIdentifier":
			sb.WriteString("\tw.WriteBool(false)\n")
		case "fireworkExplosionList":
			sb.WriteString("\tw.WriteVarInt(0)\n")
		}
	}
}
