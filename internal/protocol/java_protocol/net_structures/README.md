# net_structures

Go implementation of Minecraft Java Edition protocol data types.

Based on the [Minecraft Wiki protocol documentation](https://minecraft.wiki/w/Java_Edition_protocol/Data_types) and the decompiled vanilla Java Edition source code (see [go-mclib/mcsrc](https://github.com/go-mclib/mcsrc)).

## Data Types

### Primitives

All multi-byte integers use **big-endian** byte order (same as Java/Netty). These map directly to Go's `encoding/binary.BigEndian`.

| Protocol Type | Go Type | Size | Notes |
| ------------- | ------- | ---- | ----- |
| Boolean | `Boolean` | 1 | `0x00` = false, `0x01` = true |
| Byte | `Int8` | 1 | Signed 8-bit |
| Unsigned Byte | `Uint8` | 1 | Unsigned 8-bit |
| Short | `Int16` | 2 | Signed 16-bit |
| Unsigned Short | `Uint16` | 2 | Unsigned 16-bit |
| Int | `Int32` | 4 | Signed 32-bit |
| Long | `Int64` | 8 | Signed 64-bit |
| Float | `Float32` | 4 | IEEE 754 single-precision |
| Double | `Float64` | 8 | IEEE 754 double-precision |

### Variable-Length Integers

| Protocol Type | Go Type | Max Size | Notes |
| ------------- | ------- | -------- | ----- |
| VarInt | `VarInt` | 5 bytes | 7-bit encoding with continuation bit |
| VarLong | `VarLong` | 10 bytes | Same encoding for 64-bit |

Encoding: each byte uses bits 0-6 for data, bit 7 as continuation flag (1 = more bytes follow).

```plain
0          -> [0x00]
127        -> [0x7f]
128        -> [0x80, 0x01]
-1         -> [0xff, 0xff, 0xff, 0xff, 0x0f]
```

### Strings

| Protocol Type | Go Type | Notes |
| ------------- | ------- | ----- |
| String | `String` | VarInt byte-length prefix + UTF-8 bytes |
| Identifier | `Identifier` | Same as String, format: `namespace:path` |

### Complex Types

| Protocol Type | Go Type | Notes |
| ------------- | ------- | ----- |
| Position | `Position` | Block coordinates packed into int64: X(26 bits) + Z(26 bits) + Y(12 bits) |
| UUID | `UUID` | 128-bit, stored as `[16]byte` |
| Angle | `Angle` | Rotation in 1/256 of a full turn (1 byte) |
| Byte Array | `ByteArray` | VarInt length prefix + raw bytes |
| LpVec3 | `LpVec3` | Low-precision 3D vector for entity velocity |

### Composite Types

These types handle common patterns like length-prefixed arrays, boolean-prefixed optionals, and bit sets.

| Protocol Type | Go Type | Wire Format |
| ------------- | ------- | ----------- |
| Prefixed Array | `PrefixedArray[T]` | VarInt length + elements |
| Prefixed Optional | `PrefixedOptional[T]` | Boolean + value (if true) |
| BitSet | `BitSet` | VarInt length (in longs) + int64 array |
| Fixed BitSet | `FixedBitSet` | ceil(n/8) bytes (no length prefix) |
| ID Set | `IDSet` | VarInt type + tag name or IDs |
| X or Y | `XOrY[X, Y]` | Boolean selector + X or Y value |
| ID or X | `IDOrX[T]` | VarInt ID (0 = inline value follows) |

## Usage

```go
import ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"

// writing
buf := ns.NewWriter()
buf.WriteVarInt(25565)
buf.WriteString("localhost")
buf.WriteUint16(25565)
buf.WritePosition(ns.Position{X: 100, Y: 64, Z: -200})
data := buf.Bytes()

// reading
buf := ns.NewReader(data)
version, _ := buf.ReadVarInt()
address, _ := buf.ReadString(255)
port, _ := buf.ReadUint16()
pos, _ := buf.ReadPosition()

// low-level streaming (directly with net.Conn)
buf := ns.NewWriterTo(conn)
buf.WriteVarInt(0x00)

buf := ns.NewReaderFrom(conn)
packetID, _ := buf.ReadVarInt()
```

### Composite Types Usage

```go
// PrefixedArray - VarInt length-prefixed array
type MyPacket struct {
    Names ns.PrefixedArray[ns.String]
}

func (p *MyPacket) Read(buf *ns.PacketBuffer) error {
    return p.Names.DecodeWith(buf, func(b *ns.PacketBuffer) (ns.String, error) {
        return b.ReadString(32767)
    })
}

func (p *MyPacket) Write(buf *ns.PacketBuffer) error {
    return p.Names.EncodeWith(buf, func(b *ns.PacketBuffer, v ns.String) error {
        return b.WriteString(v)
    })
}

// PrefixedOptional - Boolean-prefixed optional
type MyPacket2 struct {
    Title ns.PrefixedOptional[ns.String]
}

// create optionals
title := ns.Some("Hello")       // present
noTitle := ns.None[ns.String]() // absent

// BitSet - dynamic bit set
bits := ns.NewBitSet(128)
bits.Set(5)
bits.Get(5) // true
bits.Encode(buf)

// FixedBitSet - fixed-size bit set
fixed := ns.NewFixedBitSet(20) // 20 bits = 3 bytes
fixed.Set(0)
fixed.Encode(buf)

// IDSet - registry ID set
tagSet := ns.NewTagIDSet("minecraft:climbable")
inlineSet := ns.NewInlineIDSet([]ns.VarInt{1, 2, 3})
```

### XOrY - Boolean-Selected Variant

`XOrY[X, Y]` represents a value that can be one of two types, selected by a boolean.

```go
// wire format: Boolean (isX) + X or Y value
type MyPacket struct {
    // either an inline value or a registry reference
    Data ns.XOrY[InlineData, ns.VarInt]
}

// create variants
xVal := ns.NewX[InlineData, ns.VarInt](myData) // isX = true
yVal := ns.NewY[InlineData, ns.VarInt](42)     // isX = false

// decode
var v ns.XOrY[InlineData, ns.VarInt]
v.DecodeWith(buf,
    func(b *ns.PacketBuffer) (InlineData, error) { return decodeInline(b) },
    func(b *ns.PacketBuffer) (ns.VarInt, error) { return b.ReadVarInt() },
)

// check which variant
x, y, isX := v.Get()
if isX {
    // use x
} else {
    // use y
}
```

### IDOrX - Registry ID or Inline Value

`IDOrX[T]` represents either a registry ID reference or an inline value.

```go
// wire format: VarInt (0 = inline follows, >0 = ID + 1)
type MyPacket struct {
    Effect ns.IDOrX[EffectData]
}

// create variants
byID := ns.NewIDRef[EffectData](5)           // references registry ID 5
inline := ns.NewInlineValue(EffectData{...}) // inline value

// decode
var v ns.IDOrX[EffectData]
v.DecodeWith(buf, func(b *ns.PacketBuffer) (EffectData, error) {
    return decodeEffect(b)
})

// check which variant
id, value, isInline := v.Get()
if isInline {
    // use value
} else {
    // look up id in registry
}
```

### LpVec3 - Low-Precision Vector

`LpVec3` encodes 3 float64 values in typically 6 bytes using 15-bit scaled values. Used for entity velocity.

```go
// write
vel := ns.LpVec3{X: 0.5, Y: -0.1, Z: 0.0}
buf.WriteLpVec3(vel)

// read
vel, _ := buf.ReadLpVec3()
fmt.Printf("velocity: %.4f, %.4f, %.4f\n", vel.X, vel.Y, vel.Z)
```

Wire format:

- If all components are essentially zero (< 3.05e-5), sends single `0x00` byte
- Otherwise: 6 bytes encoding scale (3 bits) + X/Y/Z (15 bits each)

### GameProfile

`GameProfile` represents a player's profile with UUID, username, and properties.

```go
// read
profile, _ := buf.ReadGameProfile()
fmt.Printf("player: %s (%s)\n", profile.Username, profile.UUID)

// write
profile := ns.GameProfile{
    UUID:     playerUUID,
    Username: "Steve",
}
buf.WriteGameProfile(profile)

// with properties (e.g., textures)
profile.Properties = ns.PrefixedArray[ns.ProfileProperty]{
    {
        Name:  "textures",
        Value: base64TextureData,
        Signature: ns.Some(signatureData),
    },
}
```

`ResolvableProfile` is a variant that can be partial (for lookups) or complete:

```go
// partial profile (for server-side resolution)
partial := ns.NewPartialProfile()
partial.PartialUsername = ns.Some("Steve")

// complete profile
complete := ns.NewCompleteProfile(gameProfile)
complete.BodyModel = ns.Some(ns.Identifier("minecraft:slim"))
```

### NBT (Named Binary Tag)

NBT is used for complex structured data in packets. The `nbt` package supports both file format (with root name) and **network format** (nameless root) used in packets. For communication with the server, use the network format.

#### Direct decoding

```go
import "github.com/zeozeozeo/minego/internal/protocol/nbt"

type EntityData struct {
    Name     string `nbt:"Name"`
    Position int64  `nbt:"Position"`
    OnGround bool   `nbt:"OnGround"`
}

type S2CSomePacket struct {
    EntityID   ns.VarInt
    Data       EntityData
    ExtraField ns.VarInt
}

func (p *S2CSomePacket) Read(buf *ns.PacketBuffer) error {
    var err error
    if p.EntityID, err = buf.ReadVarInt(); err != nil {
        return err
    }

    // use nbt.NewReaderFrom to read the NBT data, which stops at TAG_End
    nbtReader := nbt.NewReaderFrom(buf.Reader())
    tag, _, err := nbtReader.ReadTag(true) // true = network format
    if err != nil {
        return err
    }
    if err := nbt.UnmarshalTag(tag, &p.Data); err != nil {
        return err
    }
    p.ExtraField, err = buf.ReadVarInt()
    return err
}

func (p *S2CSomePacket) Write(buf *ns.PacketBuffer) error {
    if err := buf.WriteVarInt(p.EntityID); err != nil {
        return err
    }
    nbtData, err := nbt.MarshalNetwork(p.Data)
    if err != nil {
        return err
    }
    if _, err := buf.Write(nbtData); err != nil {
        return err
    }
    return buf.WriteVarInt(p.ExtraField)
}
```

#### Storing as `nbt.Tag` (lazy processing)

For packets where you want to defer NBT processing (maybe the NBT data is too large, or dynamic):

```go
type S2CSomePacket struct {
    EntityID ns.VarInt
    Data     nbt.Tag   // store as generic Tag
}

func (p *S2CSomePacket) Read(buf *ns.PacketBuffer) error {
    var err error
    if p.EntityID, err = buf.ReadVarInt(); err != nil {
        return err
    }
    nbtReader := nbt.NewReaderFrom(buf.Reader())
    p.Data, _, err = nbtReader.ReadTag(true)
    return err
}

// later, convert to struct to grab values when needed
// extra fields in the NBT data that are not present in
// the EntityData struct will be skipped
var entityData EntityData
err := nbt.UnmarshalTag(packet.Data, &entityData)
```

#### Empty/Optional NBT

Some packets use a single `TAG_End` byte (`0x00`) to indicate empty or absent NBT data. Check for `nbt.End{}` type after reading:

```go
tag, _, err := nbtReader.ReadTag(true)
if _, isEmpty := tag.(nbt.End); isEmpty {
    // no NBT data present
}
```

### Text Component

Text components are used for chat messages, item names, titles, and other formatted text. Simple text-only components use NBT String tags, complex components use NBT Compound tags. JSON serialization is also supported (all fields carry `json` struct tags).

```go
// simple text
tc := ns.NewTextComponent("Hello, World!")

// with style
bold := true
tc := ns.TextComponent{
    Text:  "Styled text",
    Color: "red",
    Bold:  &bold,
}

// translatable
tc := ns.NewTranslateComponent("chat.type.text",
    ns.NewTextComponent("Player"),
    ns.NewTextComponent("Hello"),
)

// with children
tc := ns.TextComponent{
    Text: "Hello, ",
    Extra: []ns.TextComponent{
        {Text: "World", Color: "gold"},
        {Text: "!"},
    },
}

// read/write (NBT format)
buf.WriteTextComponent(tc)
tc, _ := buf.ReadTextComponent()

// JSON (handles both plain strings and objects)
json.Unmarshal([]byte(`"Hello"`), &tc)        // plain string
json.Unmarshal([]byte(`{"text":"Hello"}`), &tc) // object
data, _ := json.Marshal(tc)
```

#### Rendering

Text components can be converted to various text formats:

```go
tc := ns.TextComponent{Text: "Hello", Color: "red", Bold: &bold}

tc.String()     // "Hello"                - plain text, no formatting
tc.ANSI()       // "\033[91m\033[1mHello\033[0m" - ANSI terminal colors
tc.ColorCodes() // "§c§lHello"            - Bukkit-style § color codes
tc.MiniMessage() // "<red><bold>Hello</bold></red>" - Adventure MiniMessage
```

All renderers recurse into `Extra` and `With` children. `ANSI()` supports named colors, hex colors (`#rrggbb` via 24-bit ANSI), bold, italic, underline, strikethrough, and obfuscated. `MiniMessage()` emits `<lang:key:args>` for translatable components and `<key:name>` for keybinds.

### Slot (Item Stack)

Slots represent item stacks with data components. This package stores components as raw bytes - callers should use a higher-level package to parse specific component types.

Wire format:

- `VarInt count` - item count (0 = empty slot)
- `VarInt item_id` - registry ID (only if count > 0)
- `VarInt add_count` - components to add
- `VarInt remove_count` - components to remove
- Components: each is `VarInt id` + component-specific data

```go
// empty slot
slot := ns.EmptySlot()

// basic item
slot := ns.NewSlot(1, 64) // item ID 1, 64 count

// add raw component data
slot.AddComponent(3, []byte{0x32}) // component ID 3 with data

// remove a component type
slot.RemoveComponent(4)

// reading requires a decoder that knows component sizes
slot, err := buf.ReadSlot(func(buf *ns.PacketBuffer, id ns.VarInt) ([]byte, error) {
    // decode component based on ID, return raw bytes
    switch id {
    case 3: // damage component
        v, err := buf.ReadVarInt()
        if err != nil {
            return nil, err
        }
        return encodeVarInt(v), nil
    default:
        return nil, fmt.Errorf("unknown component: %d", id)
    }
})

// writing
buf.WriteSlot(slot)

// get raw component by ID
if comp := slot.GetComponent(3); comp != nil {
    // comp.ID, comp.Data
}
```

### Chunk Data

`ChunkData` represents chunk section data and block entities. Heightmaps are stored as raw NBT, chunk sections as raw bytes. Parsing block data requires knowledge of the current registry.

```go
// read chunk data
chunkData, err := buf.ReadChunkData()

// heightmaps as NBT compound
fmt.Printf("heightmaps: %v\n", chunkData.Heightmaps)

// raw chunk section data (needs registry to parse)
fmt.Printf("data size: %d bytes\n", len(chunkData.Data))

// block entities
for _, be := range chunkData.BlockEntities {
    x, z := be.X(), be.Z() // relative coords 0-15
    y := be.Y             // absolute Y
    typeID := be.Type     // block entity type registry ID
    data := be.Data       // NBT data
}

// write
buf.WriteChunkData(chunkData)
```

### Light Data

`LightData` represents lighting information for a chunk, including sky and block light.

```go
// read light data
lightData, err := buf.ReadLightData()

// check which sections have light data
if lightData.SkyLightMask.Get(5) {
    // section 5 has sky light data
}

// light arrays are 2048 bytes each (4096 nibbles for 16x16x16 blocks)
for i, arr := range lightData.SkyLightArrays {
    // each byte contains 2 light values (4 bits each)
}

// write
buf.WriteLightData(lightData)
```

Wire format:

- `SkyLightMask` - BitSet indicating sections with sky light
- `BlockLightMask` - BitSet indicating sections with block light
- `EmptySkyLightMask` - BitSet indicating sections with all-zero sky light
- `EmptyBlockLightMask` - BitSet indicating sections with all-zero block light
- `SkyLightArrays` - VarInt count + 2048-byte arrays
- `BlockLightArrays` - VarInt count + 2048-byte arrays

## References

- [Minecraft Wiki - Data Types](https://minecraft.wiki/w/Java_Edition_protocol/Data_types)
- [Minecraft Wiki - Protocol](https://minecraft.wiki/w/Java_Edition_protocol)
- [Minecraft Wiki - Chunk Format](https://minecraft.wiki/w/Chunk_format)
