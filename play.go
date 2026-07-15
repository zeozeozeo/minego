package minego

import (
	"fmt"

	"github.com/zeozeozeo/minego/internal/data/chunks"
	dataentities "github.com/zeozeozeo/minego/internal/data/versions/v26_2/entities"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/lang"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packet_ids"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	jp "github.com/zeozeozeo/minego/internal/protocol/java_protocol"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"github.com/zeozeozeo/minego/internal/protocol/nbt"
)

func (b *Bot) handlePlay(w *jp.WirePacket) error {
	if handled, err := b.handleLegacyObservation(w); handled {
		return err
	}
	switch w.PacketID {
	case packet_ids.S2CKeepAlivePlayID:
		var p packets.S2CKeepAlivePlay
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		return b.client.WritePacket(&packets.C2SKeepAlivePlay{KeepAliveId: p.KeepAliveId})
	case packet_ids.S2CPingPlayID:
		var p packets.S2CPingPlay
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		return b.client.WritePacket(&packets.C2SPongPlay{Id: p.Id})
	case packet_ids.S2CDisconnectPlayID:
		var p packets.S2CDisconnectPlay
		_ = w.ReadInto(&p)
		return fmt.Errorf("disconnected: %s", p.Reason.String())
	case packet_ids.S2CStartConfigurationID:
		if err := b.client.WritePacket(&packets.C2SConfigurationAcknowledged{}); err != nil {
			return err
		}
		b.client.SetState(jp.StateConfiguration)
		return nil
	case packet_ids.S2CLoginID:
		return b.handleJoin(w)
	case packet_ids.S2CRespawnID:
		return b.handleRespawn(w)
	case packet_ids.S2CPlayerPositionID:
		return b.handlePlayerPosition(w)
	case packet_ids.S2CSetHealthID:
		var p packets.S2CSetHealth
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.Self.update(func(s *SelfState) {
			s.Health = float32(p.Health)
			s.Food = int32(p.Food)
			s.Saturation = float32(p.FoodSaturation)
		})
		if b.beginRespawn(float32(p.Health)) {
			// Action 0 is Perform Respawn. Suppress movement until the server's
			// new authoritative position arrives.
			b.Navigator.Stop()
			return b.client.WritePacket(&packets.C2SClientCommand{ActionId: 0})
		}
	case packet_ids.S2CSetTimeID:
		var p packets.S2CSetTime
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		time := TimeState{WorldAge: int64(p.WorldAge), Clocks: make([]ClockState, len(p.ClockUpdates))}
		for i, clock := range p.ClockUpdates {
			time.Clocks[i] = ClockState{ID: int32(clock.WorldClock), TotalTicks: int64(clock.TotalTicks), PartialTick: float32(clock.PartialTick), Rate: float32(clock.Rate)}
		}
		b.World.updateTime(time)
	case packet_ids.S2CGameEventID:
		return b.handleGameEvent(w)
	case packet_ids.S2CTabListID:
		var p packets.S2CTabList
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.TabList.update(TabListState{Header: p.Header.Render(lang.Translate), Footer: p.Footer.Render(lang.Translate)})
	case packet_ids.S2CBossEventID:
		return b.handleBossBar(w)
	case packet_ids.S2CSetObjectiveID:
		return b.handleObjective(w.Data)
	case packet_ids.S2CSetDisplayObjectiveID:
		return b.handleDisplayObjective(w)
	case packet_ids.S2CSetScoreID:
		return b.handleScore(w.Data)
	case packet_ids.S2CResetScoreID:
		return b.handleResetScore(w)
	case packet_ids.S2CSetPlayerTeamID:
		return b.handleTeam(w.Data)
	case packet_ids.S2CSoundID:
		return b.handleSound(w.Data, false)
	case packet_ids.S2CSoundEntityID:
		return b.handleSound(w.Data, true)
	case packet_ids.S2CExplodeID:
		return b.handleExplosion(w.Data)
	case packet_ids.S2CChangeDifficultyID:
		var p packets.S2CChangeDifficulty
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.Server.update(func(s *ServerState) {
			s.Difficulty = uint8(p.Difficulty)
			s.DifficultyLocked = bool(p.DifficultyLocked)
		})
	case packet_ids.S2CSetSimulationDistanceID:
		var p packets.S2CSetSimulationDistance
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.Server.update(func(s *ServerState) { s.SimulationDistance = int32(p.SimulationDistance) })
	case packet_ids.S2CServerDataID:
		var p packets.S2CServerData
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.Server.update(func(s *ServerState) {
			s.MOTD = p.Motd.Render(lang.Translate)
			s.Icon = nil
			if p.Icon.Present {
				s.Icon = append([]byte(nil), p.Icon.Value...)
			}
		})
	case packet_ids.S2CGameRuleValuesID:
		if b.Version().Protocol >= 775 {
			var p packets.S2CGameRuleValues
			if err := w.ReadInto(&p); err != nil {
				return err
			}
			b.Server.update(func(s *ServerState) {
				if s.GameRules == nil {
					s.GameRules = map[string]string{}
				}
				for _, v := range p.Values {
					s.GameRules[string(v.Key)] = string(v.Value)
				}
			})
		}
	case packet_ids.S2CUpdateMobEffectID:
		var p packets.S2CUpdateMobEffect
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		if int32(p.EntityId) == b.Self.State().EntityID {
			b.Self.update(func(s *SelfState) {
				s.Effects[int32(p.EffectId)] = Effect{int32(p.EffectId), int32(p.Amplifier), int32(p.Duration), int8(p.Flags)}
			})
		}
	case packet_ids.S2CRemoveMobEffectID:
		var p packets.S2CRemoveMobEffect
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		if int32(p.EntityId) == b.Self.State().EntityID {
			b.Self.update(func(s *SelfState) { delete(s.Effects, int32(p.EffectId)) })
		}
	case packet_ids.S2CLevelChunkWithLightID:
		return b.handleChunk(w)
	case packet_ids.S2CForgetLevelChunkID:
		return b.handleForgetChunk(w)
	case packet_ids.S2CBlockUpdateID:
		return b.handleBlockUpdate(w)
	case packet_ids.S2CSectionBlocksUpdateID:
		return b.handleSectionUpdate(w)
	case packet_ids.S2CChunkBatchFinishedID:
		return b.client.WritePacket(&packets.C2SChunkBatchReceived{ChunksPerTick: 25})
	case packet_ids.S2CSetChunkCacheCenterID:
		var p packets.S2CSetChunkCacheCenter
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.World.mu.Lock()
		b.World.center = chunkKey{int32(p.ChunkX), int32(p.ChunkZ)}
		b.World.mu.Unlock()
	case packet_ids.S2CSetChunkCacheRadiusID:
		var p packets.S2CSetChunkCacheRadius
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.World.mu.Lock()
		b.World.viewDistance = int32(p.ViewDistance)
		b.World.mu.Unlock()
		b.Server.update(func(s *ServerState) { s.ViewDistance = int32(p.ViewDistance) })
	case packet_ids.S2CAddEntityID:
		return b.handleAddEntity(w)
	case packet_ids.S2CMoveEntityPosID:
		return b.handleMoveEntity(w, false)
	case packet_ids.S2CMoveEntityPosRotID:
		return b.handleMoveEntity(w, true)
	case packet_ids.S2CMoveEntityRotID:
		return b.handleRotateEntity(w)
	case packet_ids.S2CTeleportEntityID:
		return b.handleTeleportEntity(w)
	case packet_ids.S2CSetEntityMotionID:
		return b.handleEntityMotion(w)
	case packet_ids.S2CSetEntityDataID:
		return b.handleEntityData(w)
	case packet_ids.S2CRemoveEntitiesID:
		return b.handleRemoveEntities(w)
	case packet_ids.S2CContainerSetContentID:
		return b.handleInventoryContent(w)
	case packet_ids.S2CContainerSetSlotID:
		return b.handleInventorySlot(w)
	case packet_ids.S2CContainerSetDataID:
		var p packets.S2CContainerSetData
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.Inventory.mu.Lock()
		if int32(p.WindowId) == b.Inventory.window.ID {
			if b.Inventory.window.Properties == nil {
				b.Inventory.window.Properties = make(map[int16]int16)
			}
			b.Inventory.window.Properties[int16(p.Property)] = int16(p.Value)
		}
		snapshot := cloneWindow(b.Inventory.window)
		b.Inventory.mu.Unlock()
		b.Inventory.onWindow.emit(WindowChange{Window: snapshot})
	case packet_ids.S2CMerchantOffersID:
		var p packets.S2CMerchantOffers
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.Inventory.mu.Lock()
		if int32(p.WindowId) == b.Inventory.window.ID {
			b.Inventory.window.Offers = append([]byte(nil), p.Data...)
		}
		snapshot := cloneWindow(b.Inventory.window)
		b.Inventory.mu.Unlock()
		b.Inventory.onWindow.emit(WindowChange{Window: snapshot})
	case packet_ids.S2COpenScreenID:
		var p packets.S2COpenScreen
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.Inventory.mu.Lock()
		b.Inventory.window = WindowSnapshot{ID: int32(p.WindowId), Type: int32(p.WindowType), Title: p.WindowTitle.Render(lang.Translate), Properties: make(map[int16]int16)}
		snapshot := cloneWindow(b.Inventory.window)
		b.Inventory.mu.Unlock()
		b.Inventory.onWindow.emit(WindowChange{Window: snapshot})
	case packet_ids.S2CContainerCloseID:
		var p packets.S2CContainerClose
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.Inventory.mu.Lock()
		b.Inventory.window = WindowSnapshot{}
		snapshot := cloneWindow(b.Inventory.window)
		b.Inventory.mu.Unlock()
		b.Inventory.onWindow.emit(WindowChange{Window: snapshot})
	case packet_ids.S2CSetHeldSlotID:
		var p packets.S2CSetHeldSlot
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.Inventory.mu.Lock()
		b.Inventory.selected = int(p.Slot)
		b.Inventory.mu.Unlock()
		b.Self.update(func(s *SelfState) { s.SelectedSlot = int(p.Slot) })
	case packet_ids.S2CPlayerAbilitiesID:
		var p packets.S2CPlayerAbilities
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.Self.update(func(s *SelfState) {
			s.Flying = int8(p.Flags)&2 != 0
			s.CanFly = int8(p.Flags)&4 != 0
			s.FlyingSpeed = float32(p.FlyingSpeed)
		})
	case packet_ids.S2CSetPassengersID:
		return b.handleSetPassengers(w.Data)
	case packet_ids.S2CDamageEventID:
		var p packets.S2CDamageEvent
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.Combat.onEvent.emit(CombatEvent{Kind: "damage", EntityID: int32(p.EntityId), SourceID: int32(p.SourceCauseId)})
	case packet_ids.S2CLevelParticlesID:
		return b.handleParticle(w.Data)
	case packet_ids.S2CPlayerInfoUpdateID:
		// This packet's action-bitset shape is not expressible by the generic
		// packet generator, so the stable Players service decodes its payload.
		_ = b.Players.decodeUpdate(w.Data)
	case packet_ids.S2CPlayerInfoRemoveID:
		_ = b.Players.decodeRemove(w.Data)
	case packet_ids.S2CSystemChatID:
		var p packets.S2CSystemChat
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.Chat.onMessage.emit(ChatMessage{Kind: ChatSystem, Text: p.Content.Render(lang.Translate)})
	case packet_ids.S2CDisguisedChatID:
		var p packets.S2CDisguisedChat
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.Chat.onMessage.emit(ChatMessage{Kind: ChatDisguised, Sender: p.SenderName.Render(lang.Translate), Text: p.Message.Render(lang.Translate)})
	case packet_ids.S2CPlayerChatID:
		var p packets.S2CPlayerChat
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		text := string(p.Body.Content)
		if p.UnsignedContent.Present {
			text = p.UnsignedContent.Value.Render(lang.Translate)
		}
		b.Chat.onMessage.emit(ChatMessage{Kind: ChatPlayer, Sender: p.Sender.String(), Text: text, Verified: p.Signature.Present})
	}
	return nil
}

