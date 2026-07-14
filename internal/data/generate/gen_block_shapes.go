package main

import (
	"fmt"
	"strconv"
	"strings"
)

// mcdumpBlocks is the relevant slice of mcdump/blocks.json.
type mcdumpBlocks struct {
	Blocks map[string]struct {
		States []struct {
			ID    int32 `json:"id"`
			Shape int   `json:"shape"`
		} `json:"states"`
	} `json:"blocks"`
	Shapes [][][]float64 `json:"shapes"` // shape -> list of [minX,minY,minZ,maxX,maxY,maxZ]
}

// generateBlockShapes reads the mod's blocks.json (real game collision shapes per
// block state) and emits a compact shapes table + state->shape index. Shape index
// 0 is always the empty shape so HasCollision/IsFullBlock stay valid.
func generateBlockShapes(mcdumpPath, outPath string) {
	dump := loadJSON[mcdumpBlocks](mcdumpPath)

	// re-intern shapes with the empty shape pinned at index 0.
	compact := [][][]float64{{}}
	indexByKey := map[string]int{shapeKey(nil): 0}
	fullBlockIdx := 0
	intern := func(aabbs [][]float64) int {
		key := shapeKey(aabbs)
		if i, ok := indexByKey[key]; ok {
			return i
		}
		i := len(compact)
		compact = append(compact, aabbs)
		indexByKey[key] = i
		if isFullBlockShape(aabbs) {
			fullBlockIdx = i
		}
		return i
	}

	maxStateID := int32(0)
	for _, b := range dump.Blocks {
		for _, s := range b.States {
			if s.ID > maxStateID {
				maxStateID = s.ID
			}
		}
	}

	shapeByState := make([]uint16, maxStateID+1)
	for _, b := range dump.Blocks {
		for _, s := range b.States {
			shapeByState[s.ID] = uint16(intern(dump.Shapes[s.Shape]))
		}
	}
	checkMinCount("block shape states", len(shapeByState), 1000)

	var sb strings.Builder
	sb.WriteString(generatedFileHeader("blocks"))
	sb.WriteString("import \"github.com/zeozeozeo/minego/internal/data/hitboxes\"\n\n")
	sb.WriteString(fmt.Sprintf("const fullBlockShapeIdx = %d\n\n", fullBlockIdx))

	sb.WriteString(fmt.Sprintf("// shapes contains %d unique collision shapes (shape 0 is empty).\n", len(compact)))
	sb.WriteString("var shapes = [...][]hitboxes.AABB{\n")
	for _, shape := range compact {
		if len(shape) == 0 {
			sb.WriteString("\t{},\n")
			continue
		}
		sb.WriteString("\t{")
		for j, aabb := range shape {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("{%s, %s, %s, %s, %s, %s}",
				formatFloat(aabb[0]), formatFloat(aabb[1]), formatFloat(aabb[2]),
				formatFloat(aabb[3]), formatFloat(aabb[4]), formatFloat(aabb[5])))
		}
		sb.WriteString("},\n")
	}
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// shapeByState maps each of %d block state IDs to a shape index.\n", len(shapeByState)))
	sb.WriteString("var shapeByState = [...]uint16{\n\t")
	for i, idx := range shapeByState {
		if i > 0 && i%20 == 0 {
			sb.WriteString("\n\t")
		}
		sb.WriteString(fmt.Sprintf("%d, ", idx))
	}
	sb.WriteString("\n}\n")

	writeFile(outPath, sb.String())
	fmt.Printf("block shapes: %d states, %d unique shapes\n", len(shapeByState), len(compact))
}

func shapeKey(aabbs [][]float64) string {
	var sb strings.Builder
	for _, a := range aabbs {
		for _, v := range a {
			sb.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
			sb.WriteByte(',')
		}
		sb.WriteByte('|')
	}
	return sb.String()
}

func isFullBlockShape(aabbs [][]float64) bool {
	return len(aabbs) == 1 && len(aabbs[0]) == 6 &&
		aabbs[0][0] == 0 && aabbs[0][1] == 0 && aabbs[0][2] == 0 &&
		aabbs[0][3] == 1 && aabbs[0][4] == 1 && aabbs[0][5] == 1
}

func formatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}
