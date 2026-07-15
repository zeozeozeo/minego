package minego

import (
	"context"
	"fmt"
	"time"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

type particleEvent struct {
	Position Vec3
	ID       int32
}

type SpecialInteractions struct {
	bot        *Bot
	onParticle event[particleEvent]
}

func newSpecialInteractions(bot *Bot) *SpecialInteractions { return &SpecialInteractions{bot: bot} }

func (s *SpecialInteractions) Sleep(ctx context.Context, bed BlockPos) error {
	return s.bot.Interaction.ActivateBlock(ctx, BlockInteraction{Position: bed, Face: 1, Cursor: Vec3{.5, .5, .5}, Hand: MainHand})
}

func (s *SpecialInteractions) Wake(ctx context.Context) error {
	state := s.bot.Self.State()
	return s.bot.send(ctx, &packets.C2SPlayerCommand{EntityId: ns.VarInt(state.EntityID), ActionId: 2})
}

func (s *SpecialInteractions) CastFishingRod(ctx context.Context, hand Hand) error {
	return s.bot.Interaction.UseItem(ctx, hand)
}
func (s *SpecialInteractions) ReelFishingRod(ctx context.Context, hand Hand) error {
	return s.bot.Interaction.UseItem(ctx, hand)
}

// Fish casts, discovers the spawned bobber, waits for the fishing/splash
// particle at the hook, and reels in. timeout bounds the complete operation.
func (s *SpecialInteractions) Fish(ctx context.Context, timeout time.Duration, hand Hand) error {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	bobbers := make(chan Entity, 1)
	unsubEntity := s.bot.Entities.OnChange(func(change EntityChange) {
		if change.Kind == "add" && change.Entity.Type == "minecraft:fishing_bobber" {
			select {
			case bobbers <- change.Entity:
			default:
			}
		}
	})
	defer unsubEntity()
	if err := s.CastFishingRod(ctx, hand); err != nil {
		return err
	}
	var bobber Entity
	for _, entity := range s.bot.Entities.All() {
		if entity.Type == "minecraft:fishing_bobber" && entity.Position.Distance(s.bot.Self.State().Position) < 16 {
			bobber = entity
			break
		}
	}
	if bobber.ID == 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.bot.done:
			return ErrNotConnected
		case bobber = <-bobbers:
		}
	}
	bites := make(chan struct{}, 1)
	unsubParticle := s.onParticle.subscribe(func(p particleEvent) {
		current, ok := s.bot.Entities.Get(bobber.ID)
		if ok {
			bobber = current
		}
		if s.isFishingParticle(p.ID) && p.Position.Distance(bobber.Position) <= 1.5 {
			select {
			case bites <- struct{}{}:
			default:
			}
		}
	})
	defer unsubParticle()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.bot.done:
		return ErrNotConnected
	case <-bites:
		return s.ReelFishingRod(ctx, hand)
	}
}

func (s *SpecialInteractions) isFishingParticle(id int32) bool {
	switch s.bot.Version().Protocol {
	case 774:
		return id == 31 || id == 68
	case 775:
		return id == 31 || id == 70
	default:
		return id == 38 || id == 77
	}
}

// EditBook updates a writable book in a hotbar slot. A nonempty title signs it.
func (s *SpecialInteractions) EditBook(ctx context.Context, hotbar int, pages []string, title string) error {
	if hotbar < 0 || hotbar > 8 {
		return ErrInvalidSlot
	}
	if len(pages) > 100 {
		return fmt.Errorf("minego: a book cannot contain more than 100 pages")
	}
	entries := make([]ns.String, len(pages))
	for index, page := range pages {
		if len(page) > 8192 {
			return fmt.Errorf("minego: book page %d exceeds 8192 bytes", index)
		}
		entries[index] = ns.String(page)
	}
	p := &packets.C2SEditBook{Slot: ns.VarInt(hotbar), Entries: entries, Title: ns.None[ns.String]()}
	if title != "" {
		if len(title) > 128 {
			return fmt.Errorf("minego: book title exceeds 128 bytes")
		}
		p.Title = ns.Some(ns.String(title))
	}
	return s.bot.send(ctx, p)
}

func (s *SpecialInteractions) UpdateSign(ctx context.Context, pos BlockPos, front bool, lines [4]string) error {
	for _, line := range lines {
		if len(line) > 384 {
			return fmt.Errorf("minego: sign line exceeds 384 bytes")
		}
	}
	return s.bot.send(ctx, &packets.C2SSignUpdate{Location: ns.NewPosition(pos.X, pos.Y, pos.Z), IsFrontText: ns.Boolean(front), Line1: ns.String(lines[0]), Line2: ns.String(lines[1]), Line3: ns.String(lines[2]), Line4: ns.String(lines[3])})
}
