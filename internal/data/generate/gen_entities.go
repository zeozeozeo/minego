package main

import (
	"fmt"
	"sort"
	"strings"
)

// generateEntities emits the per-entity mob category map from the mod's
// entities.json (the game's own EntityType.getCategory()).
func generateEntities(mcdumpPath, outPath string) {
	ents := loadEntities(mcdumpPath)

	names := make([]string, 0, len(ents))
	for name := range ents {
		names = append(names, name)
	}
	sort.Strings(names)
	checkMinCount("mob categories", len(names), 100)

	var sb strings.Builder
	sb.WriteString(generatedFileHeader("entities"))
	sb.WriteString("var entityCategory = map[string]string{\n")
	for _, name := range names {
		cat := ents[name].Category
		if cat == "" {
			cat = "misc"
		}
		sb.WriteString(fmt.Sprintf("\t%q: %q,\n", name, cat))
	}
	sb.WriteString("}\n")

	writeFile(outPath, sb.String())
	fmt.Printf("mob categories: %d entities\n", len(names))
}
