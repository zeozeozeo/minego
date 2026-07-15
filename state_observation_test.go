package minego

import (
	"testing"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packet_ids"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	jP "github.com/zeozeozeo/minego/internal/protocol/java_protocol"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

func observationWire(t *testing.T, b *Bot, id int32, write func(*ns.PacketBuffer) error) {
	t.Helper()
	buf := ns.NewWriter()
	if err := write(buf); err != nil {
		t.Fatal(err)
	}
	if err := b.handlePlay(&jP.WirePacket{PacketID: ns.VarInt(id), Data: append([]byte(nil), buf.Bytes()...)}); err != nil {
		t.Fatal(err)
	}
}

func TestObservationServicesAndWeatherTabList(t *testing.T) {
	b := actionTestBot(t)
	if b.Weather == nil || b.TabList == nil || b.Scoreboards == nil || b.Teams == nil || b.BossBars == nil || b.Effects == nil || b.Server == nil {
		t.Fatal("an observation service was not initialized")
	}
	if !b.Supports(FeatureStateObservation) {
		t.Fatal("state observation feature is not advertised")
	}
	handleTestPacket(t, b, &packets.S2CGameEvent{Event: 2})
	handleTestPacket(t, b, &packets.S2CGameEvent{Event: 7, Value: .75})
	if s := b.Weather.State(); !s.Raining || s.RainLevel != .75 {
		t.Fatalf("weather = %#v", s)
	}
	handleTestPacket(t, b, &packets.S2CTabList{Header: ns.TextComponent{Text: "Header"}, Footer: ns.TextComponent{Text: "Footer"}})
	if s := b.TabList.State(); s.Header != "Header" || s.Footer != "Footer" {
		t.Fatalf("tab list = %#v", s)
	}
}

func TestScoreboardTeamAndBossBarReducers(t *testing.T) {
	b := actionTestBot(t)
	observationWire(t, b, packet_ids.S2CSetObjectiveID, func(w *ns.PacketBuffer) error {
		if err := w.WriteString("kills"); err != nil {
			return err
		}
		if err := w.WriteInt8(0); err != nil {
			return err
		}
		if err := w.WriteTextComponent(ns.TextComponent{Text: "Kills"}); err != nil {
			return err
		}
		if err := w.WriteVarInt(0); err != nil {
			return err
		}
		return w.WriteBool(false)
	})
	observationWire(t, b, packet_ids.S2CSetScoreID, func(w *ns.PacketBuffer) error {
		if err := w.WriteString("Steve"); err != nil {
			return err
		}
		if err := w.WriteString("kills"); err != nil {
			return err
		}
		if err := w.WriteVarInt(7); err != nil {
			return err
		}
		if err := w.WriteBool(true); err != nil {
			return err
		}
		if err := w.WriteTextComponent(ns.TextComponent{Text: "Seven"}); err != nil {
			return err
		}
		return w.WriteBool(false)
	})
	s := b.Scoreboards.Snapshot()
	if s.Objectives["kills"].DisplayName != "Kills" || s.Scores["kills"]["Steve"].Value != 7 || s.Scores["kills"]["Steve"].DisplayName != "Seven" {
		t.Fatalf("scoreboard = %#v", s)
	}
	o := s.Objectives["kills"]
	o.Data[0] = 1
	if b.Scoreboards.Snapshot().Objectives["kills"].Data[0] != 0 {
		t.Fatal("objective data shares mutable storage")
	}
	observationWire(t, b, packet_ids.S2CSetPlayerTeamID, func(w *ns.PacketBuffer) error {
		if err := w.WriteString("red"); err != nil {
			return err
		}
		if err := w.WriteInt8(0); err != nil {
			return err
		}
		if err := w.WriteTextComponent(ns.TextComponent{Text: "Red"}); err != nil {
			return err
		}
		if err := w.WriteUint8(3); err != nil {
			return err
		}
		if err := w.WriteString("always"); err != nil {
			return err
		}
		if err := w.WriteString("always"); err != nil {
			return err
		}
		if err := w.WriteVarInt(12); err != nil {
			return err
		}
		if err := w.WriteTextComponent(ns.TextComponent{Text: "[R]"}); err != nil {
			return err
		}
		if err := w.WriteTextComponent(ns.TextComponent{}); err != nil {
			return err
		}
		if err := w.WriteVarInt(1); err != nil {
			return err
		}
		return w.WriteString("Steve")
	})
	team, ok := b.Teams.Get("red")
	if !ok || team.DisplayName != "Red" || len(team.Entities) != 1 || team.Entities[0] != "Steve" {
		t.Fatalf("team = %#v, %v", team, ok)
	}
	team.Entities[0] = "mutated"
	again, _ := b.Teams.Get("red")
	if again.Entities[0] != "Steve" {
		t.Fatal("team entities share mutable storage")
	}
	data := ns.NewWriter()
	_ = data.WriteTextComponent(ns.TextComponent{Text: "Dragon"})
	_ = data.WriteFloat32(.5)
	_ = data.WriteVarInt(2)
	_ = data.WriteVarInt(1)
	_ = data.WriteUint8(5)
	handleTestPacket(t, b, &packets.S2CBossEvent{Action: packets.BossEventActionAdd, Data: data.Bytes()})
	bars := b.BossBars.All()
	if len(bars) != 1 || bars[0].Title != "Dragon" || bars[0].Health != .5 || bars[0].Flags != 5 {
		t.Fatalf("boss bars = %#v", bars)
	}
}

func TestWorldEffectsAndServerSnapshots(t *testing.T) {
	b := actionTestBot(t)
	sounds := make(chan SoundEvent, 1)
	particles := make(chan ParticleEvent, 1)
	explosions := make(chan ExplosionEvent, 1)
	b.Effects.OnSound(func(v SoundEvent) { sounds <- v })
	b.Effects.OnParticle(func(v ParticleEvent) { particles <- v })
	b.Effects.OnExplosion(func(v ExplosionEvent) { explosions <- v })
	observationWire(t, b, packet_ids.S2CSoundID, func(w *ns.PacketBuffer) error {
		if err := w.WriteVarInt(5); err != nil {
			return err
		}
		if err := w.WriteVarInt(2); err != nil {
			return err
		}
		if err := w.WriteInt32(80); err != nil {
			return err
		}
		if err := w.WriteInt32(160); err != nil {
			return err
		}
		if err := w.WriteInt32(-40); err != nil {
			return err
		}
		if err := w.WriteFloat32(1); err != nil {
			return err
		}
		if err := w.WriteFloat32(.5); err != nil {
			return err
		}
		return w.WriteInt64(9)
	})
	sound := <-sounds
	if sound.ID != 4 || sound.Position != (Vec3{10, 20, -5}) || sound.Seed != 9 {
		t.Fatalf("sound = %#v", sound)
	}
	observationWire(t, b, packet_ids.S2CLevelParticlesID, func(w *ns.PacketBuffer) error {
		_ = w.WriteBool(true)
		_ = w.WriteBool(false)
		_ = w.WriteFloat64(1)
		_ = w.WriteFloat64(2)
		_ = w.WriteFloat64(3)
		_ = w.WriteFloat32(.1)
		_ = w.WriteFloat32(.2)
		_ = w.WriteFloat32(.3)
		_ = w.WriteFloat32(.4)
		_ = w.WriteInt32(6)
		_ = w.WriteVarInt(31)
		return w.WriteFixedByteArray([]byte{8, 9})
	})
	particle := <-particles
	if particle.ID != 31 || particle.Count != 6 || len(particle.Data) != 2 {
		t.Fatalf("particle = %#v", particle)
	}
	observationWire(t, b, packet_ids.S2CExplodeID, func(w *ns.PacketBuffer) error {
		_ = w.WriteFloat64(4)
		_ = w.WriteFloat64(5)
		_ = w.WriteFloat64(6)
		return w.WriteFixedByteArray([]byte{1, 2, 3})
	})
	explosion := <-explosions
	if explosion.Position != (Vec3{4, 5, 6}) || len(explosion.Data) != 3 {
		t.Fatalf("explosion = %#v", explosion)
	}
	handleTestPacket(t, b, &packets.S2CChangeDifficulty{Difficulty: 3, DifficultyLocked: true})
	handleTestPacket(t, b, &packets.S2CGameRuleValues{Values: []packets.GameRuleEntry{{Key: "minecraft:keep_inventory", Value: "true"}}})
	handleTestPacket(t, b, &packets.S2CServerData{Motd: ns.TextComponent{Text: "MineGo"}, Icon: ns.PrefixedOptional[ns.ByteArray]{Present: true, Value: []byte{1, 2}}})
	server := b.Server.State()
	if server.Difficulty != 3 || !server.DifficultyLocked || server.MOTD != "MineGo" || server.GameRules["minecraft:keep_inventory"] != "true" {
		t.Fatalf("server = %#v", server)
	}
	server.Icon[0] = 9
	server.GameRules["minecraft:keep_inventory"] = "false"
	again := b.Server.State()
	if again.Icon[0] != 1 || again.GameRules["minecraft:keep_inventory"] != "true" {
		t.Fatal("server snapshot shares mutable state")
	}
}

func TestLegacyObservationPacketIDs(t *testing.T) {
	b, err := New(Config{Address: "localhost", Version: "1.21.11"})
	if err != nil {
		t.Fatal(err)
	}
	observationWire(t, b, 120, func(w *ns.PacketBuffer) error {
		if err := w.WriteTextComponent(ns.TextComponent{Text: "old"}); err != nil {
			return err
		}
		return w.WriteTextComponent(ns.TextComponent{Text: "footer"})
	})
	if b.TabList.State().Header != "old" {
		t.Fatalf("legacy tab = %#v", b.TabList.State())
	}
}
