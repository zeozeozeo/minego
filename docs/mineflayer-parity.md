# Mineflayer parity matrix

This file is the release checklist for MineGo's idiomatic equivalent of
Mineflayer 4.37.1. `Implemented` means a public, context-aware API has unit or
integration coverage. `Partial` is usable but does not yet match the complete
semantics. `Gated` means callers must check `Bot.Supports`. `Planned` is not a
release claim.

Last audited: 2026-07-14.

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
| players and entities | `Entities` | Partial | Spawn/move/remove basics; metadata, equipment, effects, vehicles pending |
| inventory and held item | `Inventory.Slots`, `Selected`, `Select` | Partial | Synchronization and selection; authoritative window transactions pending |
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
| use/activate item, block, entity | — | Planned | Packet primitives exist but no stable service API |
| attack/combat | — | Planned | No combat service yet |
| equip, toss, transfer, click window | — | Planned | Requires transaction/state-ID layer |
| crafting and recipes | — | Planned | Requires version-owned recipe data and windows |
| containers and furnaces | — | Planned | Requires authoritative windows |
| enchanting, anvils, villagers | — | Planned | Requires authoritative windows |
| beds, fishing, books, signs | — | Planned | No high-level API yet |
| mount/dismount and creative flight | — | Planned | No high-level API yet |

## Navigation and automation

| Capability | MineGo API | Status | Verification / gap |
| --- | --- | --- | --- |
| exact, near, adjacent goals | `GoalBlock`, `GoalNear`, `GoalAdjacent` | Implemented | Deterministic path tests |
| walk, diagonal, jump, drop, sprint | `Navigator.Navigate` | Partial | Basic collision-aware execution implemented |
| parkour, swimming, climbing, doors | navigation options/moves | Partial | Bounded forms; full simulation and hazards pending |
| breaking during routes | `AllowBreaking` | Partial | Uses `Miner.Dig`; dependency-aware repair pending |
| dynamic rerouting and correction | navigator executor | Partial | Full replans today; D*-Lite edge repair pending |
| follow, explore, escape, composite goals | — | Planned | Goal/process APIs pending |
| bridge and pillar | — | Planned | Placement primitive exists; process APIs pending |
| blueprint building | — | Planned | Native `Blueprint` and importers pending |
| incremental segmented planner | — | Planned | Current planner is bounded A* |

## Event semantics

Subscriptions currently return idempotent unsubscribe functions and copy their
handler list before delivery. Implemented normalized events are disconnect, raw
packet, chat message, block change, chunk load, entity change, navigation
progress, and mining progress. Mineflayer event headings not represented above
(window lifecycle, physics tick, health/death, weather/time, player list,
scoreboard/team/boss bar, sound/particle/explosion, title/action bar, resource
pack, and specialized interaction events) remain planned.
