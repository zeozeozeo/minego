package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// synchronizedRegistryIDs lists the registry identifiers that are sent over the
// network from server to client during the configuration phase.
// Source: decompiled RegistryDataLoader.SYNCHRONIZED_REGISTRIES (Minecraft 26.1).
var synchronizedRegistryIDs = []string{
	"minecraft:worldgen/biome",
	"minecraft:chat_type",
	"minecraft:trim_pattern",
	"minecraft:trim_material",
	"minecraft:wolf_variant",
	"minecraft:wolf_sound_variant",
	"minecraft:pig_variant",
	"minecraft:pig_sound_variant",
	"minecraft:frog_variant",
	"minecraft:cat_variant",
	"minecraft:cat_sound_variant",
	"minecraft:cow_variant",
	"minecraft:cow_sound_variant",
	"minecraft:chicken_variant",
	"minecraft:chicken_sound_variant",
	"minecraft:zombie_nautilus_variant",
	"minecraft:painting_variant",
	"minecraft:dimension_type",
	"minecraft:damage_type",
	"minecraft:banner_pattern",
	"minecraft:enchantment",
	"minecraft:jukebox_song",
	"minecraft:instrument",
	"minecraft:test_environment",
	"minecraft:test_instance",
	"minecraft:dialog",
	"minecraft:world_clock",
	"minecraft:timeline",
}

func generateRegistries(registries map[string]RegistryJSON, datapackDir string, outPath string) {
	var sb strings.Builder
	sb.WriteString(generatedFileHeader("registries"))

	// generate registry variables
	sb.WriteString("// Registry instances\nvar (\n")
	for _, name := range sortedKeys(registries) {
		reg := registries[name]
		goName := toGoVarName(name)
		entriesVar := strings.ToLower(goName[:1]) + goName[1:] + "Entries"
		sb.WriteString(fmt.Sprintf("\t%s = newRegistry(%q, %d, %s)\n", goName, name, reg.ProtocolID, entriesVar))
	}
	sb.WriteString(")\n\n")

	// generate ByIdentifier lookup map
	sb.WriteString("// ByIdentifier maps registry identifier strings to registry instances.\n")
	sb.WriteString("var ByIdentifier = map[string]*Registry{\n")
	for _, name := range sortedKeys(registries) {
		goName := toGoVarName(name)
		sb.WriteString(fmt.Sprintf("\t%q: %s,\n", name, goName))
	}
	sb.WriteString("}\n\n")

	// generate SynchronizedRegistryIDs
	sb.WriteString("// SynchronizedRegistryIDs lists registry identifiers sent over the network during configuration.\n")
	sb.WriteString("var SynchronizedRegistryIDs = [...]string{\n")
	for _, id := range synchronizedRegistryIDs {
		sb.WriteString(fmt.Sprintf("\t%q,\n", id))
	}
	sb.WriteString("}\n\n")

	// generate entry maps
	for _, name := range sortedKeys(registries) {
		reg := registries[name]
		goName := toGoVarName(name)
		varName := strings.ToLower(goName[:1]) + goName[1:] + "Entries"

		sb.WriteString(fmt.Sprintf("var %s = map[string]int32{\n", varName))
		for _, entryName := range sortedKeys(reg.Entries) {
			entry := reg.Entries[entryName]
			sb.WriteString(fmt.Sprintf("\t%q: %d,\n", entryName, entry.ProtocolID))
		}
		sb.WriteString("}\n\n")
	}

	// generate SynchronizedEntries: entry names for data-driven registries
	// scanned from the vanilla datapack (data/minecraft/<registry>/*.json)
	sb.WriteString("// SynchronizedEntries maps synchronized registry identifiers to their\n")
	sb.WriteString("// vanilla datapack entry names. For registries that also appear in\n")
	sb.WriteString("// ByIdentifier (static registries), entries come from there instead.\n")
	sb.WriteString("var SynchronizedEntries = map[string][]string{\n")
	for _, regID := range synchronizedRegistryIDs {
		// strip "minecraft:" prefix to get relative path
		relPath := strings.TrimPrefix(regID, "minecraft:")
		dir := filepath.Join(datapackDir, relPath)

		dirEntries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Printf("  warn: no datapack dir for %s: %v\n", regID, err)
			continue
		}

		var names []string
		for _, e := range dirEntries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			name := "minecraft:" + strings.TrimSuffix(e.Name(), ".json")
			names = append(names, name)
		}

		if len(names) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("\t%q: {\n", regID))
		for _, name := range names {
			sb.WriteString(fmt.Sprintf("\t\t%q,\n", name))
		}
		sb.WriteString("\t},\n")
	}
	sb.WriteString("}\n\n")

	writeFile(outPath, sb.String())
}
