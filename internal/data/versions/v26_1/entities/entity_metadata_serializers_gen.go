// Code generated for Minecraft 26.1 (Protocol 775); DO NOT EDIT.

package entities

// Entity metadata serializer type IDs.
const (
	SerializerBYTE                             = 0
	SerializerINT                              = 1
	SerializerLONG                             = 2
	SerializerFLOAT                            = 3
	SerializerSTRING                           = 4
	SerializerCOMPONENT                        = 5
	SerializerOPTIONAL_COMPONENT               = 6
	SerializerITEM_STACK                       = 7
	SerializerBOOLEAN                          = 8
	SerializerROTATIONS                        = 9
	SerializerBLOCK_POS                        = 10
	SerializerOPTIONAL_BLOCK_POS               = 11
	SerializerDIRECTION                        = 12
	SerializerOPTIONAL_LIVING_ENTITY_REFERENCE = 13
	SerializerBLOCK_STATE                      = 14
	SerializerOPTIONAL_BLOCK_STATE             = 15
	SerializerPARTICLE                         = 16
	SerializerPARTICLES                        = 17
	SerializerVILLAGER_DATA                    = 18
	SerializerOPTIONAL_UNSIGNED_INT            = 19
	SerializerPOSE                             = 20
	SerializerCAT_VARIANT                      = 21
	SerializerCAT_SOUND_VARIANT                = 22
	SerializerCOW_VARIANT                      = 23
	SerializerCOW_SOUND_VARIANT                = 24
	SerializerWOLF_VARIANT                     = 25
	SerializerWOLF_SOUND_VARIANT               = 26
	SerializerFROG_VARIANT                     = 27
	SerializerPIG_VARIANT                      = 28
	SerializerPIG_SOUND_VARIANT                = 29
	SerializerCHICKEN_VARIANT                  = 30
	SerializerCHICKEN_SOUND_VARIANT            = 31
	SerializerZOMBIE_NAUTILUS_VARIANT          = 32
	SerializerOPTIONAL_GLOBAL_POS              = 33
	SerializerPAINTING_VARIANT                 = 34
	SerializerSNIFFER_STATE                    = 35
	SerializerARMADILLO_STATE                  = 36
	SerializerCOPPER_GOLEM_STATE               = 37
	SerializerWEATHERING_COPPER_STATE          = 38
	SerializerVECTOR3                          = 39
	SerializerQUATERNION                       = 40
	SerializerRESOLVABLE_PROFILE               = 41
	SerializerHUMANOID_ARM                     = 42
)

// serializerNames maps serializer IDs to names.
var serializerNames = map[int32]string{
	0:  "BYTE",
	1:  "INT",
	2:  "LONG",
	3:  "FLOAT",
	4:  "STRING",
	5:  "COMPONENT",
	6:  "OPTIONAL_COMPONENT",
	7:  "ITEM_STACK",
	8:  "BOOLEAN",
	9:  "ROTATIONS",
	10: "BLOCK_POS",
	11: "OPTIONAL_BLOCK_POS",
	12: "DIRECTION",
	13: "OPTIONAL_LIVING_ENTITY_REFERENCE",
	14: "BLOCK_STATE",
	15: "OPTIONAL_BLOCK_STATE",
	16: "PARTICLE",
	17: "PARTICLES",
	18: "VILLAGER_DATA",
	19: "OPTIONAL_UNSIGNED_INT",
	20: "POSE",
	21: "CAT_VARIANT",
	22: "CAT_SOUND_VARIANT",
	23: "COW_VARIANT",
	24: "COW_SOUND_VARIANT",
	25: "WOLF_VARIANT",
	26: "WOLF_SOUND_VARIANT",
	27: "FROG_VARIANT",
	28: "PIG_VARIANT",
	29: "PIG_SOUND_VARIANT",
	30: "CHICKEN_VARIANT",
	31: "CHICKEN_SOUND_VARIANT",
	32: "ZOMBIE_NAUTILUS_VARIANT",
	33: "OPTIONAL_GLOBAL_POS",
	34: "PAINTING_VARIANT",
	35: "SNIFFER_STATE",
	36: "ARMADILLO_STATE",
	37: "COPPER_GOLEM_STATE",
	38: "WEATHERING_COPPER_STATE",
	39: "VECTOR3",
	40: "QUATERNION",
	41: "RESOLVABLE_PROFILE",
	42: "HUMANOID_ARM",
}

// serializerWireTypes maps serializer IDs to wire types.
var serializerWireTypes = map[int32]string{
	0:  "byte",
	1:  "varint",
	2:  "varlong",
	3:  "float32",
	4:  "string",
	5:  "nbt",
	6:  "optional_nbt",
	7:  "slot",
	8:  "bool",
	9:  "rotations",
	10: "position",
	11: "optional_position",
	12: "varint",
	13: "optional_uuid",
	14: "varint",
	15: "varint",
	16: "particle",
	17: "particle_list",
	18: "villager_data",
	19: "optional_varint",
	20: "varint",
	21: "varint",
	22: "varint",
	23: "varint",
	24: "varint",
	25: "varint",
	26: "varint",
	27: "varint",
	28: "varint",
	29: "varint",
	30: "varint",
	31: "varint",
	32: "varint",
	33: "optional_global_pos",
	34: "id_or_inline",
	35: "varint",
	36: "varint",
	37: "varint",
	38: "varint",
	39: "vector3f",
	40: "quaternionf",
	41: "resolvable_profile",
	42: "varint",
}
