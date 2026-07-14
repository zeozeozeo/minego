# MineGo

MineGo is a Go 1.25+ library for autonomous Minecraft Java Edition clients.
This first version targets **Minecraft 26.2 / protocol 776 only** and includes
the protocol, authentication, generated registry data, world state, movement,
pathfinding, and mining code in this repository. It has no go-mclib module
dependency.

## Connect and chat

```go
bot, err := minego.New(minego.Config{
    Address: "localhost:25565",
    Auth:    minego.Offline("MineGo"),
})
if err != nil { log.Fatal(err) }

ctx := context.Background()
if err := bot.Connect(ctx); err != nil { log.Fatal(err) }
defer bot.Close()
if err := bot.WaitReady(ctx); err != nil { log.Fatal(err) }
if err := bot.Chat.Send(ctx, "hello from Go"); err != nil { log.Fatal(err) }
```

For online-mode servers, register an Azure public/native application that
allows the device-code flow and supply its client ID:

```go
auth := minego.Microsoft(clientID, accountHint, func(code minego.DeviceCode) {
    fmt.Printf("Open %s and enter %s\n", code.VerificationURI, code.UserCode)
})
```

Tokens are cached by the internal restrictive-permission file store by default;
applications can provide a `TokenStore` or use `MemoryTokenStore`. MineGo also
negotiates a Mojang player certificate and signs outgoing player chat.

## World, navigation, and mining

World queries distinguish an unloaded location (`ok == false`) from air. Public
snapshots are safe to read concurrently, and subscriptions return an idempotent
unsubscribe function.

```go
block, loaded := bot.World.Block(minego.BlockPos{X: 0, Y: 64, Z: 0})
unsubscribe := bot.World.OnBlockChange(func(change minego.BlockChange) { /* ... */ })
defer unsubscribe()

_, err = bot.Navigator.Navigate(ctx,
    minego.GoalNear{Position: minego.BlockPos{X: 100, Y: 64, Z: 40}, Radius: 2},
    minego.NavigationOptions{
        Sprint: true, MaxDrop: 3,
        AllowParkour: true, MaxParkourGap: 2,
    },
)

result, err := bot.Miner.Mine(ctx,
    minego.Tags("minecraft:diamond_ores"), 3, minego.MineOptions{},
)

placed, err := bot.Builder.Place(ctx,
    minego.BlockPos{X: 10, Y: 64, Z: 10},
    minego.PlaceOptions{Item: "cobblestone"},
)
```

Navigation uses loaded authoritative block states and emits ordinary 20 Hz
movement and input packets. It supports cardinal and diagonal walking,
sprinting, one-block jumps, safe drops, water and shore exits, climbable
blocks, wooden-door interaction, dynamic replanning, opt-in block breaking,
and opt-in bounded sprint-jump parkour. Diagonal paths reject blocked-corner
cuts. `Builder.Place` selects a matching hotbar block, validates reach,
replaceability, line of sight, and a supporting face, then waits for the
authoritative block update. Navigation never teleports the bot. Exploratory
mining walks toward safe loaded-chunk frontiers to reveal normal terrain; it
does not branch-mine hidden blocks.

## Version packs and generation

`version.Pack` isolates protocol IDs, packet factories, registries, block states,
and collision data. `versions.V26_2` is selected by default. A future version is
added as another pack while the `Bot` and service APIs stay stable.

Regenerate all official inputs and Go tables with:

```text
go generate ./...
```

The `tools/mcgen` command is fully written in Go. It downloads Mojang client and
server artifacts, verifies their SHA-1 hashes, runs vanilla datagen, extracts
the bundled server and libraries with `archive/zip`, compiles/runs `Dump.java`,
and invokes the Go table generators. Java and `javac` are the only external
runtime requirements; no shell, curl, Python, symlinks, or go-mclib checkout is
used. See `docs/adding-a-version.md` for packet-schema changes.

## Current scope

The inventory API covers synchronization, held-slot selection, mining-tool
choice, and hotbar block placement. Container clicks, crafting, arbitrary
inventory transactions, automated bridging and pillaring, Bedrock, Realms
discovery, proxies, and server hosting are not part of this release.
