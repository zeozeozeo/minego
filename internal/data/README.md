# Minecraft Data Packages

Go bindings for Minecraft protocol data including registries, blocks, block states, items, item components, entities, and translations.

## Packages

### `registries`

Contains all 95 Minecraft registries with bidirectional lookups.

```go
import "github.com/zeozeozeo/minego/internal/data/registries"

// get protocol ID for a block
id := registries.Block.Get("minecraft:stone")  // returns 1

// reverse lookup
name := registries.Block.ByID(1)  // returns "minecraft:stone"

// available registries
registries.Block          // 1166 entries
registries.Item           // 1505 entries
registries.EntityType     // 157 entries
registries.MobEffect      // 40 entries
// ... 95 total registries
```

### `blocks`

Contains block protocol IDs, block state calculations, and lookups.

```go
import "github.com/zeozeozeo/minego/internal/data/blocks"

// block ID constants
blocks.Stone          // 1
blocks.OakPlanks      // 15
blocks.DiamondBlock   // 126

// string to ID
id := blocks.BlockID("minecraft:oak_planks")  // 15

// ID to string
name := blocks.BlockName(15)  // "minecraft:oak_planks"

// calculate state ID from block + properties
stateID := blocks.StateID(blocks.OakDoor, map[string]string{
    "facing": "north",
    "half":   "lower",
    "hinge":  "left",
    "open":   "false",
    "powered": "false",
})

// reverse lookup: get block and properties from state ID
blockID, props := blocks.StateProperties(stateID)

// get default state for a block
defaultID := blocks.DefaultStateID(blocks.OakDoor)
```

### `items`

Contains item protocol IDs, lookups, default component data, and slot decoding.

```go
import "github.com/zeozeozeo/minego/internal/data/items"

// item ID constants
items.DiamondSword    // 876
items.Apple           // 918
items.IronPickaxe     // 860

// string to ID
id := items.ItemID("minecraft:diamond_sword")  // 876

// ID to string
name := items.ItemName(876)  // "minecraft:diamond_sword"

// get default components
comps := items.DefaultComponents(items.Apple)
if comps.Food != nil {
    fmt.Printf("nutrition: %d\n", comps.Food.Nutrition)  // 4
}
```

#### ItemStack and the Component Patch Model

In the Minecraft protocol, an item on the wire is not sent as a full component list.
Instead, the Java client uses a `DataComponentPatch` system: each item type has a
*prototype* (default components), and the wire format only carries a **patch** — a list
of component **adds** (overrides) and **removes** (deletions) relative to that prototype.
Components that match the prototype are never sent.

`ItemStack` mirrors this model. Each `Components` struct carries an internal bitset that
tracks which component IDs are explicitly *present*. `ToSlot()` only encodes present
components that differ from the item's defaults:

```go
// create a stack with only specific components (sparse patch)
stack := items.NewStackWithComponents(items.DiamondSword, 1, &items.Components{
    CustomName: &items.ItemNameComponent{Text: "Excalibur"},
    Unbreakable: true,
})
// ToSlot() encodes just these 2 components; defaults are untouched

// create a stack from defaults (full component set)
stack := items.NewStack(items.DiamondSword, 1)
stack.Components.Damage = 100
// ToSlot() encodes only Damage (the one field that differs from defaults)

// opt in to full defaults at any time
stack := items.NewStackWithComponents(items.DiamondSword, 1, &items.Components{
    Unbreakable: true,
})
stack.SetDefaultComponents() // load all defaults, mark all present

// read from wire — full state reconstruction
stack, err := items.ReadSlot(buf)
// FromSlot applies the incoming patch on top of defaults;
// all components are marked as present for faithful re-encoding

// write back to wire
err := stack.WriteSlot(buf)

// convert from/to raw Slot
stack, err := items.FromSlot(rawSlot)
rawSlot, err := stack.ToSlot()
```

For advanced use, the presence bitset can be manipulated directly using the
component ID constants:

```go
stack.Components.SetPresent(items.ComponentDamage)   // mark as present
stack.Components.ClearPresent(items.ComponentDamage)  // mark as absent
stack.Components.HasComponent(items.ComponentDamage)   // query
```

Component type constants (104 types) are generated from the registry:

```go
items.ComponentDamage         // 3
items.ComponentFood           // 23
items.ComponentEnchantments   // 13
items.MaxComponentID          // 103
```

### `entities`

Contains entity type protocol IDs, lookups, and entity metadata parsing.