func (b *Bot) handleSetPassengers(data []byte) error {
	buf := ns.NewReader(data)
	vehicle, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	selfID, aboard := b.Self.State().EntityID, false
	for index := int32(0); index < int32(count); index++ {
		id, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		if int32(id) == selfID {
			aboard = true
		}
	}
	b.Self.update(func(s *SelfState) {
		if aboard {
			s.VehicleID = int32(vehicle)
		} else if s.VehicleID == int32(vehicle) {
			s.VehicleID = 0
		}
	})
	return nil
}

func (b *Bot) handleJoin(w *jp.WirePacket) error {
	var p packets.S2CLogin
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	b.World.reset(string(p.DimensionName))
	b.Self.update(func(s *SelfState) { s.EntityID = int32(p.EntityId); s.GameMode = uint8(p.GameMode) })
	b.Server.update(func(s *ServerState) {
		s.Hardcore = bool(p.IsHardcore)
		s.MaxPlayers = int32(p.MaxPlayers)
		s.ViewDistance = int32(p.ViewDistance)
		s.SimulationDistance = int32(p.SimulationDistance)
		s.ReducedDebugInfo = bool(p.ReducedDebugInfo)
		s.RespawnScreen = bool(p.EnableRespawnScreen)
		s.LimitedCrafting = bool(p.DoLimitedCrafting)
		s.Debug = bool(p.IsDebug)
		s.Flat = bool(p.IsFlat)
		s.OnlineMode = bool(p.OnlineMode)
		s.EnforcesSecureChat = bool(p.EnforcesSecureChat)
	})
	return nil
}
func (b *Bot) handleRespawn(w *jp.WirePacket) error {
	var p packets.S2CRespawn
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	b.World.reset(string(p.DimensionName))
	b.Entities.mu.Lock()
	b.Entities.values = map[int32]Entity{}
	b.Entities.mu.Unlock()
	b.Self.update(func(s *SelfState) {
		s.GameMode = uint8(p.GameMode)
		s.Effects = map[int32]Effect{}
		s.Velocity = Vec3{}
		s.OnGround = false
	})
	return nil
}

