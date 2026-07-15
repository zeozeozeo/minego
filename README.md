# MineGo

MineGo is a Go 1.25+ library for autonomous Minecraft Java Edition clients.
The currently compiled packs target **Minecraft 26.1 / protocol 775** and
**Minecraft 26.2 / protocol 776**, and include
the protocol, authentication, generated registry data, world state, movement,
pathfinding, and mining code in this repository. It has no go-mclib module
dependency.

[demo.webm](https://github.com/user-attachments/assets/bddff7a1-6c1b-4dc3-89c5-aa94579e7642)

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

err = bot.Navigator.Follow(ctx,
    minego.FollowTarget{PlayerName: "Steve"},
    minego.FollowOptions{Distance: 3},
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

With `AllowBreaking` and `AllowPlacing`, route nodes may dig both body cells,
bridge gaps, or pillar upward. `TemporaryBlocks` restricts usable materials;
`AcquireTemporary` can mine an allowlisted block before executing a route that
is short on scaffolding. Long routes are split at safe loaded-chunk frontiers,
and unrelated block updates do not trigger a repair. `Navigator.Follow`
accepts entity IDs, player UUIDs, or tab-list names, while
`Navigator.Explore` exposes bounded frontier exploration.

## Crafting, elytra, and blueprints

`Bot.Crafter` provides generated vanilla shaped and shapeless recipes,
recursive dependency resolution, inventory 2x2 crafting, and authoritative
crafting-table windows. Pass a table position for 3x3 recipes:

```go
table := minego.BlockPos{X: 12, Y: 64, Z: 8}
_, err := bot.Crafter.Craft(ctx, "oak_door", 1,
    minego.CraftOptions{Table: &table, Recursive: true})
```

`Bot.Elytra.Fly` handles equipment, launch, rockets, clearance correction, and
a safe landing near a standard block goal. `Builder.FindSite`,
`Builder.Materials`, and `Builder.Build` support reusable relative blueprints.
The `examples/wood-house` program mines any log family, crafts matching
materials using both grid sizes, finds a clear site, and builds a 7x7 house.

## Version packs and generation

`version.Pack` isolates protocol IDs, packet factories, registries, block states,
and collision data. Applications select releases with a string such as
`Version: "26.1"` or `Version: "26.2"`; the empty default probes server status and selects by protocol.
`minego.SupportedVersions()` reports the packs actually compiled into the build,
and unsupported servers return a typed error rather than trying guessed codecs.
Regenerate one or more isolated data packs with `go run ./tools/mcgen -versions
1.21.11,26.1,26.2`. Generation records a packet-adapter checklist next to each pack;
data alone is deliberately not treated as connection support.

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

The committed [Mineflayer parity matrix](docs/mineflayer-parity.md) is the
feature-completeness checklist. The inventory API covers player and crafting
window synchronization, held-slot selection, crafting transactions,
mining-tool choice, and hotbar block placement. General container transfers,
furnace processing, custom server recipe discovery, Bedrock, Realms discovery,
proxies, and server hosting are not part of this release.
