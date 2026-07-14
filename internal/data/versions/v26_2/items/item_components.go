package items

import "maps"

// Components holds all item component data.
// The present bitset tracks which component IDs are explicitly set,
// mirroring Java's DataComponentPatch — only present components are
// encoded on the wire as adds or removes relative to the item's defaults.
type Components struct {
	present [2]uint64 // bitset: which component IDs are explicitly set

	AdditionalTradeCost    int32
	AttributeModifiers     []AttributeModifier
	BlocksAttacks          *BlocksAttacks
	BreakSound             string
	Consumable             *Consumable
	Container              []any
	CustomName             *ItemNameComponent
	Damage                 int32
	DamageResistant        *DamageResistant
	DamageType             string
	DeathProtection        *DeathProtection
	Enchantable            *Enchantable
	Enchantments           map[string]int32
	Equippable             *Equippable
	Fireworks              *Fireworks
	Food                   *Food
	Glider                 bool
	Instrument             string
	ItemModel              string
	ItemName               *ItemNameComponent
	JukeboxPlayable        string
	KineticWeapon          *KineticWeapon
	Lore                   []string
	MapColor               int32
	MaxDamage              int32
	MaxStackSize           int32
	MinimumAttackCharge    float64
	OminousBottleAmplifier int32
	PiercingWeapon         *PiercingWeapon
	PotionContents         *PotionContents
	PotionDurationScale    float64
	ProvidesBannerPatterns string
	ProvidesTrimMaterial   string
	Rarity                 string
	Recipes                []any
	Repairable             *Repairable
	RepairCost             int32
	StoredEnchantments     map[string]int32
	Tool                   *Tool
	TooltipDisplay         *TooltipDisplay
	Unbreakable            bool
	UseCooldown            *UseCooldown
	UseEffects             *UseEffects
	UseRemainder           *UseRemainder
	Weapon                 *Weapon
}

type AttributeModifier struct {
	Type      string
	Amount    float64
	ID        string
	Operation string
	Slot      string
}

type BlocksAttacks struct {
	BlockDelaySeconds float64
	BlockSound        string
	BypassedBy        string
	DisabledSound     string
	ItemDamage        *DamageSpec
}

type DamageSpec struct {
	Base      float64
	Factor    float64
	Threshold float64
}

type Consumable struct {
	ConsumeSeconds   float64
	Animation        string
	OnConsumeEffects []ConsumeEffect
}

type ConsumeEffect struct {
	Type        string
	Effects     []StatusEffect
	Probability float64
}

type StatusEffect struct {
	Duration  int32
	ID        string
	ShowIcon  bool
	Amplifier int32
}

type DamageResistant struct {
	Types string
}

type DeathProtection struct {
	DeathEffects []DeathEffect
}

type DeathEffect struct {
	Type    string
	Effects []StatusEffect
}

type Enchantable struct {
	Value int32
}

type Equippable struct {
	Slot            string
	EquipSound      string
	AssetID         string
	AllowedEntities []string
	Swappable       bool
	CanBeSheared    bool
	ShearingSound   string
}

type Fireworks struct {
	FlightDuration int32
}

type Food struct {
	Nutrition  int32
	Saturation float64
}

type ItemNameComponent struct {
	Text      string
	Translate string
}

type KineticWeapon struct {
	DamageConditions    *KineticConditions
	DamageMultiplier    float64
	DelayTicks          int32
	DismountConditions  *KineticConditions
	ForwardMovement     float64
	HitSound            string
	KnockbackConditions *KineticConditions
	Sound               string
}

type KineticConditions struct {
	MaxDurationTicks int32
	MinRelativeSpeed float64
	MinSpeed         float64
}

type PiercingWeapon struct {
	HitSound string
	Sound    string
}

type PotionContents struct {
	// contents vary by potion type
}

type Repairable struct {
	Items string
}

type Tool struct {
	Rules                      []ToolRule
	DamagePerBlock             int32
	CanDestroyBlocksInCreative bool
}

type ToolRule struct {
	Blocks          string
	Speed           float64
	CorrectForDrops bool
}

type TooltipDisplay struct {
	HideTooltip      bool
	HiddenComponents []int32
}

type UseCooldown struct {
	Seconds float64
}

type UseEffects struct {
	CanSprint          bool
	InteractVibrations bool
	SpeedMultiplier    float64
}

type UseRemainder struct {
	Count int32
	ID    string
}

type Weapon struct {
	DisableBlockingForSeconds float64
	ItemDamagePerAttack       int32
}

type AttackRange struct {
	HitboxMargin     float64
	MaxCreativeReach float64
	MaxReach         float64
	MinCreativeReach float64
	MinReach         float64
	MobFactor        float64
}