func (b *Bot) beginRespawn(health float32) bool {
	return health <= 0 && b.respawning.CompareAndSwap(false, true)
}

func (b *Bot) handlePlayerPosition(w *jp.WirePacket) error {
	var p packets.S2CPlayerPosition
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	b.Self.update(func(s *SelfState) {
		if p.Flags&1 != 0 {
			s.Position.X += float64(p.X)
		} else {
			s.Position.X = float64(p.X)
		}
		if p.Flags&2 != 0 {
			s.Position.Y += float64(p.Y)
		} else {
			s.Position.Y = float64(p.Y)
		}
		if p.Flags&4 != 0 {
			s.Position.Z += float64(p.Z)
		} else {
			s.Position.Z = float64(p.Z)
		}
		if p.Flags&8 != 0 {
			s.Rotation.Yaw += float32(p.Yaw)
		} else {
			s.Rotation.Yaw = float32(p.Yaw)
		}
		if p.Flags&16 != 0 {
			s.Rotation.Pitch += float32(p.Pitch)
		} else {
			s.Rotation.Pitch = float32(p.Pitch)
		}
		s.Velocity = Vec3{float64(p.VelocityX), float64(p.VelocityY), float64(p.VelocityZ)}
	})
	if err := b.client.WritePacket(&packets.C2SAcceptTeleportation{TeleportId: p.TeleportId}); err != nil {
		return err
	}
	b.respawning.Store(false)
	b.positionReady.Store(true)
	b.tryReady()
	return nil
}

