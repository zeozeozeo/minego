package main

import (
	"fmt"
	"sort"
	"strings"
)

type mcdumpHardness struct {
	Blocks map[string]struct {
		Hardness     float64 `json:"hardness"`
		RequiresTool bool    `json:"requiresTool"`
	} `json:"blocks"`
}

// generateBlockHardness emits block hardness (destroy time) and requires-correct-tool
// from the mod's blocks.json. Hardness -1 means unbreakable; -2 (from the lookup)
// means the block isn't known.
func generateBlockHardness(mcdumpPath, outPath string) {
	dump := loadJSON[mcdumpHardness](mcdumpPath)

	names := make([]string, 0, len(dump.Blocks))
	for name := range dump.Blocks {
		names = append(names, name)
	}
	sort.Strings(names)
	checkMinCount("block hardness", len(names), 500)

	var sb strings.Builder
	sb.WriteString(generatedFileHeader("blocks"))
	sb.WriteString("// BlockHardness returns the destroy time (hardness) for a block by registry name.\n")
	sb.WriteString("// Returns -1 for unbreakable blocks, 0 for instant break, -2 if not found.\n")
	sb.WriteString("func BlockHardness(name string) float32 {\n")
	sb.WriteString("\tv, ok := blockHardness[name]\n")
	sb.WriteString("\tif !ok { return -2 }\n")
	sb.WriteString("\treturn v\n")
	sb.WriteString("}\n\n")

	sb.WriteString("// BlockRequiresCorrectTool returns whether the block needs the correct tool to mine at full speed.\n")
	sb.WriteString("func BlockRequiresCorrectTool(name string) bool {\n")
	sb.WriteString("\treturn blockRequiresCorrectTool[name]\n")
	sb.WriteString("}\n\n")

	sb.WriteString("var blockHardness = map[string]float32{\n")
	for _, name := range names {
		sb.WriteString(fmt.Sprintf("\t%q: %s,\n", name, formatFloat32(dump.Blocks[name].Hardness)))
	}
	sb.WriteString("}\n\n")

	sb.WriteString("var blockRequiresCorrectTool = map[string]bool{\n")
	for _, name := range names {
		if dump.Blocks[name].RequiresTool {
			sb.WriteString(fmt.Sprintf("\t%q: true,\n", name))
		}
	}
	sb.WriteString("}\n")

	writeFile(outPath, sb.String())
	fmt.Printf("block hardness: %d blocks\n", len(names))
}
