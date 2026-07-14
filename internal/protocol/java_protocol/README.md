# Java Protocol

This package implements the low-level Minecraft: Java Edition protocol, providing primitives for reading and writing packets over TCP connections.

## Overview

The Minecraft server accepts connections from TCP clients and communicates using packets. A packet is a sequence of bytes where the meaning depends on both its packet ID and the current connection state.

```plain
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Packet Structure                                  │
└─────────────────────────────────────────────────────────────────────────────┘

Without compression:
┌──────────────────┬──────────────────┬──────────────────────────────────────┐
│  Packet Length   │    Packet ID     │                Data                  │
│    (VarInt)      │    (VarInt)      │            (ByteArray)               │
└──────────────────┴──────────────────┴──────────────────────────────────────┘

With compression (when size >= threshold):
┌──────────────────┬──────────────────┬────────────────────────────────────────┐
│  Packet Length   │   Data Length    │    Compressed (Packet ID + Data)       │
│    (VarInt)      │    (VarInt)      │              (zlib)                    │
└──────────────────┴──────────────────┴────────────────────────────────────────┘

With compression (when size < threshold):
┌──────────────────┬──────────────────┬──────────────────┬─────────────────────┐
│  Packet Length   │   Data Length    │    Packet ID     │        Data         │
│    (VarInt)      │   (VarInt = 0)   │    (VarInt)      │     (ByteArray)     │
└──────────────────┴──────────────────┴──────────────────┴─────────────────────┘
```

## Connection States

The protocol has 5 states, with automatic transitions:

```plain
                    ┌─────────────┐
                    │  Handshake  │  (Initial state)
                    └──────┬──────┘
                           │
              ┌────────────┴────────────┐
              │                         │
              ▼                         ▼
       ┌─────────────┐           ┌─────────────┐
       │   Status    │           │    Login    │
       │  (ping/SLP) │           │   (auth)    │
       └─────────────┘           └──────┬──────┘
                                        │
                                        ▼
                                 ┌─────────────┐
                                 │Configuration│
                                 │  (1.20.2+)  │
                                 └──────┬──────┘
                                        │
                                        ▼
                                 ┌─────────────┐
                                 │    Play     │
                                 │  (in-game)  │
                                 └─────────────┘
```

## Package Structure

### `conn.go` - Encrypted Connection Wrapper

Wraps `net.Conn` with transparent AES/CFB8 encryption/decryption:

```go
// create connection wrapper
conn := java_protocol.NewConn(netConn)

// enable encryption (after key exchange during login)
conn.Encryption().Enable(sharedSecret)

// read/Write automatically encrypt/decrypt when enabled
conn.Read(buf)   // decrypts if enabled
conn.Write(data) // encrypts if enabled
```

### `packet.go` - Packet Types

Defines `Packet` interface and `WirePacket` struct:

```go
// Packet is the interface all typed packets implement
// Each packet knows its ID, state, and direction
type Packet interface {
    ID() ns.VarInt
    State() State
    Bound() Bound
    Read(buf *ns.PacketBuffer) error
    Write(buf *ns.PacketBuffer) error
}

// WirePacket is the raw wire format (what actually goes over the network)
type WirePacket struct {
    Length   ns.VarInt
    PacketID ns.VarInt
    Data     ns.ByteArray
}

// Example packet implementation
type LoginStartPacket struct {
    Username ns.String
}

func (p *LoginStartPacket) ID() ns.VarInt   { return 0x00 }
func (p *LoginStartPacket) State() State    { return StateLogin }
func (p *LoginStartPacket) Bound() Bound    { return C2S }
func (p *LoginStartPacket) Read(buf *ns.PacketBuffer) error {
    var err error
    p.Username, err = buf.ReadString(16)
    return err
}
func (p *LoginStartPacket) Write(buf *ns.PacketBuffer) error {
    return buf.WriteString(p.Username)
}

// Convert to wire format, then write to connection
wire, err := java_protocol.ToWire(&LoginStartPacket{Username: "Player"})
err = wire.WriteTo(conn, threshold) // handles compression automatically
```

### `tcp_client.go` - Protocol Client

Minimal client for connecting to Minecraft servers:

```go
client := java_protocol.NewTCPClient()

// connect (also handles SRV record resolution)
host, port, err := client.Connect("example.com")

// set protocol state
client.SetState(java_protocol.StateLogin)

// enable compression after server requests it
client.SetCompressionThreshold(256)

// write packets (packet knows its own state/bound/id)
client.WritePacket(&HandshakePacket{...})

// read raw wire packet
wire, err := client.ReadWirePacket()

// deserialize using generics (type-safe, recommended)
login, err := java_protocol.ReadPacket[LoginSuccessPacket](wire)

// or deserialize into a pre-allocated struct
var loginSuccess LoginSuccessPacket
err = wire.ReadInto(&loginSuccess)
```

## Protocol States

| State | Value | Description |
| ----- | ----- | ----------- |
| `StateHandshake` | 0 | Initial state, client sends intention |
| `StateStatus` | 1 | Server List Ping (SLP) |
| `StateLogin` | 2 | Authentication and encryption |
| `StateConfiguration` | 3 | Server configuration (1.20.2+) |
| `StatePlay` | 4 | Gameplay packets |

## Packet Direction