func (b *Bot) handleChunk(w *jp.WirePacket) error {
	var p packets.S2CLevelChunkWithLight
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	col, err := chunks.ParseChunkColumn(int32(p.ChunkX), int32(p.ChunkZ), p.ChunkData, &p.LightData)
	if err != nil {
		return err
	}
	key := chunkKey{int32(p.ChunkX), int32(p.ChunkZ)}
	b.World.mu.Lock()
	b.World.chunks[key] = col
	for _, be := range col.BlockEntities {
		if data, ok := be.Data.(nbt.Compound); ok {
			b.World.blockEntities[BlockPos{int(key.X)*16 + be.X(), int(be.Y), int(key.Z)*16 + be.Z()}] = BlockEntity{int32(be.Type), data}
		}
	}
	b.World.mu.Unlock()
	b.World.onChunk.emit(key)
	b.tryReady()
	return nil
}
func (b *Bot) handleForgetChunk(w *jp.WirePacket) error {
	var p packets.S2CForgetLevelChunk
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	key := chunkKey{int32(p.ChunkX), int32(p.ChunkZ)}
	b.World.mu.Lock()
	delete(b.World.chunks, key)
	b.World.mu.Unlock()
	return nil
}
func (b *Bot) setBlock(pos BlockPos, id int32) {
	b.World.mu.Lock()
	col := b.World.chunks[chunkKey{int32(pos.X >> 4), int32(pos.Z >> 4)}]
	var oldID int32
	if col != nil {
		oldID = col.GetBlockState(pos.X, pos.Y, pos.Z)
		col.SetBlockState(pos.X, pos.Y, pos.Z, id)
	}
	b.World.mu.Unlock()
	if col != nil {
		old, _ := b.block(pos, oldID)
		now, _ := b.block(pos, id)
		b.World.onBlock.emit(BlockChange{pos, old, now})
	}
}
func (b *Bot) handleBlockUpdate(w *jp.WirePacket) error {
	var p packets.S2CBlockUpdate
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	b.setBlock(BlockPos{p.Location.X, p.Location.Y, p.Location.Z}, int32(p.BlockId))
	return nil
}
func (b *Bot) handleSectionUpdate(w *jp.WirePacket) error {
	var p packets.S2CSectionBlocksUpdate
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	sx, sy, sz := chunks.DecodeSectionPosition(int64(p.ChunkSectionPosition))
	for _, x := range p.Blocks {
		id, lx, ly, lz := chunks.DecodeBlockEntry(int64(x))
		b.setBlock(BlockPos{int(sx)*16 + lx, int(sy)*16 + ly, int(sz)*16 + lz}, id)
	}
	return nil
}

