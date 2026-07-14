package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// tagValueJSON handles both string values and object values like {"id": "minecraft:something", "required": false}.
type tagValueJSON struct {
	ID string
}

func (v *tagValueJSON) UnmarshalJSON(data []byte) error {
	// try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		v.ID = s
		return nil
	}
	// try object with "id" field
	var obj struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	v.ID = obj.ID
	return nil
}

type tagFileJSON struct {
	Values []tagValueJSON `json:"values"`
}

type regDir struct {
	regID string
	path  string
}

// discoverRegistryDirs finds all registry root directories under tagsDir.
// Top-level dirs map directly (e.g., block → minecraft:block).
// Dirs without JSON files (like worldgen/) recurse one level deeper
// (e.g., worldgen/biome → minecraft:worldgen/biome).
func discoverRegistryDirs(tagsDir string) []regDir {
	var result []regDir

	topEntries, err := os.ReadDir(tagsDir)
	if err != nil {
		panic(fmt.Sprintf("failed to read tags dir %s: %v", tagsDir, err))
	}

	for _, e := range topEntries {
		if !e.IsDir() {
			continue
		}
		dirPath := filepath.Join(tagsDir, e.Name())
		if dirHasJSON(dirPath) {
			result = append(result, regDir{"minecraft:" + e.Name(), dirPath})
			continue
		}
		// no direct JSON files; check subdirectories (e.g., worldgen/biome)
		subEntries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		for _, sub := range subEntries {
			if !sub.IsDir() {
				continue
			}
			subPath := filepath.Join(dirPath, sub.Name())
			result = append(result, regDir{"minecraft:" + e.Name() + "/" + sub.Name(), subPath})
		}
	}

	return result
}

// dirHasJSON reports whether dir contains at least one .json file (non-recursive).
func dirHasJSON(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			return true
		}
	}
	return false
}

// generateTagData reads all tag directories under the datapack and generates
// a Go file mapping registry → tag name → []entry name (with tag refs resolved).
func generateTagData(tagsDir, outPath string) {
	registryDirs := discoverRegistryDirs(tagsDir)

	// for each registry, load all tags and resolve references
	allTags := make(map[string]map[string][]string) // regID → tagName → []entryName

	for _, reg := range registryDirs {
		raw := make(map[string][]string) // tag name → raw values (entries + #refs)

		// recursively walk the tag directory for JSON files
		err := filepath.WalkDir(reg.path, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", path, err)
			}
			var tag tagFileJSON
			if err := json.Unmarshal(data, &tag); err != nil {
				return fmt.Errorf("failed to parse %s: %w", path, err)
			}

			// tag name from relative path within the registry root
			rel, _ := filepath.Rel(reg.path, path)
			tagName := "minecraft:" + strings.TrimSuffix(rel, ".json")

			values := make([]string, len(tag.Values))
			for i, v := range tag.Values {
				values[i] = v.ID
			}
			raw[tagName] = values
			return nil
		})
		if err != nil {
			panic(fmt.Sprintf("failed to read tags for %s: %v", reg.regID, err))
		}

		// resolve #tag references recursively
		resolved := make(map[string][]string)
		var resolve func(name string, seen map[string]bool) []string
		resolve = func(name string, seen map[string]bool) []string {
			if entries, ok := resolved[name]; ok {
				return entries
			}
			if seen[name] {
				return nil
			}
			seen[name] = true

			values, ok := raw[name]
			if !ok {
				return nil
			}

			nameSet := make(map[string]bool)
			var entries []string
			for _, v := range values {
				if strings.HasPrefix(v, "#") {
					ref := v[1:]
					for _, entry := range resolve(ref, seen) {
						if !nameSet[entry] {
							nameSet[entry] = true
							entries = append(entries, entry)
						}
					}
				} else {
					if !nameSet[v] {
						nameSet[v] = true
						entries = append(entries, v)
					}
				}
			}
			resolved[name] = entries
			return entries
		}

		for name := range raw {
			resolve(name, make(map[string]bool))
		}

		allTags[reg.regID] = resolved
	}

	// generate output
	var sb strings.Builder
	sb.WriteString(generatedFileHeader("registries"))

	sb.WriteString("// TagData maps registry identifiers to their tags.\n")
	sb.WriteString("// Each tag maps a tag name to a list of entry identifiers.\n")
	sb.WriteString("var TagData = map[string]map[string][]string{\n")

	for _, regID := range sortedKeys(allTags) {
		tags := allTags[regID]
		if len(tags) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("\t%q: {\n", regID))
		for _, tagName := range sortedKeys(tags) {
			entries := tags[tagName]
			// include empty tags — the client may reference them by name
			sb.WriteString(fmt.Sprintf("\t\t%q: {\n", tagName))
			for _, entry := range entries {
				sb.WriteString(fmt.Sprintf("\t\t\t%q,\n", entry))
			}
			sb.WriteString("\t\t},\n")
		}
		sb.WriteString("\t},\n")
	}
	sb.WriteString("}\n")

	writeFile(outPath, sb.String())
}