| Direction | Constant | Description |
| --------- | -------- | ----------- |
| Serverbound | `C2S` | Client → Server |
| Clientbound | `S2C` | Server → Client |

## Compression

Compression is enabled by the server via `Set Compression` packet during login:

```go
// After receiving Set Compression packet
client.SetCompressionThreshold(threshold)

// Threshold behavior:
// -1: compression disabled
// 0+: packets >= threshold bytes are zlib compressed
```

## Encryption

Encryption is enabled during the login sequence after key exchange:

```go
// After completing Encryption Request/Response handshake
sharedSecret := generateSharedSecret()
client.Conn().Encryption().Enable(sharedSecret)

// All subsequent packets are encrypted with AES/CFB8
```

## Address Resolution

The `Connect` method automatically resolves Minecraft server addresses:

1. If port is specified (`host:port`), uses it directly
2. Otherwise, looks up SRV record `_minecraft._tcp.<host>`
3. Falls back to default port 25565

```go
// All equivalent:
client.Connect("mc.example.com")        // uses SRV or :25565
client.Connect("mc.example.com:25565")  // explicit port
client.Connect("play.hypixel.net")      // SRV → mc.hypixel.net:25565
```

## Debug Logging

Enable debug logging to trace packet I/O:

```go
client.EnableDebug(true)
client.SetLogger(log.New(os.Stdout, "[MC] ", log.LstdFlags))

// Output:
// [MC] -> send: state=2 bound=0 id=0x00 len=23 bytes=...
// [MC] <- recv: length=42
// [MC] <- recv: compressed id=0x02 data_len=38
```

## Data Types

The `net_structures` subpackage provides Minecraft protocol data types:

| Type | Go Type | Description |
| ---- | ------- | ----------- |
| `VarInt` | `int32` | Variable-length 32-bit integer |
| `VarLong` | `int64` | Variable-length 64-bit integer |
| `String` | `string` | UTF-8 with VarInt length prefix |
| `Boolean` | `bool` | Single byte (0x00/0x01) |
| `Int8`-`Int64` | `int8`-`int64` | Big-endian signed integers |
| `Uint8`-`Uint16` | `uint8`-`uint16` | Big-endian unsigned integers |
| `Float32`/`Float64` | `float32`/`float64` | IEEE 754 floats |
| `Position` | struct | Block coordinates packed into 64 bits |
| `UUID` | `[16]byte` | 128-bit unique identifier |
| `Angle` | `uint8` | Rotation (1/256 of full turn) |
| `Identifier` | `string` | Namespaced ID (e.g., `minecraft:stone`) |
| `ByteArray` | `[]byte` | Raw bytes with VarInt length prefix |

## Implementing Packets

Define packets by implementing the `Packet` interface:

```go
type HandshakePacket struct {
    ProtocolVersion ns.VarInt
    ServerAddress   ns.String
    ServerPort      ns.Uint16
    NextState       ns.VarInt
}

// Metadata - each packet knows its ID, state, and direction
func (p *HandshakePacket) ID() ns.VarInt   { return 0x00 }
func (p *HandshakePacket) State() State    { return StateHandshake }
func (p *HandshakePacket) Bound() Bound    { return C2S }

func (p *HandshakePacket) Read(buf *ns.PacketBuffer) error {
    var err error
    p.ProtocolVersion, err = buf.ReadVarInt()
    if err != nil { return err }
    p.ServerAddress, err = buf.ReadString(255)
    if err != nil { return err }
    p.ServerPort, err = buf.ReadUint16()
    if err != nil { return err }
    p.NextState, err = buf.ReadVarInt()
    return err
}

func (p *HandshakePacket) Write(buf *ns.PacketBuffer) error {
    if err := buf.WriteVarInt(p.ProtocolVersion); err != nil { return err }
    if err := buf.WriteString(p.ServerAddress); err != nil { return err }
    if err := buf.WriteUint16(p.ServerPort); err != nil { return err }
    return buf.WriteVarInt(p.NextState)
}
```

## Reading Packets

Three ways to read packets:

```go
// 1. Generic function (recommended) - type-safe, returns concrete type
wire, _ := client.ReadWirePacket()
login, err := java_protocol.ReadPacket[LoginSuccessPacket](wire)

// 2. ReadInto - fill an existing struct
var login LoginSuccessPacket
wire, _ := client.ReadWirePacket()
err := wire.ReadInto(&login)

// 3. Manual - read wire packet, switch on ID
wire, _ := client.ReadWirePacket()
switch wire.PacketID {
case 0x00:
    var disconnect DisconnectPacket
    wire.ReadInto(&disconnect)
case 0x02:
    var success LoginSuccessPacket
    wire.ReadInto(&success)
}
```

## Packet Size Limits

- Maximum packet size: 2,097,151 bytes (2^21 - 1)
- Maximum uncompressed serverbound size: 8,388,608 bytes (2^23)
- Packet length field: max 3 bytes

## References

- [Minecraft Protocol - Minecraft Wiki](https://minecraft.wiki/w/Java_Edition_protocol/Packets)
- [Protocol Data Types - Minecraft Wiki](https://minecraft.wiki/w/Java_Edition_protocol/Data_types)
- [Protocol Encryption - Minecraft Wiki](https://minecraft.wiki/w/Java_Edition_protocol/Encryption)