```go
import "github.com/zeozeozeo/minego/internal/data/entities"

// entity type ID constants
entities.Player       // 128
entities.Zombie       // 156
entities.Creeper      // 32

// string to ID
id := entities.EntityTypeID("minecraft:zombie")  // 156

// ID to string
name := entities.EntityTypeName(156)  // "minecraft:zombie"
```

#### Entity Metadata

Entity metadata is read/written using the wire format: `[Index(UByte)][Type(VarInt)][Value]...[0xFF terminator]`

```go
// read metadata from packet buffer
metadata, err := entities.ReadMetadata(buf)

// access raw data by index
if data := metadata.Get(entities.EntityIndexFlags); data != nil {
    flags := data[0]
    isOnFire := flags&0x01 != 0
    isSneaking := flags&0x02 != 0
}

// set/update metadata
metadata.Set(entities.EntityIndexPose, entities.SerializerPOSE, poseData)

// write metadata back to packet buffer
err := entities.WriteMetadata(buf, metadata)
```

Serializer types (39 types) define how values are encoded on the wire:

```go
entities.SerializerBYTE      // 0 - single byte
entities.SerializerINT       // 1 - VarInt
entities.SerializerFLOAT     // 3 - float32
entities.SerializerSTRING    // 4 - length-prefixed string
entities.SerializerBOOLEAN   // 8 - single byte bool
entities.SerializerROTATIONS // 9 - 3x float32 (x,y,z)
entities.SerializerBLOCK_POS // 10 - packed position
```

Per-entity metadata structs and field indices are generated:

```go
// Entity base class (all entities inherit these)
entities.EntityIndexFlags         // 0 - on_fire|sneaking|sprinting|swimming|invisible|glowing|fall_flying
entities.EntityIndexAirSupply     // 1 - air ticks remaining
entities.EntityIndexCustomName    // 2 - optional name component
entities.EntityIndexPose          // 6 - standing/sneaking/sleeping/etc.

// Creeper-specific (inherits from Mob)
entities.CreeperIndexSwellDir     // 16 - fuse state (-1=idle, 1=fuse)
entities.CreeperIndexIsPowered    // 17 - charged creeper
entities.CreeperIndexIsIgnited    // 18 - ignited

// Player-specific (inherits from LivingEntity)
entities.PlayerIndexAdditionalHearts // 15 - absorption hearts
entities.PlayerIndexSkinParts        // 17 - visible skin parts flags
```

### `lang`

Contains English translations for Minecraft translation keys.

```go
import "github.com/zeozeozeo/minego/internal/data/lang"

// translate a key to English
text := lang.Translate("item.minecraft.iron_sword")  // "Iron Sword"
text := lang.Translate("block.minecraft.stone")      // "Stone"

// returns empty string if key not found
text := lang.Translate("invalid.key")  // ""
```

## Code Generation

The packages are generated from Minecraft server reports. To regenerate:

```bash
cd pkg/data
go generate ./...
go fmt ./...
```

JSON data files must be present in `generate/` (symlinked from `vanilla_server_reports/generated/reports/`).

## Block State Algorithm

Block state IDs are calculated using a right-to-left multiplier approach:

```plain
offset = Σ(property_value_index[i] × ∏(cardinality[j] for j < i))
state_id = base_id + offset
```

Properties are iterated right-to-left, where the rightmost property changes fastest. This matches Minecraft's internal state registration order.

## Caching

`StateID` results are cached globally (default 4096 entries) for repeated lookups with the same inputs:

```go
// configure cache size
blocks.SetCacheSize(8192)  // increase cache
blocks.SetCacheSize(0)     // disable caching
blocks.ClearCache()        // clear cached entries
```

`StateProperties` uses O(log n) binary search and doesn't require caching.

## Performance

Benchmarks on Apple M2 (`go test -bench=. -benchmem ./...`):

| Function | Time | Allocations |
| -------- | ---- | ----------- |
| `StateID` (cached) | ~83 ns/op | 0 |
| `StateID` (uncached) | ~84 ns/op | 0 |
| `StateProperties` | ~138 ns/op | 2 |

## Data Sources

Generated from Minecraft server reports (not committed). To regenerate:

```bash
# generate server reports (see vanilla_server_reports/)
java -DbundlerMainClass=net.minecraft.data.Main -jar server.jar --reports

# regenerate Go code
cd pkg/data && go generate ./... && go fmt ./...
```

Source JSON files:

- `blocks.json`: 1,166 blocks, 29,671 total states, 92 unique properties
- `items.json`: 1,505 items, 104 component types
- `registries.json`: 95 registries

## Testing

```bash
cd pkg/data
go test -v ./...
go test -bench=. -benchmem ./...
```

The test suite verifies all 29,000+ block states against the source JSON.