func angle(v ns.Angle) float32 { return float32(v) * 360 / 256 }
func (b *Bot) entityChange(kind string, e Entity) {
	b.Entities.onChange.emit(EntityChange{kind, cloneEntity(e)})
}
func (b *Bot) handleAddEntity(w *jp.WirePacket) error {
	var p packets.S2CAddEntity
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	e := Entity{ID: int32(p.EntityId), UUID: p.EntityUuid.String(), Type: dataentities.EntityTypeName(int32(p.Type)), Position: Vec3{float64(p.X), float64(p.Y), float64(p.Z)}, Velocity: Vec3{float64(p.Velocity.X), float64(p.Velocity.Y), float64(p.Velocity.Z)}, Rotation: Rotation{angle(p.Yaw), angle(p.Pitch)}, Metadata: map[int32]any{}}
	b.Entities.mu.Lock()
	b.Entities.values[e.ID] = e
	b.Entities.mu.Unlock()
	b.Players.linkEntity(e)
	b.entityChange("add", e)
	return nil
}
func (b *Bot) updateEntity(id int32, fn func(*Entity)) {
	b.Entities.mu.Lock()
	e, ok := b.Entities.values[id]
	if ok {
		fn(&e)
		b.Entities.values[id] = e
	}
	b.Entities.mu.Unlock()
	if ok {
		b.entityChange("update", e)
	}
}
func (b *Bot) handleMoveEntity(w *jp.WirePacket, rot bool) error {
	if rot {
		var p packets.S2CMoveEntityPosRot
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.updateEntity(int32(p.EntityId), func(e *Entity) {
			e.Position.X += float64(p.DeltaX) / 4096
			e.Position.Y += float64(p.DeltaY) / 4096
			e.Position.Z += float64(p.DeltaZ) / 4096
			e.Rotation = Rotation{angle(p.Yaw), angle(p.Pitch)}
		})
	} else {
		var p packets.S2CMoveEntityPos
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.updateEntity(int32(p.EntityId), func(e *Entity) {
			e.Position.X += float64(p.DeltaX) / 4096
			e.Position.Y += float64(p.DeltaY) / 4096
			e.Position.Z += float64(p.DeltaZ) / 4096
		})
	}
	return nil
}
func (b *Bot) handleRotateEntity(w *jp.WirePacket) error {
	var p packets.S2CMoveEntityRot
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	b.updateEntity(int32(p.EntityId), func(e *Entity) { e.Rotation = Rotation{angle(p.Yaw), angle(p.Pitch)} })
	return nil
}
func (b *Bot) handleTeleportEntity(w *jp.WirePacket) error {
	var p packets.S2CTeleportEntity
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	b.updateEntity(int32(p.EntityId), func(e *Entity) {
		e.Position = Vec3{float64(p.X), float64(p.Y), float64(p.Z)}
		e.Velocity = Vec3{float64(p.VelocityX), float64(p.VelocityY), float64(p.VelocityZ)}
		e.Rotation = Rotation{float32(p.Yaw), float32(p.Pitch)}
	})
	return nil
}
func (b *Bot) handleEntityMotion(w *jp.WirePacket) error {
	var p packets.S2CSetEntityMotion
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	b.updateEntity(int32(p.EntityId), func(e *Entity) {
		e.Velocity = Vec3{float64(p.Velocity.X), float64(p.Velocity.Y), float64(p.Velocity.Z)}
	})
	return nil
}
func (b *Bot) handleEntityData(w *jp.WirePacket) error {
	var p packets.S2CSetEntityData
	if err := w.ReadInto(&p); err != nil {
		// Entity metadata is observational state. A newly introduced serializer
		// or component codec must not tear down an otherwise healthy play
		// connection (and stop keepalives/navigation). Preserve the framed
		// stream, skip this one update, and retain enough detail to extend the
		// decoder later.
		b.cfg.Logger.Warn("ignored unsupported entity metadata", "error", err, "bytes", len(w.Data), "data", fmt.Sprintf("%x", w.Data))
		return nil
	}
	b.updateEntity(int32(p.EntityId), func(e *Entity) {
		if e.Metadata == nil {
			e.Metadata = map[int32]any{}
		}
		for _, m := range p.Metadata {
			e.Metadata[int32(m.Index)] = append([]byte(nil), m.Data...)
		}
	})
	return nil
}
func (b *Bot) handleRemoveEntities(w *jp.WirePacket) error {
	var p packets.S2CRemoveEntities
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	for _, id := range p.EntityIds {
		b.Entities.mu.Lock()
		e, ok := b.Entities.values[int32(id)]
		delete(b.Entities.values, int32(id))
		b.Entities.mu.Unlock()
		if ok {
			b.entityChange("remove", e)
		}
	}
	return nil
}

