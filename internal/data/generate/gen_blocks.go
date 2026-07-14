package main

import (
	"fmt"
	"sort"
	"strings"
)

func generateBlocks(_ map[string]RegistryJSON, outPath string) {
	// block IDs now live in registries.Block — no duplicates needed here.
	writeFile(outPath, generatedFileHeader("blocks"))
}

func generateBlockStates(blocks map[string]BlockJSON, registries map[string]RegistryJSON, outPath string) {
	blockRegistry := registries["minecraft:block"]
	// collect all unique properties and their values
	propertyValues := make(map[string][]string)
	for _, block := range blocks {
		for prop, values := range block.Properties {
			if existing, ok := propertyValues[prop]; ok {
				// merge values
				seen := make(map[string]bool)
				for _, v := range existing {
					seen[v] = true
				}
				for _, v := range values {
					if !seen[v] {
						propertyValues[prop] = append(propertyValues[prop], v)
						seen[v] = true
					}
				}
			} else {
				propertyValues[prop] = append([]string{}, values...)
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(generatedFileHeader("blocks"))

	// property value indices lookup
	sb.WriteString("var propertyValueIndices = map[string]map[string]int{\n")
	for _, prop := range sortedKeys(propertyValues) {
		values := propertyValues[prop]
		sb.WriteString(fmt.Sprintf("\t%q: {\n", prop))
		for i, v := range values {
			sb.WriteString(fmt.Sprintf("\t\t%q: %d,\n", v, i))
		}
		sb.WriteString("\t},\n")
	}
	sb.WriteString("}\n\n")

	// generate block state data map
	sb.WriteString("var blockStates = map[int32]*blockStateData{\n")

	// we need block IDs - extract from states
	blockNames := sortedKeys(blocks)
	for _, blockName := range blockNames {
		block := blocks[blockName]
		if len(block.States) == 0 {
			continue
		}

		// find base and default state IDs
		baseID := block.States[0].ID
		defaultID := baseID
		for _, state := range block.States {
			if state.Default {
				defaultID = state.ID
				break
			}
		}

		// get property order from first state (or definition)
		var propOrder []string
		if len(block.Properties) > 0 {
			// need to determine order from states
			// the order in block.Properties map is not guaranteed
			// but states are ordered, so we can infer from state patterns
			propOrder = getPropertyOrder(block)
		}

		blockID := blockRegistry.Entries[blockName].ProtocolID
		sb.WriteString(fmt.Sprintf("\t%d: { // %s\n", blockID, blockName))
		sb.WriteString(fmt.Sprintf("\t\tBaseID:    %d,\n", baseID))
		sb.WriteString(fmt.Sprintf("\t\tDefaultID: %d,\n", defaultID))

		if len(propOrder) > 0 {
			sb.WriteString("\t\tProperties: []blockProperty{\n")
			for _, prop := range propOrder {
				values := block.Properties[prop]
				sb.WriteString(fmt.Sprintf("\t\t\t{%q, []string{", prop))
				for i, v := range values {
					if i > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(fmt.Sprintf("%q", v))
				}
				sb.WriteString(fmt.Sprintf("}, %d},\n", len(values)))
			}
			sb.WriteString("\t\t},\n")
		}

		sb.WriteString("\t},\n")
	}
	sb.WriteString("}\n\n")

	// generate state range index for O(log n) lookup
	type stateRange struct {
		baseID  int32
		endID   int32
		blockID int32
	}
	var ranges []stateRange
	for _, blockName := range blockNames {
		block := blocks[blockName]
		if len(block.States) == 0 {
			continue
		}
		baseID := block.States[0].ID
		endID := block.States[len(block.States)-1].ID + 1
		blockID := blockRegistry.Entries[blockName].ProtocolID
		ranges = append(ranges, stateRange{baseID, endID, blockID})
	}
	// sort by baseID
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].baseID < ranges[j].baseID
	})

	sb.WriteString("// stateRanges is sorted by BaseID for binary search.\n")
	sb.WriteString("var stateRanges = []stateRangeEntry{\n")
	for _, r := range ranges {
		sb.WriteString(fmt.Sprintf("\t{%d, %d, %d},\n", r.baseID, r.endID, r.blockID))
	}
	sb.WriteString("}\n")

	writeFile(outPath, sb.String())
}

// getPropertyOrder determines the order of properties for a block by analyzing its states.
func getPropertyOrder(block BlockJSON) []string {
	if len(block.States) < 2 || len(block.Properties) == 0 {
		return sortedKeys(block.Properties)
	}

	// the property that changes between consecutive states (with smaller index)
	// is the rightmost (fastest changing) property
	// we need to find the order by examining state transitions

	props := sortedKeys(block.Properties)
	if len(props) == 1 {
		return props
	}

	// calculate expected cardinality products to determine order
	// analyze which property changes at each step
	order := make([]string, 0, len(props))
	remaining := make(map[string]bool)
	for _, p := range props {
		remaining[p] = true
	}

	// the rightmost property changes every 1 state
	// the next one changes every (cardinality of rightmost) states
	// etc.

	stepSize := 1
	for len(remaining) > 0 {
		// find the property that changes at this step size
		for prop := range remaining {
			// check if this property changes at the expected interval
			if propertyChangesAtInterval(block, prop, stepSize) {
				order = append(order, prop)
				delete(remaining, prop)
				stepSize *= len(block.Properties[prop])
				break
			}
		}
		// safety check to avoid infinite loop
		if len(order) == 0 || (len(remaining) > 0 && stepSize > len(block.States)) {
			// fallback to sorted order
			return sortedKeys(block.Properties)
		}
	}

	// order is from rightmost to leftmost, need to reverse
	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
		order[i], order[j] = order[j], order[i]
	}

	return order
}

func propertyChangesAtInterval(block BlockJSON, prop string, interval int) bool {
	if interval >= len(block.States) {
		return false
	}

	values := block.Properties[prop]
	if len(values) == 0 {
		return false
	}

	// check that the property cycles through its values at the given interval
	for i := 0; i < len(block.States); i++ {
		expectedValueIdx := (i / interval) % len(values)
		expectedValue := values[expectedValueIdx]
		actualValue := block.States[i].Properties[prop]
		if actualValue != expectedValue {
			return false
		}
	}
	return true
}
