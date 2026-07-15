package minego

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// Hand identifies the hand used by an interaction packet.
type Hand int32

const (
	MainHand Hand = iota
	OffHand
)

type BlockInteraction struct {
	Position BlockPos
	Face     int
	Cursor   Vec3
	Hand     Hand
	Sneaking bool
	Swing    bool
}

type EntityInteraction struct {
	EntityID int32
	Hand     Hand
	Location Vec3 // relative hit position; zero activates the entity body
	Sneaking bool
	Swing    bool
	Reach    float64
}

// Interaction provides the protocol's stable item, block, and entity actions.
type Interaction struct{ bot *Bot }

func newInteraction(bot *Bot) *Interaction { return &Interaction{bot: bot} }

func (i *Interaction) Swing(ctx context.Context, hand Hand) error {
	if err := validHand(hand); err != nil {
		return err
	}
	return i.bot.send(ctx, &packets.C2SSwing{Hand: ns.VarInt(hand)})
}

func (i *Interaction) UseItem(ctx context.Context, hand Hand) error {
	if err := validHand(hand); err != nil {
		return err
	}
	lease, err := i.bot.actions.acquire(ctx, controlHands, priorityExplicit)
	if err != nil {
		return err
	}
	defer lease.Release()
	state := i.bot.Self.State()
	seq := i.bot.nextSequence()
	return i.bot.send(lease.Context(ctx), &packets.C2SUseItem{Hand: ns.VarInt(hand), Sequence: ns.VarInt(seq), Yaw: ns.Float32(state.Rotation.Yaw), Pitch: ns.Float32(state.Rotation.Pitch)})
}

// ReleaseItem stops an in-progress use such as eating, drawing a bow, or
// charging a trident.
func (i *Interaction) ReleaseItem(ctx context.Context) error {
	seq := i.bot.nextSequence()
	return i.bot.send(ctx, &packets.C2SPlayerAction{Status: 5, Location: ns.NewPosition(0, 0, 0), Face: 0, Sequence: ns.VarInt(seq)})
}

func (i *Interaction) ActivateBlock(ctx context.Context, action BlockInteraction) error {
	if err := validHand(action.Hand); err != nil {
		return err
	}
	if action.Face < 0 || action.Face > 5 {
		return fmt.Errorf("minego: block face must be between 0 and 5")
	}
	if action.Cursor == (Vec3{}) {
		action.Cursor = Vec3{.5, .5, .5}
	}
	if action.Cursor.X < 0 || action.Cursor.X > 1 || action.Cursor.Y < 0 || action.Cursor.Y > 1 || action.Cursor.Z < 0 || action.Cursor.Z > 1 {
		return fmt.Errorf("minego: block cursor must be inside the unit cube")
	}
	lease, err := i.bot.actions.acquire(ctx, controlView|controlHands, priorityExplicit)
	if err != nil {
		return err
	}
	defer lease.Release()
	ctx = lease.Context(ctx)
	hit := Vec3{float64(action.Position.X) + action.Cursor.X, float64(action.Position.Y) + action.Cursor.Y, float64(action.Position.Z) + action.Cursor.Z}
	if err := i.lookAt(ctx, hit); err != nil {
		return err
	}
	if action.Sneaking {
		state := i.bot.Self.State()
		if err := i.bot.send(ctx, &packets.C2SPlayerCommand{EntityId: ns.VarInt(state.EntityID), ActionId: 0}); err != nil {
			return err
		}
		defer func() {
			_ = i.bot.send(context.Background(), &packets.C2SPlayerCommand{EntityId: ns.VarInt(state.EntityID), ActionId: 1})
		}()
	}
	inside := i.bot.Self.State().Position.Block() == action.Position
	p := &packets.C2SUseItemOn{Hand: ns.VarInt(action.Hand), Location: ns.NewPosition(action.Position.X, action.Position.Y, action.Position.Z), Face: ns.VarInt(action.Face), CursorPositionX: ns.Float32(action.Cursor.X), CursorPositionY: ns.Float32(action.Cursor.Y), CursorPositionZ: ns.Float32(action.Cursor.Z), InsideBlock: ns.Boolean(inside), Sequence: ns.VarInt(i.bot.nextSequence())}
	if err := i.bot.send(ctx, p); err != nil {
		return err
	}
	if action.Swing {
		return i.Swing(ctx, action.Hand)
	}
	return nil
}

