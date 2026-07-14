package chunks

// DecodeSectionPosition decodes the packed section position from S2CSectionBlocksUpdate.
//
// Bit layout: X (22 bits, signed) at bits 42-63, Z (22 bits) at 20-41, Y (20 bits) at 0-19.
func DecodeSectionPosition(packed int64) (sectionX, sectionY, sectionZ int32) {
	sectionX = int32(packed >> 42)
	sectionZ = int32(packed << 22 >> 42)
	sectionY = int32(packed << 44 >> 44)
	return
}

// DecodeBlockEntry decodes a VarLong block entry from S2CSectionBlocksUpdate.
//
// Bit layout: block state ID at bits 12+, local position at bits 0-11
// (X at bits 8-11, Z at bits 4-7, Y at bits 0-3).
func DecodeBlockEntry(entry int64) (stateID int32, localX, localY, localZ int) {
	stateID = int32(entry >> 12)
	pos := int(entry & 0xFFF)
	localX = (pos >> 8) & 0xF
	localZ = (pos >> 4) & 0xF
	localY = pos & 0xF
	return
}