// HasComponent returns true if the given component ID is marked as present.
func (c *Components) HasComponent(id int32) bool {
	if id < 0 || id >= 128 {
		return false
	}
	return c.present[id/64]&(1<<uint(id%64)) != 0
}

// SetPresent marks the given component ID as present.
func (c *Components) SetPresent(id int32) {
	if id >= 0 && id < 128 {
		c.present[id/64] |= 1 << uint(id%64)
	}
}

// ClearPresent marks the given component ID as not present.
func (c *Components) ClearPresent(id int32) {
	if id >= 0 && id < 128 {
		c.present[id/64] &^= 1 << uint(id%64)
	}
}

// Clone returns a deep copy of the Components struct.
func (c *Components) Clone() *Components {
	if c == nil {
		return &Components{}
	}

	clone := &Components{
		present:                c.present,
		BreakSound:             c.BreakSound,
		Damage:                 c.Damage,
		DamageType:             c.DamageType,
		Glider:                 c.Glider,
		Instrument:             c.Instrument,
		ItemModel:              c.ItemModel,
		JukeboxPlayable:        c.JukeboxPlayable,
		MapColor:               c.MapColor,
		MaxDamage:              c.MaxDamage,
		MaxStackSize:           c.MaxStackSize,
		MinimumAttackCharge:    c.MinimumAttackCharge,
		OminousBottleAmplifier: c.OminousBottleAmplifier,
		PotionDurationScale:    c.PotionDurationScale,
		ProvidesBannerPatterns: c.ProvidesBannerPatterns,
		ProvidesTrimMaterial:   c.ProvidesTrimMaterial,
		Rarity:                 c.Rarity,
		RepairCost:             c.RepairCost,
	}

	// clone slices
	if c.AttributeModifiers != nil {
		clone.AttributeModifiers = make([]AttributeModifier, len(c.AttributeModifiers))
		copy(clone.AttributeModifiers, c.AttributeModifiers)
	}
	if c.Container != nil {
		clone.Container = make([]any, len(c.Container))
		copy(clone.Container, c.Container)
	}
	if c.Lore != nil {
		clone.Lore = make([]string, len(c.Lore))
		copy(clone.Lore, c.Lore)
	}
	if c.Recipes != nil {
		clone.Recipes = make([]any, len(c.Recipes))
		copy(clone.Recipes, c.Recipes)
	}

	// clone maps
	if c.Enchantments != nil {
		clone.Enchantments = make(map[string]int32, len(c.Enchantments))
		maps.Copy(clone.Enchantments, c.Enchantments)
	}
	if c.StoredEnchantments != nil {
		clone.StoredEnchantments = make(map[string]int32, len(c.StoredEnchantments))
		maps.Copy(clone.StoredEnchantments, c.StoredEnchantments)
	}

	// clone pointer types
	if c.BlocksAttacks != nil {
		v := *c.BlocksAttacks
		clone.BlocksAttacks = &v
	}
	if c.Consumable != nil {
		v := *c.Consumable
		clone.Consumable = &v
	}
	if c.DamageResistant != nil {
		v := *c.DamageResistant
		clone.DamageResistant = &v
	}
	if c.DeathProtection != nil {
		v := *c.DeathProtection
		clone.DeathProtection = &v
	}
	if c.Enchantable != nil {
		v := *c.Enchantable
		clone.Enchantable = &v
	}
	if c.Equippable != nil {
		v := *c.Equippable
		clone.Equippable = &v
	}
	if c.Fireworks != nil {
		v := *c.Fireworks
		clone.Fireworks = &v
	}
	if c.Food != nil {
		v := *c.Food
		clone.Food = &v
	}
	if c.ItemName != nil {
		v := *c.ItemName
		clone.ItemName = &v
	}
	if c.KineticWeapon != nil {
		v := *c.KineticWeapon
		clone.KineticWeapon = &v
	}
	if c.PiercingWeapon != nil {
		v := *c.PiercingWeapon
		clone.PiercingWeapon = &v
	}
	if c.PotionContents != nil {
		v := *c.PotionContents
		clone.PotionContents = &v
	}
	if c.Repairable != nil {
		v := *c.Repairable
		clone.Repairable = &v
	}
	if c.Tool != nil {
		v := *c.Tool
		clone.Tool = &v
	}
	if c.TooltipDisplay != nil {
		v := *c.TooltipDisplay
		clone.TooltipDisplay = &v
	}
	if c.UseCooldown != nil {
		v := *c.UseCooldown
		clone.UseCooldown = &v
	}
	if c.UseEffects != nil {
		v := *c.UseEffects
		clone.UseEffects = &v
	}
	if c.UseRemainder != nil {
		v := *c.UseRemainder
		clone.UseRemainder = &v
	}
	if c.Weapon != nil {
		v := *c.Weapon
		clone.Weapon = &v
	}

	return clone
}
