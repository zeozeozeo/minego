// Code generated for Minecraft 26.1 (Protocol 775); DO NOT EDIT.

package entities

// Entity metadata field indices.

// AgeableMob metadata indices
const (
	AgeableMobIndexIsBaby = 16
)

// ArmorStand metadata indices
const (
	ArmorStandIndexArmorStandFlags = 15
	ArmorStandIndexHeadPose        = 16
	ArmorStandIndexBodyPose        = 17
	ArmorStandIndexLeftArmPose     = 18
	ArmorStandIndexRightArmPose    = 19
	ArmorStandIndexLeftLegPose     = 20
	ArmorStandIndexRightLegPose    = 21
)

// Arrow metadata indices
const (
	ArrowIndexArrowFlags  = 8
	ArrowIndexPierceLevel = 9
	ArrowIndexInGround    = 10
)

// Avatar metadata indices
const (
	AvatarIndexMainHand  = 15
	AvatarIndexSkinParts = 16
)

// Bat metadata indices
const (
	BatIndexBatFlags = 16
)

// Blaze metadata indices
const (
	BlazeIndexBlazeFlags = 16
)

// Boat metadata indices
const (
	BoatIndexHurtTime    = 8
	BoatIndexHurtDir     = 9
	BoatIndexDamage      = 10
	BoatIndexPaddleLeft  = 11
	BoatIndexPaddleRight = 12
	BoatIndexBubbleTime  = 13
)

// Cat metadata indices
const (
	CatIndexVariant      = 17
	CatIndexIsLying      = 18
	CatIndexIsRelaxed    = 19
	CatIndexCollarColor  = 20
	CatIndexSoundVariant = 21
)

// Chicken metadata indices
const (
	ChickenIndexVariant      = 17
	ChickenIndexSoundVariant = 18
)

// Cow metadata indices
const (
	CowIndexVariant      = 17
	CowIndexSoundVariant = 18
)

// Creeper metadata indices
const (
	CreeperIndexSwellDir  = 16
	CreeperIndexIsPowered = 17
	CreeperIndexIsIgnited = 18
)

// EnderDragon metadata indices
const (
	EnderDragonIndexPhase = 16
)

// Enderman metadata indices
const (
	EndermanIndexCarriedBlock = 16
	EndermanIndexIsScreaming  = 17
	EndermanIndexIsStaring    = 18
)

// Entity metadata indices
const (
	EntityIndexFlags             = 0
	EntityIndexAirSupply         = 1
	EntityIndexCustomName        = 2
	EntityIndexCustomNameVisible = 3
	EntityIndexSilent            = 4
	EntityIndexNoGravity         = 5
	EntityIndexPose              = 6
	EntityIndexTicksFrozen       = 7
)

// FallingBlock metadata indices
const (
	FallingBlockIndexSpawnPos = 8
)

// Fireball metadata indices
const (
	FireballIndexItem = 8
)

// Ghast metadata indices
const (
	GhastIndexIsCharging = 16
)

// ItemEntity metadata indices
const (
	ItemEntityIndexItem = 8
)

// ItemFrame metadata indices
const (
	ItemFrameIndexItem     = 8
	ItemFrameIndexRotation = 9
)

// LivingEntity metadata indices
const (
	LivingEntityIndexLivingFlags     = 8
	LivingEntityIndexHealth          = 9
	LivingEntityIndexEffectParticles = 10
	LivingEntityIndexEffectAmbient   = 11
	LivingEntityIndexArrowCount      = 12
	LivingEntityIndexStingerCount    = 13
	LivingEntityIndexSleepingPos     = 14
)

// Minecart metadata indices
const (
	MinecartIndexHurtTime      = 8
	MinecartIndexHurtDir       = 9
	MinecartIndexDamage        = 10
	MinecartIndexDisplayBlock  = 11
	MinecartIndexDisplayOffset = 12
	MinecartIndexCustomDisplay = 13
)

// Mob metadata indices
const (
	MobIndexMobFlags = 15
)

// Pig metadata indices
const (
	PigIndexBoostTime    = 17
	PigIndexVariant      = 18
	PigIndexSoundVariant = 19
)

// Player metadata indices
const (
	PlayerIndexAdditionalHearts    = 17
	PlayerIndexScore               = 18
	PlayerIndexLeftShoulderEntity  = 19
	PlayerIndexRightShoulderEntity = 20
)

// Sheep metadata indices
const (
	SheepIndexColorFlags = 17
)

// Skeleton metadata indices
const (
	SkeletonIndexIsConverting = 16
)

