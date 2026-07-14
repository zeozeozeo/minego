package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

type tagJSON struct {
	Values []string `json:"values"`
}

func generateItemTags(tagsDir string, registries map[string]RegistryJSON, outPath string) {
	itemRegistry := registries["minecraft:item"]

	// load all tag files
	raw := make(map[string][]string) // tag name -> raw values (items + #refs)
	entries, err := os.ReadDir(tagsDir)
	if err != nil {
		panic(fmt.Sprintf("failed to read tags dir %s: %v", tagsDir, err))
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(tagsDir, entry.Name()))
		if err != nil {
			panic(fmt.Sprintf("failed to read tag file %s: %v", entry.Name(), err))
		}
		var tag tagJSON
		if err := json.Unmarshal(data, &tag); err != nil {
			panic(fmt.Sprintf("failed to parse tag file %s: %v", entry.Name(), err))
		}
		name := "minecraft:" + strings.TrimSuffix(entry.Name(), ".json")
		raw[name] = tag.Values
	}

	// recursively resolve tags into concrete item IDs
	resolved := make(map[string][]int32)
	var resolve func(name string, seen map[string]bool) []int32
	resolve = func(name string, seen map[string]bool) []int32 {
		if ids, ok := resolved[name]; ok {
			return ids
		}
		if seen[name] {
			return nil // circular reference
		}
		seen[name] = true

		values, ok := raw[name]
		if !ok {
			return nil
		}

		idSet := make(map[int32]bool)
		var ids []int32
		for _, v := range values {
			if strings.HasPrefix(v, "#") {
				// nested tag reference
				ref := v[1:]
				for _, id := range resolve(ref, seen) {
					if !idSet[id] {
						idSet[id] = true
						ids = append(ids, id)
					}
				}
			} else {
				entry, ok := itemRegistry.Entries[v]
				if !ok {
					continue
				}
				if !idSet[entry.ProtocolID] {
					idSet[entry.ProtocolID] = true
					ids = append(ids, entry.ProtocolID)
				}
			}
		}
		resolved[name] = ids
		return ids
	}

	for name := range raw {
		resolve(name, make(map[string]bool))
	}

	// build reverse map: item ID -> tags
	reverse := make(map[int32][]string)
	for name, ids := range resolved {
		for _, id := range ids {
			reverse[id] = append(reverse[id], name)
		}
	}
	// sort reverse map values for deterministic output
	for id := range reverse {
		sort.Strings(reverse[id])
	}

	// generate output
	var sb strings.Builder
	sb.WriteString(generatedFileHeader("items"))

	// tag -> item IDs
	sb.WriteString("var itemTagItems = map[string][]int32{\n")
	for _, name := range sortedKeys(resolved) {
		ids := resolved[name]
		if len(ids) == 0 {
			continue
		}
		slices.Sort(ids)
		sb.WriteString(fmt.Sprintf("\t%q: {", name))
		for i, id := range ids {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%d", id))
		}
		sb.WriteString("},\n")
	}
	sb.WriteString("}\n\n")

	// item ID -> tags (reverse)
	sb.WriteString("var itemTagsByItem = map[int32][]string{\n")
	sortedIDs := make([]int32, 0, len(reverse))
	for id := range reverse {
		sortedIDs = append(sortedIDs, id)
	}
	slices.Sort(sortedIDs)
	for _, id := range sortedIDs {
		tags := reverse[id]
		sb.WriteString(fmt.Sprintf("\t%d: {", id))
		for i, t := range tags {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%q", t))
		}
		sb.WriteString("},\n")
	}
	sb.WriteString("}\n")

	writeFile(outPath, sb.String())
}