func (i *Interaction) ActivateEntity(ctx context.Context, action EntityInteraction) error {
	if err := validHand(action.Hand); err != nil {
		return err
	}
	entity, ok := i.bot.Entities.Get(action.EntityID)
	if !ok {
		return fmt.Errorf("%w: entity %d", ErrTargetLost, action.EntityID)
	}
	if action.Reach <= 0 {
		action.Reach = 4.5
	}
	eye := i.bot.Self.State().Position
	eye.Y += 1.62
	target := entity.Position
	if action.Location != (Vec3{}) {
		target.X += action.Location.X
		target.Y += action.Location.Y
		target.Z += action.Location.Z
	}
	if eye.Distance(target) > action.Reach {
		return ErrOutOfReach
	}
	lease, err := i.bot.actions.acquire(ctx, controlView|controlHands, priorityExplicit)
	if err != nil {
		return err
	}
	defer lease.Release()
	ctx = lease.Context(ctx)
	if err := i.lookAt(ctx, target); err != nil {
		return err
	}
	p := &packets.C2SInteract{EntityId: ns.VarInt(action.EntityID), Hand: ns.VarInt(action.Hand), Location: ns.LpVec3{X: action.Location.X, Y: action.Location.Y, Z: action.Location.Z}, UsingSecondaryAction: ns.Boolean(action.Sneaking)}
	if err := i.bot.send(ctx, p); err != nil {
		return err
	}
	if action.Swing {
		return i.Swing(ctx, action.Hand)
	}
	return nil
}

func validHand(hand Hand) error {
	if hand != MainHand && hand != OffHand {
		return fmt.Errorf("minego: invalid hand %d", hand)
	}
	return nil
}

func (i *Interaction) lookAt(ctx context.Context, target Vec3) error {
	state := i.bot.Self.State()
	eye := state.Position
	eye.Y += 1.62
	dx, dy, dz := target.X-eye.X, target.Y-eye.Y, target.Z-eye.Z
	yaw := float32(math.Atan2(-dx, dz) * 180 / math.Pi)
	pitch := float32(-math.Atan2(dy, math.Hypot(dx, dz)) * 180 / math.Pi)
	if err := i.bot.send(ctx, &packets.C2SMovePlayerRot{Yaw: ns.Float32(yaw), Pitch: ns.Float32(pitch), Flags: boolFlag(state.OnGround, 1)}); err != nil {
		return err
	}
	i.bot.Self.update(func(s *SelfState) { s.Rotation = Rotation{Yaw: yaw, Pitch: pitch} })
	return nil
}

func boolFlag(v bool, flag int8) ns.Int8 {
	if v {
		return ns.Int8(flag)
	}
	return 0
}

type AttackOptions struct {
	Reach    float64
	Swing    bool
	Cooldown time.Duration
}

type CombatEvent struct {
	Kind, Detail string
	EntityID     int32
	SourceID     int32
}

type Combat struct {
	bot     *Bot
	onEvent event[CombatEvent]
}

func newCombat(bot *Bot) *Combat                      { return &Combat{bot: bot} }
func (c *Combat) OnEvent(fn func(CombatEvent)) func() { return c.onEvent.subscribe(fn) }

func (c *Combat) Attack(ctx context.Context, entityID int32, opt AttackOptions) error {
	entity, ok := c.bot.Entities.Get(entityID)
	if !ok {
		return fmt.Errorf("%w: entity %d", ErrTargetLost, entityID)
	}
	if opt.Reach <= 0 {
		opt.Reach = 3
	}
	eye := c.bot.Self.State().Position
	eye.Y += 1.62
	target := entity.Position
	target.Y += .9
	if eye.Distance(target) > opt.Reach {
		return ErrOutOfReach
	}
	lease, err := c.bot.actions.acquire(ctx, controlView|controlHands, priorityExplicit)
	if err != nil {
		return err
	}
	defer lease.Release()
	ctx = lease.Context(ctx)
	if err := c.bot.Interaction.lookAt(ctx, target); err != nil {
		return err
	}
	if err := c.bot.send(ctx, &packets.C2SAttack{EntityId: ns.VarInt(entityID)}); err != nil {
		return err
	}
	if opt.Swing {
		if err := c.bot.Interaction.Swing(ctx, MainHand); err != nil {
			return err
		}
	}
	c.onEvent.emit(CombatEvent{Kind: "attack", EntityID: entityID})
	return nil
}

// Fight repeatedly attacks at the requested cooldown until the entity is
// removed or the context is cancelled. A zero cooldown uses 625 ms.
func (c *Combat) Fight(ctx context.Context, entityID int32, opt AttackOptions) error {
	if opt.Cooldown <= 0 {
		opt.Cooldown = 625 * time.Millisecond
	}
	if !opt.Swing {
		opt.Swing = true
	}
	for {
		if _, ok := c.bot.Entities.Get(entityID); !ok {
			return nil
		}
		if err := c.Attack(ctx, entityID, opt); err != nil {
			return err
		}
		t := time.NewTimer(opt.Cooldown)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-c.bot.done:
			t.Stop()
			return ErrNotConnected
		case <-t.C:
		}
	}
}