// Slime metadata indices
const (
	SlimeIndexSize = 16
)

// Spider metadata indices
const (
	SpiderIndexSpiderFlags = 16
)

// Villager metadata indices
const (
	VillagerIndexVillagerData = 17
)

// Wither metadata indices
const (
	WitherIndexTargetA          = 16
	WitherIndexTargetB          = 17
	WitherIndexTargetC          = 18
	WitherIndexInvulnerableTime = 19
)

// Wolf metadata indices
const (
	WolfIndexIsTamed      = 17
	WolfIndexVariant      = 18
	WolfIndexSoundVariant = 19
	WolfIndexIsAngry      = 20
	WolfIndexCollarColor  = 21
)

// Zombie metadata indices
const (
	ZombieIndexIsBaby          = 16
	ZombieIndexSpecialType     = 17
	ZombieIndexBecomingDrowned = 18
)

// MetadataEntry represents a single entity metadata entry.
type MetadataEntry struct {
	Index      byte
	Serializer int32
	Data       []byte // raw wire data
}

// AgeableMobMetadata contains metadata fields for AgeableMob entities.
type AgeableMobMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasIsBaby            bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	IsBaby            bool
}

// ArmorStandMetadata contains metadata fields for ArmorStand entities.
type ArmorStandMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasArmorStandFlags   bool
	HasHeadPose          bool
	HasBodyPose          bool
	HasLeftArmPose       bool
	HasRightArmPose      bool
	HasLeftLegPose       bool
	HasRightLegPose      bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	ArmorStandFlags   byte
	HeadPose          Rotations
	BodyPose          Rotations
	LeftArmPose       Rotations
	RightArmPose      Rotations
	LeftLegPose       Rotations
	RightLegPose      Rotations
}

// ArrowMetadata contains metadata fields for Arrow entities.
type ArrowMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasArrowFlags        bool
	HasPierceLevel       bool
	HasInGround          bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	ArrowFlags        byte
	PierceLevel       byte
	InGround          bool
}

// AvatarMetadata contains metadata fields for Avatar entities.
type AvatarMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMainHand          bool
	HasSkinParts         bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MainHand          byte
	SkinParts         byte
}

// BatMetadata contains metadata fields for Bat entities.
type BatMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasBatFlags          bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	BatFlags          byte
}

// BlazeMetadata contains metadata fields for Blaze entities.
type BlazeMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasBlazeFlags        bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	BlazeFlags        byte
}

// BoatMetadata contains metadata fields for Boat entities.
type BoatMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasHurtTime          bool
	HasHurtDir           bool
	HasDamage            bool
	HasPaddleLeft        bool
	HasPaddleRight       bool
	HasBubbleTime        bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	HurtTime          int32
	HurtDir           int32
	Damage            float32
	PaddleLeft        bool
	PaddleRight       bool
	BubbleTime        int32
}

// CatMetadata contains metadata fields for Cat entities.
type CatMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasIsBaby            bool
	HasVariant           bool
	HasIsLying           bool
	HasIsRelaxed         bool
	HasCollarColor       bool
	HasSoundVariant      bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	IsBaby            bool
	Variant           int32
	IsLying           bool
	IsRelaxed         bool
	CollarColor       int32
	SoundVariant      int32
}

// ChickenMetadata contains metadata fields for Chicken entities.
type ChickenMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasIsBaby            bool
	HasVariant           bool
	HasSoundVariant      bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	IsBaby            bool
	Variant           int32
	SoundVariant      int32
}

// CowMetadata contains metadata fields for Cow entities.
type CowMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasIsBaby            bool
	HasVariant           bool
	HasSoundVariant      bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	IsBaby            bool
	Variant           int32
	SoundVariant      int32
}

// CreeperMetadata contains metadata fields for Creeper entities.
type CreeperMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasSwellDir          bool
	HasIsPowered         bool
	HasIsIgnited         bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	SwellDir          int32
	IsPowered         bool
	IsIgnited         bool
}

// EnderDragonMetadata contains metadata fields for EnderDragon entities.
type EnderDragonMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasPhase             bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	Phase             int32
}

// EndermanMetadata contains metadata fields for Enderman entities.
type EndermanMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasIsScreaming       bool
	HasIsStaring         bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	CarriedBlock      *int32
	IsScreaming       bool
	IsStaring         bool
}

// EntityMetadata contains metadata fields for Entity entities.
type EntityMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
}

// ExperienceOrbMetadata contains metadata fields for ExperienceOrb entities.
type ExperienceOrbMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
}

// FallingBlockMetadata contains metadata fields for FallingBlock entities.
type FallingBlockMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasSpawnPos          bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	SpawnPos          Position
}

