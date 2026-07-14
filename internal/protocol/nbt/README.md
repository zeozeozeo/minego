# nbt

Go implementation of Minecraft's Named Binary Tag (NBT) format.

Supports both **file format** (with root tag name) and **network format** (nameless root tag, used since MC 1.20.2).

## Data Types

| Tag ID | Type | Go Type | Description |
| ------ | ---- | ------- | ----------- |
| 0 | End | `End` | Marks end of compound |
| 1 | Byte | `Byte` | Signed 8-bit integer |
| 2 | Short | `Short` | Signed 16-bit integer, big-endian |
| 3 | Int | `Int` | Signed 32-bit integer, big-endian |
| 4 | Long | `Long` | Signed 64-bit integer, big-endian |
| 5 | Float | `Float` | 32-bit IEEE 754 float, big-endian |
| 6 | Double | `Double` | 64-bit IEEE 754 double, big-endian |
| 7 | Byte Array | `ByteArray` | Length-prefixed byte array |
| 8 | String | `String` | Modified UTF-8 with 2-byte length |
| 9 | List | `List` | Homogeneous list of tags |
| 10 | Compound | `Compound` | Map of named tags |
| 11 | Int Array | `IntArray` | Length-prefixed int32 array |
| 12 | Long Array | `LongArray` | Length-prefixed int64 array |

## Usage

### Direct Tag API

```go
import "github.com/zeozeozeo/minego/internal/protocol/nbt"

// build NBT structure
compound := nbt.Compound{
    "name":  nbt.String("Steve"),
    "x":     nbt.Double(100.5),
    "y":     nbt.Double(64.0),
    "items": nbt.List{
        ElementType: nbt.TagCompound,
        Elements: []nbt.Tag{
            nbt.Compound{"id": nbt.String("minecraft:diamond"), "count": nbt.Byte(64)},
        },
    },
}

// encode to network format (nameless root)
data, err := nbt.EncodeNetwork(compound)

// encode to file format (with root name)
data, err := nbt.EncodeFile(compound, "Player")

// decode from network format
tag, err := nbt.DecodeNetwork(data)

// decode from file format
tag, rootName, err := nbt.DecodeFile(data)

// access data
c := tag.(nbt.Compound)
name := c.GetString("name")
x := c.GetDouble("x")
items := c.GetList("items")
```

### Struct Marshaling (like encoding/json)

```go
type Item struct {
    ID    string `nbt:"id"`
    Count int8   `nbt:"count"`
}

type Player struct {
    Name  string  `nbt:"name"`
    X     float64 `nbt:"x"`
    Y     float64 `nbt:"y"`
    Z     float64 `nbt:"z"`
    Items []Item  `nbt:"items"`
    Debug bool    `nbt:"debug,omitempty"` // omit if false
}

player := Player{Name: "Steve", X: 100, Y: 64, Z: -200}

// file format (for .dat files, chunks, etc.)
data, err := nbt.Marshal(player)              // empty root name
data, err := nbt.MarshalFile(player, "Player") // custom root name

var p Player
err := nbt.Unmarshal(data, &p)

// network format (for protocol packets)
data, err := nbt.MarshalNetwork(player)
err := nbt.UnmarshalNetwork(data, &p)
```

### Type Mapping

| Go Type | NBT Type |
| ------- | -------- |
| `bool` | Byte (0/1) |
| `int8` | Byte |
| `int16` | Short |
| `int32`, `int` | Int |
| `int64` | Long |
| `float32` | Float |
| `float64` | Double |
| `string` | String |
| `[]byte` | ByteArray |
| `[]int32` | IntArray |
| `[]int64` | LongArray |
| `[]T` | List |
| `struct` | Compound |
| `map[string]T` | Compound |

### Visitor Pattern (Streaming)

For large NBT files, use the visitor pattern to avoid loading everything into memory:

```go
type MyVisitor struct {
    nbt.BaseVisitor
}

func (v *MyVisitor) VisitString(s string) error {
    fmt.Println("Found string:", s)
    return nil
}

func (v *MyVisitor) VisitCompoundStart() (nbt.Visitor, error) {
    return v, nil // return self to visit entries
}

func (v *MyVisitor) VisitCompoundEntry(name string, tagType byte) (nbt.Visitor, error) {
    fmt.Println("Entry:", name)
    return v, nil
}

// visit existing tag
nbt.AcceptVisitor(tag, &MyVisitor{})

// visit from reader (streaming)
reader := nbt.NewReader(data)
nbt.VisitReader(reader, &MyVisitor{}, true) // true = network format
```

### Safety Limits

```go
// configure limits to prevent memory exhaustion
tag, err := nbt.DecodeNetwork(data,
    nbt.WithMaxDepth(512),      // max nesting depth (default: 512)
    nbt.WithMaxBytes(2*1024*1024), // max bytes to read (default: 2MB)
)
```

## Network vs File Format

**File format** (used for `.dat` files, chunks, etc.):

```plain
[Tag ID: 1 byte]
[Name length: 2 bytes, big-endian]
[Name: UTF-8 bytes]
[Payload]
```

**Network format** (used in packets since 1.20.2):

```plain
[Tag ID: 1 byte]
[Payload]  <- no name!
```

## Benchmarks

Benchmarked on Apple M2. Run with `go test -bench=. -benchmem`.

### Performance

| Operation | Simple (5 fields) | Complex (36 items + nested) |
| --------- | ----------------- | --------------------------- |
| **Encode** | 531 ns/op | 13.4 μs/op |
| **Decode** | 566 ns/op | 15.6 μs/op |
| **Marshal** | 1.0 μs/op | 30.5 μs/op |
| **Unmarshal** | 1.2 μs/op | 33.0 μs/op |
| **Round-trip** | 2.3 μs/op | 62.1 μs/op |

### Allocations

| Operation | Simple | Complex |
| --------- | -------- | --------- |
| Encode | 368 B, 27 allocs | 9.3 KB, 647 allocs |
| Decode | 552 B, 34 allocs | 19.8 KB, 899 allocs |
| Marshal | 752 B, 33 allocs | 26.9 KB, 923 allocs |
| Unmarshal | 600 B, 35 allocs | 22.4 KB, 916 allocs |

### Notes

- **Direct Tag API** (`Encode`/`Decode`) is faster than reflection-based `Marshal`/`Unmarshal`
- Use `Marshal`/`Unmarshal` for convenience, direct API for hot paths
- Complex benchmark simulates realistic Minecraft player data with 36 inventory slots and nested compound tags

## References

- [Minecraft Wiki - NBT](https://minecraft.wiki/w/NBT_format)
- [Minecraft source code](https://github.com/zeozeozeo/minego/internal/protocol/tree/main/data)
