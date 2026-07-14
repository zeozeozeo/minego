package chunks

import ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"

const lightSections = SectionCount + 2 // 1 below + 24 chunk + 1 above

// ComputeSkylight computes simple vertical sky light propagation.
// For each XZ column, sky light starts at 15 from the top and is blocked by any non-air block.
// This does not perform horizontal light spreading (sufficient for flat worlds).
func (c *ChunkColumn) ComputeSkylight() {
	skyLight := make([][]byte, lightSections)

	// above-world section: full sky light
	skyLight[lightSections-1] = fullSkyLightArray()

	for x := range 16 {
		for z := range 16 {
			light := byte(15)
			for secIdx := SectionCount - 1; secIdx >= 0; secIdx-- {
				lightIdx := secIdx + 1
				if skyLight[lightIdx] == nil {
					skyLight[lightIdx] = make([]byte, 2048)
				}

				for ly := 15; ly >= 0; ly-- {
					// opaque blocks block sky light
					sec := c.Sections[secIdx]
					if sec != nil && sec.GetBlockState(x, ly, z) != 0 {
						light = 0
					}
					setLightNibble(skyLight[lightIdx], x, ly, z, light)
				}
			}
		}
	}

	// build LightData
	skyMask := ns.NewBitSet(lightSections)
	emptySkyMask := ns.NewBitSet(lightSections)
	var skyArrays [][]byte

	for i := range lightSections {
		if skyLight[i] != nil && !isAllZero(skyLight[i]) {
			skyMask.Set(i)
			skyArrays = append(skyArrays, skyLight[i])
		} else {
			emptySkyMask.Set(i)
		}
	}

	c.Light = &ns.LightData{
		SkyLightMask:      *skyMask,
		EmptySkyLightMask: *emptySkyMask,
		SkyLightArrays:    skyArrays,
	}
}

func fullSkyLightArray() []byte {
	arr := make([]byte, 2048)
	for i := range arr {
		arr[i] = 0xFF
	}
	return arr
}

func setLightNibble(arr []byte, x, y, z int, value byte) {
	idx := (y*16+z)*16 + x
	byteIdx := idx / 2
	if idx%2 == 0 {
		arr[byteIdx] = (arr[byteIdx] & 0xF0) | (value & 0x0F)
	} else {
		arr[byteIdx] = (arr[byteIdx] & 0x0F) | ((value & 0x0F) << 4)
	}
}

func isAllZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}