// FireballMetadata contains metadata fields for Fireball entities.
type FireballMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	Item              []byte // passthrough
}

// GhastMetadata contains metadata fields for Ghast entities.
type GhastMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasIsCharging        bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	IsCharging        bool
}

// ItemEntityMetadata contains metadata fields for ItemEntity entities.
type ItemEntityMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	Item              []byte // passthrough
}

// ItemFrameMetadata contains metadata fields for ItemFrame entities.
type ItemFrameMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasRotation          bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	Item              []byte // passthrough
	Rotation          int32
}

// LivingEntityMetadata contains metadata fields for LivingEntity entities.
type LivingEntityMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
}

// MinecartMetadata contains metadata fields for Minecart entities.
type MinecartMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasHurtTime          bool
	HasHurtDir           bool
	HasDamage            bool
	HasDisplayBlock      bool
	HasDisplayOffset     bool
	HasCustomDisplay     bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	HurtTime          int32
	HurtDir           int32
	Damage            float32
	DisplayBlock      int32
	DisplayOffset     int32
	CustomDisplay     bool
}

// MobMetadata contains metadata fields for Mob entities.
type MobMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
}

// PigMetadata contains metadata fields for Pig entities.
type PigMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasIsBaby            bool
	HasBoostTime         bool
	HasVariant           bool
	HasSoundVariant      bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	IsBaby            bool
	BoostTime         int32
	Variant           int32
	SoundVariant      int32
}

// PlayerMetadata contains metadata fields for Player entities.
type PlayerMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMainHand          bool
	HasSkinParts         bool
	HasAdditionalHearts  bool
	HasScore             bool

	// field values
	Flags               byte
	AirSupply           int32
	CustomName          *string
	CustomNameVisible   bool
	Silent              bool
	NoGravity           bool
	Pose                int32
	TicksFrozen         int32
	LivingFlags         byte
	Health              float32
	EffectParticles     []byte // passthrough
	EffectAmbient       bool
	ArrowCount          int32
	StingerCount        int32
	SleepingPos         []byte // passthrough
	MainHand            byte
	SkinParts           byte
	AdditionalHearts    float32
	Score               int32
	LeftShoulderEntity  []byte // passthrough
	RightShoulderEntity []byte // passthrough
}

// SheepMetadata contains metadata fields for Sheep entities.
type SheepMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasIsBaby            bool
	HasColorFlags        bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	IsBaby            bool
	ColorFlags        byte
}

// SkeletonMetadata contains metadata fields for Skeleton entities.
type SkeletonMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasIsConverting      bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	IsConverting      bool
}

// SlimeMetadata contains metadata fields for Slime entities.
type SlimeMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasSize              bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	Size              int32
}

// SpiderMetadata contains metadata fields for Spider entities.
type SpiderMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasSpiderFlags       bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	SpiderFlags       byte
}

// VillagerMetadata contains metadata fields for Villager entities.
type VillagerMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasIsBaby            bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	IsBaby            bool
	VillagerData      []byte // passthrough
}

// WitherMetadata contains metadata fields for Wither entities.
type WitherMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasTargetA           bool
	HasTargetB           bool
	HasTargetC           bool
	HasInvulnerableTime  bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	TargetA           int32
	TargetB           int32
	TargetC           int32
	InvulnerableTime  int32
}

// WolfMetadata contains metadata fields for Wolf entities.
type WolfMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasIsBaby            bool
	HasIsTamed           bool
	HasVariant           bool
	HasSoundVariant      bool
	HasIsAngry           bool
	HasCollarColor       bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	IsBaby            bool
	IsTamed           bool
	Variant           int32
	SoundVariant      int32
	IsAngry           bool
	CollarColor       int32
}

// ZombieMetadata contains metadata fields for Zombie entities.
type ZombieMetadata struct {
	// field presence flags
	HasFlags             bool
	HasAirSupply         bool
	HasCustomNameVisible bool
	HasSilent            bool
	HasNoGravity         bool
	HasPose              bool
	HasTicksFrozen       bool
	HasLivingFlags       bool
	HasHealth            bool
	HasEffectAmbient     bool
	HasArrowCount        bool
	HasStingerCount      bool
	HasMobFlags          bool
	HasIsBaby            bool
	HasSpecialType       bool
	HasBecomingDrowned   bool

	// field values
	Flags             byte
	AirSupply         int32
	CustomName        *string
	CustomNameVisible bool
	Silent            bool
	NoGravity         bool
	Pose              int32
	TicksFrozen       int32
	LivingFlags       byte
	Health            float32
	EffectParticles   []byte // passthrough
	EffectAmbient     bool
	ArrowCount        int32
	StingerCount      int32
	SleepingPos       []byte // passthrough
	MobFlags          byte
	IsBaby            bool
	SpecialType       int32
	BecomingDrowned   bool
}

