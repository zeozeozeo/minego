package minego

import (
	"context"
	"errors"
	"testing"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	jp "github.com/zeozeozeo/minego/internal/protocol/java_protocol"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

func actionTestBot(t *testing.T) *Bot {
	t.Helper()
	b, err := New(Config{Address: "localhost", Version: "26.2"})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func handleTestPacket(t *testing.T, b *Bot, p jp.Packet) {
	t.Helper()
	w, err := jp.ToWire(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.handlePlay(w); err != nil {
		t.Fatal(err)
	}
}

func TestActionServicesAndCapabilities(t *testing.T) {
	b := actionTestBot(t)
	if b.Interaction == nil || b.Combat == nil || b.Containers == nil || b.Special == nil || b.Riding == nil {
		t.Fatal("an action service was not initialized")
	}
	for _, feature := range []Feature{FeatureInteractions, FeatureCombat, FeatureContainers, FeatureRiding} {
		if !b.Supports(feature) {
			t.Fatalf("26.2 does not advertise %s", feature)
		}
	}
}

func TestInteractionAndCombatValidation(t *testing.T) {
	b := actionTestBot(t)
	if err := b.Interaction.Swing(context.Background(), Hand(9)); err == nil {
		t.Fatal("invalid hand accepted")
	}
	if err := b.Interaction.ActivateBlock(context.Background(), BlockInteraction{Face: 9}); err == nil {
		t.Fatal("invalid face accepted")
	}
	b.Self.update(func(s *SelfState) { s.Position = Vec3{} })
	b.Entities.mu.Lock()
	b.Entities.values[7] = Entity{ID: 7, Position: Vec3{X: 20}}
	b.Entities.mu.Unlock()
	if err := b.Combat.Attack(context.Background(), 7, AttackOptions{}); !errors.Is(err, ErrOutOfReach) {
		t.Fatalf("attack error = %v", err)
	}
}

func TestWindowPropertiesOffersAndDeepCopy(t *testing.T) {
	b := actionTestBot(t)
	b.Inventory.mu.Lock()
	b.Inventory.window = WindowSnapshot{ID: 4, Properties: map[int16]int16{0: 1}}
	b.Inventory.mu.Unlock()
	handleTestPacket(t, b, &packets.S2CContainerSetData{WindowId: 4, Property: 2, Value: 30})
	handleTestPacket(t, b, &packets.S2CMerchantOffers{WindowId: 4, Data: []byte{1, 2, 3}})
	w := b.Inventory.Window()
	if w.Properties[2] != 30 || len(w.Offers) != 3 {
		t.Fatalf("window = %#v", w)
	}
	w.Properties[2] = 99
	w.Offers[0] = 9
	again := b.Inventory.Window()
	if again.Properties[2] != 30 || again.Offers[0] != 1 {
		t.Fatal("window maps or offers share mutable state")
	}
	c := &Container{service: b.Containers, ID: 4}
	state := c.Furnace()
	if state.BurnTime != 1 || state.CookTime != 30 {
		t.Fatalf("furnace = %#v", state)
	}
}

func TestAbilitiesPassengersAndDamageEvents(t *testing.T) {
	b := actionTestBot(t)
	b.Self.update(func(s *SelfState) { s.EntityID = 42 })
	handleTestPacket(t, b, &packets.S2CPlayerAbilities{Flags: 6, FlyingSpeed: .15})
	state := b.Self.State()
	if !state.Flying || !state.CanFly || state.FlyingSpeed != .15 {
		t.Fatalf("abilities = %#v", state)
	}

	payload := ns.NewWriter()
	if err := payload.WriteVarInt(8); err != nil {
		t.Fatal(err)
	}
	if err := payload.WriteVarInt(2); err != nil {
		t.Fatal(err)
	}
	if err := payload.WriteVarInt(12); err != nil {
		t.Fatal(err)
	}
	if err := payload.WriteVarInt(42); err != nil {
		t.Fatal(err)
	}
	if err := b.handleSetPassengers(payload.Bytes()); err != nil {
		t.Fatal(err)
	}
	if b.Self.State().VehicleID != 8 {
		t.Fatalf("vehicle = %d", b.Self.State().VehicleID)
	}

	events := make(chan CombatEvent, 1)
	unsub := b.Combat.OnEvent(func(e CombatEvent) { events <- e })
	defer unsub()
	handleTestPacket(t, b, &packets.S2CDamageEvent{EntityId: 42, SourceCauseId: 9})
	e := <-events
	if e.Kind != "damage" || e.EntityID != 42 || e.SourceID != 9 {
		t.Fatalf("combat event = %#v", e)
	}
}

func TestSpecialAndInventoryValidation(t *testing.T) {
	b := actionTestBot(t)
	if err := b.Special.EditBook(context.Background(), -1, nil, ""); !errors.Is(err, ErrInvalidSlot) {
		t.Fatalf("book error = %v", err)
	}
	if err := b.Special.EditBook(context.Background(), 0, make([]string, 101), ""); err == nil {
		t.Fatal("oversized book accepted")
	}
	if err := b.Inventory.CreativeSet(context.Background(), 1, ItemStack{}); !errors.Is(err, ErrInvalidGameMode) {
		t.Fatalf("creative error = %v", err)
	}
	if _, err := b.Inventory.Click(context.Background(), ClickOptions{Mode: ClickMode(20)}); err == nil {
		t.Fatal("invalid click mode accepted")
	}
}

func TestFishingParticleIDsFollowVersionPack(t *testing.T) {
	for _, tc := range []struct {
		version         string
		fishing, splash int32
	}{{"1.21.11", 31, 68}, {"26.1", 31, 70}, {"26.2", 38, 77}} {
		b, err := New(Config{Address: "localhost", Version: tc.version})
		if err != nil {
			t.Fatal(err)
		}
		if !b.Special.isFishingParticle(tc.fishing) || !b.Special.isFishingParticle(tc.splash) {
			t.Fatalf("%s fishing particle IDs rejected", tc.version)
		}
	}
}