func (b *Bot) handleInventoryContent(w *jp.WirePacket) error {
	var p packets.S2CContainerSetContent
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	slots := make([]ItemStack, len(p.Slots))
	for i, s := range p.Slots {
		slots[i] = fromNetSlot(s)
	}
	b.Inventory.mu.Lock()
	if p.WindowId == 0 {
		b.Inventory.slots = slots
	}
	if int32(p.WindowId) == b.Inventory.window.ID {
		b.Inventory.window.ID = int32(p.WindowId)
		b.Inventory.window.StateID = int32(p.StateId)
		b.Inventory.window.Slots = slots
		b.Inventory.window.Carried = fromNetSlot(p.CarriedItem)
	}
	if int32(p.WindowId) == b.Inventory.window.ID && b.Inventory.window.Type == 12 && len(slots) >= 46 {
		for source := 10; source < 46; source++ {
			destination := source - 1
			if destination < len(b.Inventory.slots) {
				b.Inventory.slots[destination] = cloneStack(slots[source])
			}
		}
	}
	snapshot := cloneWindow(b.Inventory.window)
	b.Inventory.mu.Unlock()
	b.Inventory.onWindow.emit(WindowChange{Window: snapshot})
	return nil
}
func (b *Bot) handleInventorySlot(w *jp.WirePacket) error {
	var p packets.S2CContainerSetSlot
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	idx := int(p.Slot)
	b.Inventory.mu.Lock()
	if p.WindowId == 0 && idx >= 0 && idx < len(b.Inventory.slots) {
		b.Inventory.slots[idx] = fromNetSlot(p.SlotData)
	}
	if int32(p.WindowId) == b.Inventory.window.ID {
		b.Inventory.window.StateID = int32(p.StateId)
		if idx >= 0 && idx < len(b.Inventory.window.Slots) {
			b.Inventory.window.Slots[idx] = fromNetSlot(p.SlotData)
		}
		if idx == -1 {
			b.Inventory.window.Carried = fromNetSlot(p.SlotData)
		}
		if b.Inventory.window.Type == 12 && idx >= 10 && idx < 46 {
			destination := idx - 1
			if destination < len(b.Inventory.slots) {
				b.Inventory.slots[destination] = fromNetSlot(p.SlotData)
			}
		}
	}
	if int32(p.WindowId) == -1 && idx == -1 {
		b.Inventory.window.Carried = fromNetSlot(p.SlotData)
	}
	snapshot := cloneWindow(b.Inventory.window)
	b.Inventory.mu.Unlock()
	b.Inventory.onWindow.emit(WindowChange{Window: snapshot})
	return nil
}