// FieldDef describes an entity metadata field.
type FieldDef struct {
	Index       byte
	Serializer  int32
	Name        string
	Passthrough bool
}

// entityMetadataFields maps entity class names to their field definitions.
var entityMetadataFields = map[string][]FieldDef{
	"AgeableMob": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 8, Name: "IsBaby", Passthrough: false},
	},
	"ArmorStand": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "ArmorStandFlags", Passthrough: false},
		{Index: 16, Serializer: 9, Name: "HeadPose", Passthrough: false},
		{Index: 17, Serializer: 9, Name: "BodyPose", Passthrough: false},
		{Index: 18, Serializer: 9, Name: "LeftArmPose", Passthrough: false},
		{Index: 19, Serializer: 9, Name: "RightArmPose", Passthrough: false},
		{Index: 20, Serializer: 9, Name: "LeftLegPose", Passthrough: false},
		{Index: 21, Serializer: 9, Name: "RightLegPose", Passthrough: false},
	},
	"Arrow": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "ArrowFlags", Passthrough: false},
		{Index: 9, Serializer: 0, Name: "PierceLevel", Passthrough: false},
		{Index: 10, Serializer: 8, Name: "InGround", Passthrough: false},
	},
	"Avatar": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 42, Name: "MainHand", Passthrough: false},
		{Index: 16, Serializer: 0, Name: "SkinParts", Passthrough: false},
	},
	"Bat": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 0, Name: "BatFlags", Passthrough: false},
	},
	"Blaze": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 0, Name: "BlazeFlags", Passthrough: false},
	},
	"Boat": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 1, Name: "HurtTime", Passthrough: false},
		{Index: 9, Serializer: 1, Name: "HurtDir", Passthrough: false},
		{Index: 10, Serializer: 3, Name: "Damage", Passthrough: false},
		{Index: 11, Serializer: 8, Name: "PaddleLeft", Passthrough: false},
		{Index: 12, Serializer: 8, Name: "PaddleRight", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "BubbleTime", Passthrough: false},
	},
	"Cat": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 8, Name: "IsBaby", Passthrough: false},
		{Index: 17, Serializer: 21, Name: "Variant", Passthrough: false},
		{Index: 18, Serializer: 8, Name: "IsLying", Passthrough: false},
		{Index: 19, Serializer: 8, Name: "IsRelaxed", Passthrough: false},
		{Index: 20, Serializer: 1, Name: "CollarColor", Passthrough: false},
		{Index: 21, Serializer: 22, Name: "SoundVariant", Passthrough: false},
	},
	"Chicken": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 8, Name: "IsBaby", Passthrough: false},
		{Index: 17, Serializer: 30, Name: "Variant", Passthrough: false},
		{Index: 18, Serializer: 31, Name: "SoundVariant", Passthrough: false},
	},
	"Cow": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 8, Name: "IsBaby", Passthrough: false},
		{Index: 17, Serializer: 23, Name: "Variant", Passthrough: false},
		{Index: 18, Serializer: 24, Name: "SoundVariant", Passthrough: false},
	},
	"Creeper": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 1, Name: "SwellDir", Passthrough: false},
		{Index: 17, Serializer: 8, Name: "IsPowered", Passthrough: false},
		{Index: 18, Serializer: 8, Name: "IsIgnited", Passthrough: false},
	},
	"EnderDragon": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 1, Name: "Phase", Passthrough: false},
	},
	"Enderman": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 15, Name: "CarriedBlock", Passthrough: false},
		{Index: 17, Serializer: 8, Name: "IsScreaming", Passthrough: false},
		{Index: 18, Serializer: 8, Name: "IsStaring", Passthrough: false},
	},
	"Entity": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
	},
	"ExperienceOrb": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
	},
	"FallingBlock": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 10, Name: "SpawnPos", Passthrough: false},
	},
	"Fireball": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 7, Name: "Item", Passthrough: true},
	},
	"Ghast": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 8, Name: "IsCharging", Passthrough: false},
	},
	"ItemEntity": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 7, Name: "Item", Passthrough: true},
	},
	"ItemFrame": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 7, Name: "Item", Passthrough: true},
		{Index: 9, Serializer: 1, Name: "Rotation", Passthrough: false},
	},
	"LivingEntity": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
	},
	"Minecart": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 1, Name: "HurtTime", Passthrough: false},
		{Index: 9, Serializer: 1, Name: "HurtDir", Passthrough: false},
		{Index: 10, Serializer: 3, Name: "Damage", Passthrough: false},
		{Index: 11, Serializer: 1, Name: "DisplayBlock", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "DisplayOffset", Passthrough: false},
		{Index: 13, Serializer: 8, Name: "CustomDisplay", Passthrough: false},
	},
	"Mob": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
	},
	"Pig": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 8, Name: "IsBaby", Passthrough: false},
		{Index: 17, Serializer: 1, Name: "BoostTime", Passthrough: false},
		{Index: 18, Serializer: 28, Name: "Variant", Passthrough: false},
		{Index: 19, Serializer: 29, Name: "SoundVariant", Passthrough: false},
	},
	"Player": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 42, Name: "MainHand", Passthrough: false},
		{Index: 16, Serializer: 0, Name: "SkinParts", Passthrough: false},
		{Index: 17, Serializer: 3, Name: "AdditionalHearts", Passthrough: false},
		{Index: 18, Serializer: 1, Name: "Score", Passthrough: false},
		{Index: 19, Serializer: 19, Name: "LeftShoulderEntity", Passthrough: true},
		{Index: 20, Serializer: 19, Name: "RightShoulderEntity", Passthrough: true},
	},
	"Sheep": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 8, Name: "IsBaby", Passthrough: false},
		{Index: 17, Serializer: 0, Name: "ColorFlags", Passthrough: false},
	},
	"Skeleton": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 8, Name: "IsConverting", Passthrough: false},
	},
	"Slime": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 1, Name: "Size", Passthrough: false},
	},
	"Spider": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 0, Name: "SpiderFlags", Passthrough: false},
	},
	"Villager": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 8, Name: "IsBaby", Passthrough: false},
		{Index: 17, Serializer: 18, Name: "VillagerData", Passthrough: true},
	},
	"Wither": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 1, Name: "TargetA", Passthrough: false},
		{Index: 17, Serializer: 1, Name: "TargetB", Passthrough: false},
		{Index: 18, Serializer: 1, Name: "TargetC", Passthrough: false},
		{Index: 19, Serializer: 1, Name: "InvulnerableTime", Passthrough: false},
	},
	"Wolf": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 8, Name: "IsBaby", Passthrough: false},
		{Index: 17, Serializer: 8, Name: "IsTamed", Passthrough: false},
		{Index: 18, Serializer: 25, Name: "Variant", Passthrough: false},
		{Index: 19, Serializer: 26, Name: "SoundVariant", Passthrough: false},
		{Index: 20, Serializer: 8, Name: "IsAngry", Passthrough: false},
		{Index: 21, Serializer: 1, Name: "CollarColor", Passthrough: false},
	},
	"Zombie": {
		{Index: 0, Serializer: 0, Name: "Flags", Passthrough: false},
		{Index: 1, Serializer: 1, Name: "AirSupply", Passthrough: false},
		{Index: 2, Serializer: 6, Name: "CustomName", Passthrough: false},
		{Index: 3, Serializer: 8, Name: "CustomNameVisible", Passthrough: false},
		{Index: 4, Serializer: 8, Name: "Silent", Passthrough: false},
		{Index: 5, Serializer: 8, Name: "NoGravity", Passthrough: false},
		{Index: 6, Serializer: 20, Name: "Pose", Passthrough: false},
		{Index: 7, Serializer: 1, Name: "TicksFrozen", Passthrough: false},
		{Index: 8, Serializer: 0, Name: "LivingFlags", Passthrough: false},
		{Index: 9, Serializer: 3, Name: "Health", Passthrough: false},
		{Index: 10, Serializer: 17, Name: "EffectParticles", Passthrough: true},
		{Index: 11, Serializer: 8, Name: "EffectAmbient", Passthrough: false},
		{Index: 12, Serializer: 1, Name: "ArrowCount", Passthrough: false},
		{Index: 13, Serializer: 1, Name: "StingerCount", Passthrough: false},
		{Index: 14, Serializer: 11, Name: "SleepingPos", Passthrough: true},
		{Index: 15, Serializer: 0, Name: "MobFlags", Passthrough: false},
		{Index: 16, Serializer: 8, Name: "IsBaby", Passthrough: false},
		{Index: 17, Serializer: 1, Name: "SpecialType", Passthrough: false},
		{Index: 18, Serializer: 8, Name: "BecomingDrowned", Passthrough: false},
	},
}
