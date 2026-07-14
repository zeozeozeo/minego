package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// EntityMetadataJSON represents the entity_metadata.include.json structure.
type EntityMetadataJSON struct {
	Serializers map[string]SerializerJSON `json:"serializers"`
	Entities    map[string]EntityJSON     `json:"entities"`
}

type SerializerJSON struct {
	Name     string `json:"name"`
	WireType string `json:"wireType"`
}

type EntityJSON struct {
	Parent string      `json:"parent"`
	Fields []FieldJSON `json:"fields"`
}

type FieldJSON struct {
	Index       int    `json:"index"`
	Name        string `json:"name"`
	Serializer  int    `json:"serializer"`
	GoField     string `json:"goField"`
	GoType      string `json:"goType"`
	Description string `json:"description,omitempty"`
	Passthrough bool   `json:"passthrough,omitempty"`
}

func generateEntityMetadata(includePath, outDir string) {
	data, err := os.ReadFile(includePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not read %s: %v\n", includePath, err)
		return
	}

	var metadata EntityMetadataJSON
	if err := json.Unmarshal(data, &metadata); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing entity metadata JSON: %v\n", err)
		return
	}

	generateSerializers(metadata.Serializers, filepath.Join(outDir, "entity_metadata_serializers_gen.go"))
	generateEntityMetadataTypes(metadata, filepath.Join(outDir, "entity_metadata_gen.go"))
}

func generateSerializers(serializers map[string]SerializerJSON, outPath string) {
	var sb strings.Builder
	sb.WriteString(generatedFileHeader("entities"))

	sb.WriteString("// Entity metadata serializer type IDs.\n")
	sb.WriteString("const (\n")

	// sort by ID (numeric)
	ids := make([]int, 0, len(serializers))
	for idStr := range serializers {
		var id int
		fmt.Sscanf(idStr, "%d", &id)
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		s := serializers[fmt.Sprintf("%d", id)]
		sb.WriteString(fmt.Sprintf("\tSerializer%s = %d\n", s.Name, id))
	}
	sb.WriteString(")\n\n")

	sb.WriteString("// serializerNames maps serializer IDs to names.\n")
	sb.WriteString("var serializerNames = map[int32]string{\n")
	for _, id := range ids {
		s := serializers[fmt.Sprintf("%d", id)]
		sb.WriteString(fmt.Sprintf("\t%d: %q,\n", id, s.Name))
	}
	sb.WriteString("}\n\n")

	sb.WriteString("// serializerWireTypes maps serializer IDs to wire types.\n")
	sb.WriteString("var serializerWireTypes = map[int32]string{\n")
	for _, id := range ids {
		s := serializers[fmt.Sprintf("%d", id)]
		sb.WriteString(fmt.Sprintf("\t%d: %q,\n", id, s.WireType))
	}
	sb.WriteString("}\n")

	writeFile(outPath, sb.String())
}

func generateEntityMetadataTypes(metadata EntityMetadataJSON, outPath string) {
	var sb strings.Builder
	sb.WriteString(generatedFileHeader("entities"))

	// collect all fields for each entity, including inherited ones
	resolvedEntities := resolveEntityHierarchy(metadata)

	// generate field index constants per entity
	sb.WriteString("// Entity metadata field indices.\n")

	entityNames := sortedKeys(metadata.Entities)
	for _, name := range entityNames {
		entity := metadata.Entities[name]
		if len(entity.Fields) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("\n// %s metadata indices\n", name))
		sb.WriteString("const (\n")
		for _, f := range entity.Fields {
			constName := fmt.Sprintf("%sIndex%s", name, f.GoField)
			sb.WriteString(fmt.Sprintf("\t%s = %d\n", constName, f.Index))
		}
		sb.WriteString(")\n")
	}

	// generate EntityMetadataEntry type
	sb.WriteString("\n// MetadataEntry represents a single entity metadata entry.\n")
	sb.WriteString("type MetadataEntry struct {\n")
	sb.WriteString("\tIndex      byte\n")
	sb.WriteString("\tSerializer int32\n")
	sb.WriteString("\tData       []byte // raw wire data\n")
	sb.WriteString("}\n\n")

	// generate struct for each entity with all resolved fields
	for _, name := range entityNames {
		fields := resolvedEntities[name]

		sb.WriteString(fmt.Sprintf("// %sMetadata contains metadata fields for %s entities.\n", name, name))
		sb.WriteString(fmt.Sprintf("type %sMetadata struct {\n", name))

		// field present flags
		sb.WriteString("\t// field presence flags\n")
		for _, f := range fields {
			if f.GoType[0] == '*' || f.Passthrough {
				continue // pointers and passthrough already handle presence
			}
			sb.WriteString(fmt.Sprintf("\tHas%s bool\n", f.GoField))
		}

		sb.WriteString("\n\t// field values\n")
		for _, f := range fields {
			if f.Passthrough {
				sb.WriteString(fmt.Sprintf("\t%s []byte // passthrough\n", f.GoField))
			} else {
				sb.WriteString(fmt.Sprintf("\t%s %s\n", f.GoField, f.GoType))
			}
		}
		sb.WriteString("}\n\n")
	}

	// generate entity metadata registry (maps entity type name to field definitions)
	sb.WriteString("// FieldDef describes an entity metadata field.\n")
	sb.WriteString("type FieldDef struct {\n")
	sb.WriteString("\tIndex       byte\n")
	sb.WriteString("\tSerializer  int32\n")
	sb.WriteString("\tName        string\n")
	sb.WriteString("\tPassthrough bool\n")
	sb.WriteString("}\n\n")

	sb.WriteString("// entityMetadataFields maps entity class names to their field definitions.\n")
	sb.WriteString("var entityMetadataFields = map[string][]FieldDef{\n")
	for _, name := range entityNames {
		fields := resolvedEntities[name]
		sb.WriteString(fmt.Sprintf("\t%q: {\n", name))
		for _, f := range fields {
			sb.WriteString(fmt.Sprintf("\t\t{Index: %d, Serializer: %d, Name: %q, Passthrough: %v},\n",
				f.Index, f.Serializer, f.GoField, f.Passthrough))
		}
		sb.WriteString("\t},\n")
	}
	sb.WriteString("}\n")

	writeFile(outPath, sb.String())
}

// resolveEntityHierarchy resolves parent fields for each entity.
func resolveEntityHierarchy(metadata EntityMetadataJSON) map[string][]FieldJSON {
	result := make(map[string][]FieldJSON)

	var resolve func(name string) []FieldJSON
	resolve = func(name string) []FieldJSON {
		if cached, ok := result[name]; ok {
			return cached
		}

		entity := metadata.Entities[name]
		var fields []FieldJSON

		if entity.Parent != "" {
			fields = append(fields, resolve(entity.Parent)...)
		}
		fields = append(fields, entity.Fields...)

		result[name] = fields
		return fields
	}

	for name := range metadata.Entities {
		resolve(name)
	}

	return result
}
