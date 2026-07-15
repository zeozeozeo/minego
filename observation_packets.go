package minego

import (
	"fmt"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/lang"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	jP "github.com/zeozeozeo/minego/internal/protocol/java_protocol"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

func renderText(buf *ns.PacketBuffer) (string, error) {
	v, err := buf.ReadTextComponent()
	if err != nil {
		return "", err
	}
	return v.Render(lang.Translate), nil
}

// 1.21.11 predates two play packets added in 26.1. Its later packet IDs are
// therefore shifted even though the normalized payloads below are unchanged.
func (b *Bot) handleLegacyObservation(w *jP.WirePacket) (bool, error) {
	if b.Version().Protocol != 774 {
		return false, nil
	}
	switch w.PacketID {
	case 46:
		return true, b.handleParticle(w.Data)
	case 77:
		r := ns.NewReader(w.Data)
		entity, err := r.ReadString(32767)
		if err != nil {
			return true, err
		}
		objective := ""
		present, err := r.ReadBool()
		if err != nil {
			return true, err
		}
		if present {
			v, e := r.ReadString(32767)
			if e != nil {
				return true, e
			}
			objective = string(v)
		}
		b.Scoreboards.resetScore(string(entity), objective)
		return true, nil
	case 84:
		r := ns.NewReader(w.Data)
		motd, err := renderText(r)
		if err != nil {
			return true, err
		}
		present, err := r.ReadBool()
		if err != nil {
			return true, err
		}
		var icon []byte
		if present {
			v, e := r.ReadByteArray(1 << 20)
			if e != nil {
				return true, e
			}
			icon = append([]byte(nil), v...)
		}
		b.Server.update(func(s *ServerState) { s.MOTD = motd; s.Icon = icon })
		return true, nil
	case 93:
		r := ns.NewReader(w.Data)
		distance, err := r.ReadVarInt()
		if err != nil {
			return true, err
		}
		b.World.mu.Lock()
		b.World.viewDistance = int32(distance)
		b.World.mu.Unlock()
		b.Server.update(func(s *ServerState) { s.ViewDistance = int32(distance) })
		return true, nil
	case 96:
		r := ns.NewReader(w.Data)
		slot, err := r.ReadVarInt()
		if err != nil {
			return true, err
		}
		name, err := r.ReadString(32767)
		if err != nil {
			return true, err
		}
		b.Scoreboards.setDisplay(int32(slot), string(name))
		return true, nil
	case 104:
		return true, b.handleObjective(w.Data)
	case 107:
		return true, b.handleTeam(w.Data)
	case 108:
		return true, b.handleScore(w.Data)
	case 109:
		r := ns.NewReader(w.Data)
		distance, err := r.ReadVarInt()
		if err != nil {
			return true, err
		}
		b.Server.update(func(s *ServerState) { s.SimulationDistance = int32(distance) })
		return true, nil
	case 114:
		return true, b.handleSound(w.Data, true)
	case 115:
		return true, b.handleSound(w.Data, false)
	case 120:
		r := ns.NewReader(w.Data)
		header, err := renderText(r)
		if err != nil {
			return true, err
		}
		footer, err := renderText(r)
		if err != nil {
			return true, err
		}
		b.TabList.update(TabListState{Header: header, Footer: footer})
		return true, nil
	}
	return false, nil
}

func (b *Bot) handleGameEvent(w *jP.WirePacket) error {
	var p packets.S2CGameEvent
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	switch uint8(p.Event) {
	case 1:
		b.Weather.update(func(s *WeatherState) { s.Raining = false })
	case 2:
		b.Weather.update(func(s *WeatherState) { s.Raining = true })
	case 3:
		b.Self.update(func(s *SelfState) { s.GameMode = uint8(p.Value) })
	case 7:
		b.Weather.update(func(s *WeatherState) { s.RainLevel = float32(p.Value); s.Raining = s.RainLevel > 0 })
	case 8:
		b.Weather.update(func(s *WeatherState) { s.ThunderLevel = float32(p.Value) })
	}
	return nil
}

func (b *Bot) handleObjective(data []byte) error {
	r := ns.NewReader(data)
	name, err := r.ReadString(32767)
	if err != nil {
		return err
	}
	mode, err := r.ReadInt8()
	if err != nil {
		return err
	}
	if mode == 1 {
		b.Scoreboards.removeObjective(string(name))
		return nil
	}
	display, err := renderText(r)
	if err != nil {
		return err
	}
	render, err := r.ReadVarInt()
	if err != nil {
		return err
	}
	tail, err := r.ReadRemaining()
	if err != nil {
		return err
	}
	b.Scoreboards.setObjective(Objective{Name: string(name), DisplayName: display, RenderType: int32(render), Data: tail})
	return nil
}
func (b *Bot) handleDisplayObjective(w *jP.WirePacket) error {
	var p packets.S2CSetDisplayObjective
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	b.Scoreboards.setDisplay(int32(p.Position), string(p.ScoreName))
	return nil
}
func (b *Bot) handleScore(data []byte) error {
	r := ns.NewReader(data)
	entity, err := r.ReadString(32767)
	if err != nil {
		return err
	}
	objective, err := r.ReadString(32767)
	if err != nil {
		return err
	}
	value, err := r.ReadVarInt()
	if err != nil {
		return err
	}
	display := ""
	present, err := r.ReadBool()
	if err != nil {
		return err
	}
	if present {
		display, err = renderText(r)
		if err != nil {
			return err
		}
	}
	tail, err := r.ReadRemaining()
	if err != nil {
		return err
	}
	b.Scoreboards.setScore(Score{Entity: string(entity), Objective: string(objective), DisplayName: display, Value: int32(value), Data: tail})
	return nil
}
func (b *Bot) handleResetScore(w *jP.WirePacket) error {
	var p packets.S2CResetScore
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	objective := ""
	if p.ObjectiveName.Present {
		objective = string(p.ObjectiveName.Value)
	}
	b.Scoreboards.resetScore(string(p.EntityName), objective)
	return nil
}

func readStrings(r *ns.PacketBuffer) ([]string, error) {
	n, err := r.ReadVarInt()
	if err != nil {
		return nil, err
	}
	if n < 0 || n > 100000 {
		return nil, fmt.Errorf("invalid string list length %d", n)
	}
	out := make([]string, int(n))
	for i := range out {
		v, e := r.ReadString(32767)
		if e != nil {
			return nil, e
		}
		out[i] = string(v)
	}
	return out, nil
}
func (b *Bot) handleTeam(data []byte) error {
	r := ns.NewReader(data)
	name, err := r.ReadString(32767)
	if err != nil {
		return err
	}
	method, err := r.ReadInt8()
	if err != nil {
		return err
	}
	key := string(name)
	if method == 1 {
		b.Teams.remove(key)
		return nil
	}
	v, _ := b.Teams.Get(key)
	v.Name = key
	if method == 0 || method == 2 {
		if v.DisplayName, err = renderText(r); err != nil {
			return err
		}
		flags, e := r.ReadUint8()
		if e != nil {
			return e
		}
		v.FriendlyFlags = uint8(flags)
		visibility, e := r.ReadString(32767)
		if e != nil {
			return e
		}
		v.NameTagVisibility = string(visibility)
		collision, e := r.ReadString(32767)
		if e != nil {
			return e
		}
		v.CollisionRule = string(collision)
		color, e := r.ReadVarInt()
		if e != nil {
			return e
		}
		v.Color = int32(color)
		if v.Prefix, err = renderText(r); err != nil {
			return err
		}
		if v.Suffix, err = renderText(r); err != nil {
			return err
		}
	}
	if method == 0 || method == 3 || method == 4 {
		members, e := readStrings(r)
		if e != nil {
			return e
		}
		set := map[string]bool{}
		for _, x := range v.Entities {
			set[x] = true
		}
		if method == 4 {
			for _, x := range members {
				delete(set, x)
			}
		} else {
			for _, x := range members {
				set[x] = true
			}
		}
		v.Entities = v.Entities[:0]
		for x := range set {
			v.Entities = append(v.Entities, x)
		}
	}
	kind := "update"
	if method == 0 {
		kind = "add"
	}
	b.Teams.put(kind, v)
	return nil
}

func (b *Bot) handleBossBar(w *jP.WirePacket) error {
	var p packets.S2CBossEvent
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	id := p.Uuid.String()
	switch p.Action {
	case packets.BossEventActionAdd:
		d, e := p.DataActionAdd()
		if e != nil {
			return e
		}
		b.BossBars.update(id, "add", func(v *BossBar) {
			v.Title = d.Title.Render(lang.Translate)
			v.Health = float32(d.Health)
			v.Color = int32(d.Color)
			v.Division = int32(d.Division)
			v.Flags = uint8(d.Flags)
		})
	case packets.BossEventActionRemove:
		b.BossBars.update(id, "remove", func(*BossBar) {})
	case packets.BossEventActionUpdateHealth:
		d, e := p.DataActionUpdateHealth()
		if e != nil {
			return e
		}
		b.BossBars.update(id, "update", func(v *BossBar) { v.Health = float32(d.Health) })
	case packets.BossEventActionUpdateTitle:
		d, e := p.DataActionUpdateTitle()
		if e != nil {
			return e
		}
		b.BossBars.update(id, "update", func(v *BossBar) { v.Title = d.Title.Render(lang.Translate) })
	case packets.BossEventActionUpdateStyle:
		d, e := p.DataActionUpdateStyle()
		if e != nil {
			return e
		}
		b.BossBars.update(id, "update", func(v *BossBar) { v.Color = int32(d.Color); v.Division = int32(d.Division) })
	case packets.BossEventActionUpdateFlags:
		d, e := p.DataActionUpdateFlags()
		if e != nil {
			return e
		}
		b.BossBars.update(id, "update", func(v *BossBar) { v.Flags = uint8(d.Flags) })
	default:
		return fmt.Errorf("unknown boss bar action %d", p.Action)
	}
	return nil
}

func readSoundHolder(r *ns.PacketBuffer) (id int32, name string, fixed *float32, err error) {
	raw, e := r.ReadVarInt()
	if e != nil {
		err = e
		return
	}
	id = int32(raw) - 1
	if raw == 0 {
		v, e := r.ReadIdentifier()
		if e != nil {
			err = e
			return
		}
		name = string(v)
		present, e := r.ReadBool()
		if e != nil {
			err = e
			return
		}
		if present {
			x, e := r.ReadFloat32()
			if e != nil {
				err = e
				return
			}
			f := float32(x)
			fixed = &f
		}
	}
	return
}
func (b *Bot) handleSound(data []byte, entity bool) error {
	r := ns.NewReader(data)
	id, name, fixed, err := readSoundHolder(r)
	if err != nil {
		return err
	}
	cat, err := r.ReadVarInt()
	if err != nil {
		return err
	}
	v := SoundEvent{ID: id, Name: name, Category: int32(cat), FixedRange: fixed}
	if entity {
		eid, e := r.ReadVarInt()
		if e != nil {
			return e
		}
		v.EntityID = int32(eid)
		if x, ok := b.Entities.Get(v.EntityID); ok {
			v.Position = x.Position
		}
	} else {
		x, e := r.ReadInt32()
		if e != nil {
			return e
		}
		y, e := r.ReadInt32()
		if e != nil {
			return e
		}
		z, e := r.ReadInt32()
		if e != nil {
			return e
		}
		v.Position = Vec3{float64(x) / 8, float64(y) / 8, float64(z) / 8}
	}
	volume, e := r.ReadFloat32()
	if e != nil {
		return e
	}
	pitch, e := r.ReadFloat32()
	if e != nil {
		return e
	}
	seed, e := r.ReadInt64()
	if e != nil {
		return e
	}
	v.Volume = float32(volume)
	v.Pitch = float32(pitch)
	v.Seed = int64(seed)
	b.Effects.onSound.emit(v)
	return nil
}
func (b *Bot) handleExplosion(data []byte) error {
	r := ns.NewReader(data)
	x, e := r.ReadFloat64()
	if e != nil {
		return e
	}
	y, e := r.ReadFloat64()
	if e != nil {
		return e
	}
	z, e := r.ReadFloat64()
	if e != nil {
		return e
	}
	tail, e := r.ReadRemaining()
	if e != nil {
		return e
	}
	b.Effects.onExplosion.emit(ExplosionEvent{Position: Vec3{float64(x), float64(y), float64(z)}, Data: append([]byte(nil), tail...)})
	return nil
}
func (b *Bot) handleParticle(data []byte) error {
	r := ns.NewReader(data)
	long, e := r.ReadBool()
	if e != nil {
		return e
	}
	always, e := r.ReadBool()
	if e != nil {
		return e
	}
	x, e := r.ReadFloat64()
	if e != nil {
		return e
	}
	y, e := r.ReadFloat64()
	if e != nil {
		return e
	}
	z, e := r.ReadFloat64()
	if e != nil {
		return e
	}
	ox, e := r.ReadFloat32()
	if e != nil {
		return e
	}
	oy, e := r.ReadFloat32()
	if e != nil {
		return e
	}
	oz, e := r.ReadFloat32()
	if e != nil {
		return e
	}
	speed, e := r.ReadFloat32()
	if e != nil {
		return e
	}
	count, e := r.ReadInt32()
	if e != nil {
		return e
	}
	id, e := r.ReadVarInt()
	if e != nil {
		return e
	}
	tail, e := r.ReadRemaining()
	if e != nil {
		return e
	}
	pos := Vec3{float64(x), float64(y), float64(z)}
	b.Special.onParticle.emit(particleEvent{Position: pos, ID: int32(id)})
	b.Effects.onParticle.emit(ParticleEvent{ID: int32(id), Count: int32(count), Position: pos, Offset: Vec3{float64(ox), float64(oy), float64(oz)}, MaxSpeed: float32(speed), LongDistance: bool(long), AlwaysVisible: bool(always), Data: append([]byte(nil), tail...)})
	return nil
}
