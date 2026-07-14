package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateRegistryData reads datapack JSON files for synchronized registries
// and embeds them as raw JSON in Go code. The server converts JSON→NBT at
// startup using the nbt package's JSON conversion.
func generateRegistryData(datapackDir, outPath string) {
	var sb strings.Builder
	sb.WriteString(generatedFileHeader("registries"))
	sb.WriteString("import \"encoding/json\"\n\n")

	sb.WriteString("// SynchronizedRegistryData contains the raw JSON data for each entry\n")
	sb.WriteString("// in each synchronized registry, from the vanilla datapack.\n")
	sb.WriteString("// The server converts these to NBT at startup for S2CRegistryData packets.\n")
	sb.WriteString("var SynchronizedRegistryData = map[string]map[string]json.RawMessage{\n")

	for _, regID := range synchronizedRegistryIDs {
		relPath := strings.TrimPrefix(regID, "minecraft:")
		dir := filepath.Join(datapackDir, relPath)

		dirEntries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		type entry struct {
			name string
			data []byte
		}
		var entries []entry

		for _, e := range dirEntries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			name := "minecraft:" + strings.TrimSuffix(e.Name(), ".json")
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				fmt.Printf("  warn: failed to read %s/%s: %v\n", regID, e.Name(), err)
				continue
			}
			// compact the JSON
			var buf bytes.Buffer
			if err := json.Compact(&buf, data); err != nil {
				entries = append(entries, entry{name, data})
			} else {
				entries = append(entries, entry{name, buf.Bytes()})
			}
		}

		if len(entries) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("\t%q: {\n", regID))
		for _, e := range entries {
			// use backtick raw string to avoid escaping issues
			sb.WriteString(fmt.Sprintf("\t\t%q: json.RawMessage(`%s`),\n", e.name, e.data))
		}
		sb.WriteString("\t},\n")
	}
	sb.WriteString("}\n")

	writeFile(outPath, sb.String())
}
