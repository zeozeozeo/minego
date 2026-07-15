# Mineflayer parity matrix

This file is the release checklist for MineGo's idiomatic equivalent of
Mineflayer 4.37.1. `Implemented` means a public, context-aware API has unit or
integration coverage. `Partial` is usable but does not yet match the complete
semantics. `Gated` means callers must check `Bot.Supports`. `Planned` is not a
release claim.

Last audited: 2026-07-15.

## Connection and extensibility

| Mineflayer capability | MineGo API | Status | Verification / gap |
| --- | --- | --- | --- |
| create bot, connect, end, reconnect | `New`, `Connect`, `Close`, `Done` | Partial | Offline login and disconnect tested; reconnect needs a new Bot |
| select protocol version | `Config.Version`, `SupportedVersions`, `Bot.Version` | Partial | 26.1 and 26.2 are compiled and status-detected; 1.21.11 (774) awaits an obfuscation/mappings dumper adapter and offline-server coverage |
| feature detection | `Bot.Supports` | Implemented | Immutable pack capabilities |
| plugins | `Plugin.Register`, `PluginContext`, `Bot.Service` | Implemented | Registration, services, normalized services, controlled raw packets |
| raw packet observation | `OnPacket`, `PluginContext.OnRawPacket` | Gated | Wire schemas remain pack-specific |
| connection telemetry | `Observability.Snapshot` | Implemented | Atomic wire packet/byte counters, action count, last-packet time |
| Microsoft authentication | `Microsoft`, `TokenStore` | Partial | Device login and certificates implemented; soak coverage pending |
| resource packs | `ResourcePackPolicy` | Partial | Accept/decline only; download and hash validation planned |

## Version-pack delivery ledger

| Release | Protocol | Generated data | Packet adapters / connection | Status |
| --- | --- | --- | --- | --- |
| 26.2 | 776 | checked-in under `internal/data/versions/v26_2` | implemented current adapter | supported |
| 26.1 | 775 | checked-in under `internal/data/versions/v26_1` | version-owned generated packet codecs; runtime ID-compatibility guard | supported; offline-server suite pending |
| 1.21.11 | 774 | blocked at authoritative dumper input | not started | not advertised |

`tools/mcgen -versions 26.1,1.21.11` now generates each release into its own
directory and never overwrites another release. It intentionally refuses to
advertise a pack until its packet schema and normalized adapters have tests.

## State and observation

| Mineflayer capability / event group | MineGo API | Status | Verification / gap |
| --- | --- | --- | --- |
| blocks, chunks, block updates | `World.Block`, `LoadedChunks`, `OnBlockChange`, `OnChunkLoad` | Partial | Basic chunk lifecycle; searches, block entities, raycast API pending |
| dimensions and respawn | `World.Dimension`, self state | Partial | Reset/respawn covered; dimension registry snapshot pending |
| bot position, health, hunger, oxygen, XP | `Self.State` | Partial | Core fields tracked; effects/attributes and richer events pending |
| players and entities | `Entities`, `Players` | Partial | Spawn/move/remove and tab-list identities; equipment, effects, vehicles pending |
| inventory and held item | `Inventory.Slots`, `Window`, `Selected`, `Select` | Partial | Player/crafting windows and component snapshots; general container transfers pending |
| chat/system messages | `Chat.OnMessage`, `Send`, `Command` | Partial | Signing supported; patterns, whispers, await, titles, completion pending |
| time | `World.Time`, `World.OnTime` | Implemented | Immutable world-clock snapshots from `S2CSetTime` |
| weather, tab list | — | Planned | No public service yet |
| scoreboards, teams, boss bars | — | Planned | No public service yet |
| sounds, particles, explosions | — | Planned | No public event yet |
| server settings/game rules | — | Planned | No public snapshot yet |

## Actions and interactions

