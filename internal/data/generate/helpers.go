package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// Protocol version info is supplied by mcgen for isolated releases. The
// defaults preserve the checked-in 26.2 generation command.
var (
	ProtocolVersion  int32 = 776
	MinecraftVersion       = "26.2"
)

// lowCountWarnings accumulates scrapers that produced implausibly few entries.
var lowCountWarnings []string

// checkMinCount records a sanity-check failure when a scraper produced fewer
// entries than expected. This catches the silent failure mode where Mojang
// refactors the decompiled source layout and a Java-scraping parser quietly
// matches nothing (e.g. "extracted 0 entities") but generation still completes.
// Recorded failures are reported together by reportSanityChecks.
func checkMinCount(label string, got, min int) {
	if got < min {
		lowCountWarnings = append(lowCountWarnings, fmt.Sprintf("%s: scraped %d entries (expected >= %d)", label, got, min))
	}
}

// reportSanityChecks prints any accumulated low-count failures and exits
// non-zero, so a broken scraper fails generation loudly instead of silently
// shipping near-empty data.
func reportSanityChecks() {
	if len(lowCountWarnings) == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, "\n!!! SANITY CHECK FAILED: a parser likely broke (decompiled source layout may have changed):")
	for _, w := range lowCountWarnings {
		fmt.Fprintf(os.Stderr, "  - %s\n", w)
	}
	fmt.Fprintln(os.Stderr, "")
	os.Exit(1)
}

// generatedFileHeader returns the standard header for generated Go files.
func generatedFileHeader(pkg string) string {
	return fmt.Sprintf(`// Code generated for Minecraft %s (Protocol %d); DO NOT EDIT.

package %s

`, MinecraftVersion, ProtocolVersion, pkg)
}

func loadJSON[T any](path string) T {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %v", path, err))
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		panic(fmt.Sprintf("failed to parse %s: %v", path, err))
	}
	return result
}

func writeFile(path, content string) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create dir %s: %v", dir, err))
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		panic(fmt.Sprintf("failed to write %s: %v", path, err))
	}
	fmt.Printf("wrote %s\n", path)
}

func toGoName(id string) string {
	// minecraft:acacia_button -> AcaciaButton
	id = strings.TrimPrefix(id, "minecraft:")
	parts := strings.Split(id, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			// handle special cases like worldgen/biome_source
			subparts := strings.SplitSeq(part, "/")
			for sp := range subparts {
				if len(sp) > 0 {
					result.WriteString(strings.ToUpper(sp[:1]))
					result.WriteString(sp[1:])
				}
			}
		}
	}
	return result.String()
}

func toGoVarName(id string) string {
	name := toGoName(id)
	// handle names starting with numbers
	if len(name) > 0 && unicode.IsDigit(rune(name[0])) {
		name = "N" + name
	}
	return name
}

// loadItems loads item component data. It supports both the legacy single-file
// format (items.json) and the 26.1+ per-item file format (items/ directory).
func loadItems(baseDir string) map[string]ItemJSON {
	// try legacy single-file format first
	legacyPath := filepath.Join(baseDir, "items.json")
	if _, err := os.Stat(legacyPath); err == nil {
		return loadJSON[map[string]ItemJSON](legacyPath)
	}

	// 26.1+ per-item file format: items/<item_name>.json
	itemsDir := filepath.Join(baseDir, "items")
	entries, err := os.ReadDir(itemsDir)
	if err != nil {
		panic(fmt.Sprintf("failed to read items directory %s: %v", itemsDir, err))
	}
	items := make(map[string]ItemJSON, len(entries))
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := "minecraft:" + strings.TrimSuffix(entry.Name(), ".json")
		items[name] = loadJSON[ItemJSON](filepath.Join(itemsDir, entry.Name()))
	}
	return items
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// generateVersion creates version_gen.go with protocol version constants.
func generateVersion(outPath string) {
	var sb strings.Builder
	sb.WriteString(generatedFileHeader("data"))
	sb.WriteString(fmt.Sprintf(`// ProtocolVersion is the Minecraft protocol version this package was generated for.
const ProtocolVersion = %d

// MinecraftVersion is the Minecraft game version this package was generated for.
const MinecraftVersion = "%s"
`, ProtocolVersion, MinecraftVersion))

	writeFile(outPath, sb.String())
}
