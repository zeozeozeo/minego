package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// mcdumpEntity is one entry of mcdump/entities.json.
type mcdumpEntity struct {
	ProtocolID    int     `json:"protocolId"`
	Width         float64 `json:"width"`
	Height        float64 `json:"height"`
	EyeHeight     float64 `json:"eyeHeight"`
	Category      string  `json:"category"`
	TrackingRange int     `json:"trackingRange"`
}

func loadEntities(mcdumpPath string) map[string]mcdumpEntity {
	return loadJSON[map[string]mcdumpEntity](mcdumpPath)
}

// generateEntityHitboxes emits standing hitbox dimensions (width, height, real eye
// height) for every entity type, straight from the mod's entities.json.
func generateEntityHitboxes(mcdumpPath, outPath string) {
	ents := loadEntities(mcdumpPath)

	names := make([]string, 0, len(ents))
	for name := range ents {
		names = append(names, name)
	}
	sort.Strings(names)
	checkMinCount("entity hitboxes", len(names), 100)

	var sb strings.Builder
	sb.WriteString(generatedFileHeader("entities"))
	sb.WriteString(fmt.Sprintf("// Dimensions contains the standing hitbox dimensions for %d entity types.\n", len(names)))
	sb.WriteString("//\n// Eye height of -1 means the entity uses the default formula: height * 0.85.\n")
	sb.WriteString("var dimensions = [...]struct {\n\tIdentifier string\n\tWidth      float32\n\tHeight     float32\n\tEyeHeight  float32\n}{\n")
	for _, name := range names {
		e := ents[name]
		sb.WriteString(fmt.Sprintf("\t{%q, %s, %s, %s},\n",
			name, formatFloat32(e.Width), formatFloat32(e.Height), formatFloat32(e.EyeHeight)))
	}
	sb.WriteString("}\n\n")

	sb.WriteString("var dimensionsByName = map[string]int{\n")
	for i, name := range names {
		sb.WriteString(fmt.Sprintf("\t%q: %d,\n", name, i))
	}
	sb.WriteString("}\n")

	writeFile(outPath, sb.String())
	fmt.Printf("entity hitboxes: %d entities\n", len(names))
}

func formatFloat32(f float64) string {
	s := strconv.FormatFloat(f, 'f', -1, 32)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}