| Mineflayer capability | MineGo API | Status | Verification / gap |
| --- | --- | --- | --- |
| cancellable coordinated actions | context-taking methods, internal leases | Implemented | Atomic claims, cancellation, explicit-action preemption tested |
| dig block and select tool | `Miner.Dig` | Partial | Reach, sight, timing, authoritative update; full enchant/effect timing pending |
| find and mine blocks | `Miner.Mine`, `Blocks`, `Tags` | Partial | Loaded search and frontier exploration; pickup and blacklisting pending |
| place block | `Builder.Place` | Partial | Hotbar, support, sight, update; orientation/scaffolding pending |
| use/activate item, block, entity | `Interaction.UseItem`, `ReleaseItem`, `ActivateBlock`, `ActivateEntity`, `Swing` | Implemented | Coordinated view/hands, reach validation, sequence IDs, both hands and relative entity hits tested |
| attack/combat | `Combat.Attack`, `Fight`, `OnEvent` | Implemented | Reach/view validation, cooldown loop, swing control and normalized damage events |
| equip, toss, transfer, click window | `Inventory.Equip`, `Toss`, `Transfer`, `Click`, `CreativeSet` | Implemented | State-ID clicks wait for authoritative updates; equipment, partial counts, quick move, drop and creative slots covered |
| crafting and recipes | `Crafter.RecipesFor`, `Craft`, `RegisterRecipe` | Partial | Generated vanilla shaped/shapeless recipes, recursive dependencies, 2x2/3x3 windows; custom server recipes and processing stations pending |
| containers and furnaces | `Containers.Open`, `Container`, furnace helpers | Implemented | Window lifecycle, slots, carried item, properties, close, fuel/input/output and stale-handle checks |
| enchanting, anvils, villagers | `ChooseEnchantment`, `Rename`, `OpenEntity`, `SelectTrade` | Implemented | Button/rename/trade packets, authoritative window operations and immutable offer snapshots |
| beds, fishing, books, signs | `Special.Sleep`, `Wake`, `Fish`, `EditBook`, `UpdateSign` | Implemented | Separate cast/reel control, book/sign bounds and signing, bed enter/leave actions |
| mount/dismount and creative flight | `Riding.Mount`, `Dismount`, `MoveVehicle`, `Paddle`, `SetCreativeFlight`, `FlyTo` | Implemented | Authoritative passenger/ability state, vehicle movement and creative-mode validation |

## Navigation and automation

| Capability | MineGo API | Status | Verification / gap |
| --- | --- | --- | --- |
| exact, near, adjacent goals | `GoalBlock`, `GoalNear`, `GoalAdjacent` | Implemented | Deterministic path tests |
| walk, diagonal, jump, drop, sprint | `Navigator.Navigate` | Partial | Basic collision-aware execution implemented |
| parkour, swimming, climbing, doors | navigation options/moves | Partial | Bounded forms; full simulation and hazards pending |
| breaking during routes | `AllowBreaking` | Partial | Multi-block body clearance through `Miner.Dig`; hazard policy pending |
| dynamic rerouting and correction | navigator executor | Partial | D*-Lite repairs relevant changed edges for standard goals; custom goals fall back to bounded A* |
| follow, explore, escape, composite goals | `Follow`, `Explore` | Partial | Entity/player following and frontier exploration; escape/composite goals pending |
| bridge and pillar | `AllowPlacing`, `TemporaryBlocks` | Partial | Allowlisted acquisition, bridge and jump-pillar actions; route blocks remain in place |
| blueprint building | `Blueprint`, `FindSite`, `Build` | Partial | Relative blueprints, materials, obstruction handling and ordered placement; importers/orientation pending |
| incremental segmented planner | navigator planner | Implemented | Retained D*-Lite state inside loaded segments with bounded A* frontier fallback |
| elytra flight | `Elytra.Fly` | Partial | Automatic equipment, rockets, clearance correction and safe landing; durability policy and server soak pending |

## Event semantics

Subscriptions currently return idempotent unsubscribe functions and copy their
handler list before delivery. Implemented normalized events are disconnect, raw
packet, chat message, block change, chunk load, entity change, navigation
progress, and mining progress. Mineflayer event headings not represented above
(window lifecycle, physics tick, health/death, weather/time, player list,
scoreboard/team/boss bar, sound/particle/explosion, title/action bar, resource
pack, and specialized interaction events) remain planned.
