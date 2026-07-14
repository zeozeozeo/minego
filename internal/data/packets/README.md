# Packet Definitions

This directory contains Minecraft protocol packet definitions. Each packet implements the `java_protocol.Packet` interface from the [`go-mclib/protocol`](https://github.com/zeozeozeo/minego/internal/protocol) package.

## Imports

```go
package packets

import (
    jp "github.com/zeozeozeo/minego/internal/protocol/java_protocol"
    ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)
```

## Defining a Packet

Each packet is a struct that implements the `Packet` interface:

```go
type Packet interface {
    ID() ns.VarInt
    State() State
    Bound() Bound
    Read(buf *ns.PacketBuffer) error
    Write(buf *ns.PacketBuffer) error
}
```

There is also a lower-level `WirePacket` struct that represents the raw packet as it is sent over the network, but for defining the actual packets, we use only the `Packet` interface.

### Example: Simple Packet

Each packet has some fields, and then the `Read` and `Write` methods to serialize and deserialize the packet, so it can be sent over the network in form of raw bytes.

A packet in its raw form looks like this:

```plain
┌──────────────────┬──────────────────┬──────────────────────────────────────┐
│  Packet Length   │    Packet ID     │                Data                  │
│    (VarInt)      │    (VarInt)      │            (ByteArray)               │
└──────────────────┴──────────────────┴──────────────────────────────────────┘
```

`Read` and `Write` methods are responsible for reading and writing the packet data to and from the `Data` field. Do not include logic for reading and writing the packet length and packet ID, as that part is handled automatically by `go-mclib/protocol`.

To define a simple packet:

```go
// adding docstrings (e.g. from Minecraft Wiki & friends) is recommended, such as:
// C2SHandshake is sent by the client to initiate a connection.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Handshake
type C2SHandshake struct {
    ProtocolVersion ns.VarInt
    ServerAddress   ns.String
    ServerPort      ns.Uint16
    NextState       ns.VarInt
}

func (p *C2SHandshake) ID() ns.VarInt { return 0x00 }
func (p *C2SHandshake) State() jp.State { return jp.StateHandshake }
func (p *C2SHandshake) Bound() jp.Bound { return jp.C2S }

func (p *C2SHandshake) Read(buf *ns.PacketBuffer) error {
    var err error
    if p.ProtocolVersion, err = buf.ReadVarInt(); err != nil {
        return err
    }
    if p.ServerAddress, err = buf.ReadString(255); err != nil {
        return err
    }
    if p.ServerPort, err = buf.ReadUint16(); err != nil {
        return err
    }
    p.NextState, err = buf.ReadVarInt()
    return err
}

func (p *C2SHandshake) Write(buf *ns.PacketBuffer) error {
    if err := buf.WriteVarInt(p.ProtocolVersion); err != nil {
        return err
    }
    if err := buf.WriteString(p.ServerAddress); err != nil {
        return err
    }
    if err := buf.WriteUint16(p.ServerPort); err != nil {
        return err
    }
    return buf.WriteVarInt(p.NextState)
}
```

### Example: Packet with No Fields

There are cases when a packet has no fields, so the `Read` and `Write` methods return `nil` for no data:

```go
// C2SClientTickEnd signals the end of a client tick.
type C2SClientTickEnd struct{}

func (p *C2SClientTickEnd) ID() ns.VarInt { return 0x0C }
func (p *C2SClientTickEnd) State() jp.State { return jp.StatePlay }
func (p *C2SClientTickEnd) Bound() jp.Bound { return jp.C2S }
func (p *C2SClientTickEnd) Read(buf *ns.PacketBuffer) error { return nil }
func (p *C2SClientTickEnd) Write(buf *ns.PacketBuffer) error { return nil }
```

## Example: More Complex Packet

Some packets contain arrays of complex nested structures. Define separate structs for the nested types, then handle the array manually:

```go
// KnownPack represents a single known resource pack entry.
type KnownPack struct {
    Namespace ns.String
    ID        ns.String
    Version   ns.String
}

// C2SSelectKnownPacks is sent by the client to select known packs.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Select_Known_Packs
type C2SSelectKnownPacks struct {
    KnownPacks []KnownPack
}

func (p *C2SSelectKnownPacks) ID() ns.VarInt    { return 0x07 }
func (p *C2SSelectKnownPacks) State() jp.State  { return jp.StateConfiguration }
func (p *C2SSelectKnownPacks) Bound() jp.Bound  { return jp.C2S }

func (p *C2SSelectKnownPacks) Read(buf *ns.PacketBuffer) error {
    // read the array length prefix
    count, err := buf.ReadVarInt()
    if err != nil {
        return err
    }

    // allocate and read each element
    p.KnownPacks = make([]KnownPack, count)
    for i := range p.KnownPacks {
        if p.KnownPacks[i].Namespace, err = buf.ReadString(32767); err != nil {
            return err
        }
        if p.KnownPacks[i].ID, err = buf.ReadString(32767); err != nil {
            return err
        }
        if p.KnownPacks[i].Version, err = buf.ReadString(32767); err != nil {
            return err
        }
    }
    return nil
}

func (p *C2SSelectKnownPacks) Write(buf *ns.PacketBuffer) error {
    // write the array length prefix
    if err := buf.WriteVarInt(ns.VarInt(len(p.KnownPacks))); err != nil {
        return err
    }

    // write each element
    for _, pack := range p.KnownPacks {
        if err := buf.WriteString(pack.Namespace); err != nil {
            return err
        }
        if err := buf.WriteString(pack.ID); err != nil {
            return err
        }
        if err := buf.WriteString(pack.Version); err != nil {
            return err
        }
    }
    return nil
}
```

### Using Composite Types (Recommended)

For arrays and optionals, use the composite types from `net_structures` for cleaner code:

```go
// KnownPack represents a single known resource pack entry.
type KnownPack struct {
    Namespace ns.String
    ID        ns.String
    Version   ns.String
}

// C2SSelectKnownPacks using ns.PrefixedArray
type C2SSelectKnownPacks struct {
    KnownPacks ns.PrefixedArray[KnownPack]
}

func (p *C2SSelectKnownPacks) Read(buf *ns.PacketBuffer) error {
    return p.KnownPacks.DecodeWith(buf, func(buf *ns.PacketBuffer) (KnownPack, error) {
        var pack KnownPack
        var err error
        if pack.Namespace, err = buf.ReadString(32767); err != nil {
            return pack, err
        }
        if pack.ID, err = buf.ReadString(32767); err != nil {
            return pack, err
        }
        pack.Version, err = buf.ReadString(32767)
        return pack, err
    })
}

func (p *C2SSelectKnownPacks) Write(buf *ns.PacketBuffer) error {
    return p.KnownPacks.EncodeWith(buf, func(buf *ns.PacketBuffer, pack KnownPack) error {
        if err := buf.WriteString(pack.Namespace); err != nil {
            return err
        }
        if err := buf.WriteString(pack.ID); err != nil {
            return err
        }
        return buf.WriteString(pack.Version)
    })
}
```

For primitive types, define inline decoder/encoder functions:

```go
type C2SExamplePacket struct {
    IDs   ns.PrefixedArray[ns.VarInt]
    Title ns.PrefixedOptional[ns.String]
}

func (p *C2SExamplePacket) Read(buf *ns.PacketBuffer) error {
    if err := p.IDs.DecodeWith(buf, func(b *ns.PacketBuffer) (ns.VarInt, error) {
        return b.ReadVarInt()
    }); err != nil {
        return err
    }
    return p.Title.DecodeWith(buf, func(b *ns.PacketBuffer) (ns.String, error) {
        return b.ReadString(32767)
    })
}

func (p *C2SExamplePacket) Write(buf *ns.PacketBuffer) error {
    if err := p.IDs.EncodeWith(buf, func(b *ns.PacketBuffer, v ns.VarInt) error {
        return b.WriteVarInt(v)
    }); err != nil {
        return err
    }
    return p.Title.EncodeWith(buf, func(b *ns.PacketBuffer, v ns.String) error {
        return b.WriteString(v)
    })
}
```

## Available Types

### Primitive Types

| Type | Read Method | Write Method |
| ---- | ----------- | ------------ |
| `ns.Boolean` | `buf.ReadBool()` | `buf.WriteBool(v)` |
| `ns.Int8` | `buf.ReadInt8()` | `buf.WriteInt8(v)` |
| `ns.Uint8` | `buf.ReadUint8()` | `buf.WriteUint8(v)` |
| `ns.Int16` | `buf.ReadInt16()` | `buf.WriteInt16(v)` |
| `ns.Uint16` | `buf.ReadUint16()` | `buf.WriteUint16(v)` |
| `ns.Int32` | `buf.ReadInt32()` | `buf.WriteInt32(v)` |
| `ns.Int64` | `buf.ReadInt64()` | `buf.WriteInt64(v)` |
| `ns.Float32` | `buf.ReadFloat32()` | `buf.WriteFloat32(v)` |
| `ns.Float64` | `buf.ReadFloat64()` | `buf.WriteFloat64(v)` |

### Variable-Length Types

| Type | Read Method | Write Method |
| ---- | ----------- | ------------ |
| `ns.VarInt` | `buf.ReadVarInt()` | `buf.WriteVarInt(v)` |
| `ns.VarLong` | `buf.ReadVarLong()` | `buf.WriteVarLong(v)` |

### Complex Types

| Type | Read Method | Write Method |
| ---- | ----------- | ------------ |
| `ns.String` | `buf.ReadString(maxLen)` | `buf.WriteString(v)` |
| `ns.Identifier` | `buf.ReadIdentifier()` | `buf.WriteIdentifier(v)` |
| `ns.UUID` | `buf.ReadUUID()` | `buf.WriteUUID(v)` |
| `ns.Position` | `buf.ReadPosition()` | `buf.WritePosition(v)` |
| `ns.Angle` | `buf.ReadAngle()` | `buf.WriteAngle(v)` |
| `ns.ByteArray` | `buf.ReadByteArray(maxLen)` | `buf.WriteByteArray(v)` |

### Fixed-Size Arrays

```go
// read exactly N bytes (no length prefix)
data, err := buf.ReadFixedByteArray(256)

// write bytes without length prefix
err := buf.WriteFixedByteArray(data)
```

### Composite Types

| Type | Description | Wire Format |
| ---- | ----------- | ----------- |
| `ns.PrefixedArray[T]` | Length-prefixed array | VarInt length + elements |
| `ns.PrefixedOptional[T]` | Boolean-prefixed optional | Boolean + value (if true) |
| `ns.BitSet` | Dynamic bit set | VarInt length (longs) + longs |
| `ns.FixedBitSet` | Fixed-size bit set | ceil(n/8) bytes |
| `ns.IDSet` | Registry ID set | VarInt type + tag/IDs |

## Handling Optional Fields

Optional fields require manual handling with a boolean prefix:

```go
type ExamplePacket struct {
    HasValue ns.Boolean
    Value    ns.VarInt // only present if HasValue is true
}

func (p *ExamplePacket) Read(buf *ns.PacketBuffer) error {
    var err error
    if p.HasValue, err = buf.ReadBool(); err != nil {
        return err
    }
    if p.HasValue {
        p.Value, err = buf.ReadVarInt()
    }
    return err
}

func (p *ExamplePacket) Write(buf *ns.PacketBuffer) error {
    if err := buf.WriteBool(p.HasValue); err != nil {
        return err
    }
    if p.HasValue {
        return buf.WriteVarInt(p.Value)
    }
    return nil
}
```

> **Tip:** Use `composite.PrefixedOptional[T]` for cleaner optional handling. See the "Using Composite Types" section above.

## Handling Arrays

Arrays require reading the length prefix first, then iterating:

```go
type ExamplePacket struct {
    Count  ns.VarInt
    Values []ns.String
}

func (p *ExamplePacket) Read(buf *ns.PacketBuffer) error {
    var err error
    if p.Count, err = buf.ReadVarInt(); err != nil {
        return err
    }
    p.Values = make([]ns.String, p.Count)
    for i := range p.Values {
        if p.Values[i], err = buf.ReadString(32767); err != nil {
            return err
        }
    }
    return nil
}

func (p *ExamplePacket) Write(buf *ns.PacketBuffer) error {
    if err := buf.WriteVarInt(ns.VarInt(len(p.Values))); err != nil {
        return err
    }
    for _, v := range p.Values {
        if err := buf.WriteString(v); err != nil {
            return err
        }
    }
    return nil
}
```

> **Tip:** Use `composite.PrefixedArray[T]` for cleaner array handling. See the "Using Composite Types" section above.

## States and Directions

### States

| Constant | Value | Description |
| -------- | ----- | ----------- |
| `jp.StateHandshake` | 0 | Initial connection state |
| `jp.StateStatus` | 1 | Server list ping |
| `jp.StateLogin` | 2 | Authentication |
| `jp.StateConfiguration` | 3 | Server configuration (1.20.2+) |
| `jp.StatePlay` | 4 | Gameplay |

### Directions

| Constant | Description |
| -------- | ----------- |
| `jp.C2S` | Client to Server (serverbound) |
| `jp.S2C` | Server to Client (clientbound) |

## File Naming Convention

- `c2s_<state>.go` - Client-to-server packets for that state
- `s2c_<state>.go` - Server-to-client packets for that state

Examples: `c2s_handshaking.go`, `s2c_login.go`, `c2s_play.go`

## Complete Example

```go
package packets

import (
    jp "github.com/zeozeozeo/minego/internal/protocol/java_protocol"
    ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// C2SLoginStart initiates the login sequence.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Login_Start
type C2SLoginStart struct {
    // Player's username (max 16 characters)
    Username ns.String
    // Player's UUID
    PlayerUUID ns.UUID
}

func (p *C2SLoginStart) ID() ns.VarInt    { return 0x00 }
func (p *C2SLoginStart) State() jp.State  { return jp.StateLogin }
func (p *C2SLoginStart) Bound() jp.Bound  { return jp.C2S }

func (p *C2SLoginStart) Read(buf *ns.PacketBuffer) error {
    var err error
    if p.Username, err = buf.ReadString(16); err != nil {
        return err
    }
    p.PlayerUUID, err = buf.ReadUUID()
    return err
}

func (p *C2SLoginStart) Write(buf *ns.PacketBuffer) error {
    if err := buf.WriteString(p.Username); err != nil {
        return err
    }
    return buf.WriteUUID(p.PlayerUUID)
}

// ...other relevant packet definitions...
```

## References

- [Minecraft Protocol Wiki](https://minecraft.wiki/w/Java_Edition_protocol/Packets)
- [Protocol Data Types](https://minecraft.wiki/w/Java_Edition_protocol/Data_types)
