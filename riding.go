package minego

import (
	"context"
	"fmt"
	"time"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

type Riding struct{ bot *Bot }

func newRiding(bot *Bot) *Riding { return &Riding{bot: bot} }

func (r *Riding) Vehicle() (Entity, bool) {
	id := r.bot.Self.State().VehicleID
	if id == 0 {
		return Entity{}, false
	}
	return r.bot.Entities.Get(id)
}

func (r *Riding) Mount(ctx context.Context, entityID int32) error {
	if err := r.bot.Interaction.ActivateEntity(ctx, EntityInteraction{EntityID: entityID, Hand: MainHand, Reach: 4.5}); err != nil {
		return err
	}
	return r.waitVehicle(ctx, entityID)
}

func (r *Riding) Dismount(ctx context.Context) error {
	state := r.bot.Self.State()
	if state.VehicleID == 0 {
		return nil
	}
	if err := r.bot.send(ctx, &packets.C2SPlayerCommand{EntityId: ns.VarInt(state.EntityID), ActionId: 0}); err != nil {
		return err
	}
	err := r.waitVehicle(ctx, 0)
	_ = r.bot.send(context.Background(), &packets.C2SPlayerCommand{EntityId: ns.VarInt(state.EntityID), ActionId: 1})
	return err
}

func (r *Riding) MoveVehicle(ctx context.Context, position Vec3, rotation Rotation, onGround bool) error {
	if r.bot.Self.State().VehicleID == 0 {
		return fmt.Errorf("minego: bot is not mounted")
	}
	return r.bot.send(ctx, &packets.C2SMoveVehicle{X: ns.Float64(position.X), Y: ns.Float64(position.Y), Z: ns.Float64(position.Z), Yaw: ns.Float32(rotation.Yaw), Pitch: ns.Float32(rotation.Pitch), OnGround: ns.Boolean(onGround)})
}

func (r *Riding) Paddle(ctx context.Context, left, right bool) error {
	return r.bot.send(ctx, &packets.C2SPaddleBoat{LeftPaddleTurning: ns.Boolean(left), RightPaddleTurning: ns.Boolean(right)})
}

func (r *Riding) SetCreativeFlight(ctx context.Context, flying bool) error {
	state := r.bot.Self.State()
	if !state.CanFly && state.GameMode != 1 && state.GameMode != 3 {
		return ErrInvalidGameMode
	}
	flags := int8(0)
	if flying {
		flags = 2
	}
	if err := r.bot.send(ctx, &packets.C2SPlayerAbilities{Flags: ns.Int8(flags)}); err != nil {
		return err
	}
	r.bot.Self.update(func(s *SelfState) { s.Flying = flying })
	return nil
}

func (r *Riding) FlyTo(ctx context.Context, position Vec3, rotation Rotation) error {
	if !r.bot.Self.State().Flying {
		return fmt.Errorf("minego: creative flight is not enabled")
	}
	if err := r.bot.send(ctx, &packets.C2SMovePlayerPosRot{X: ns.Float64(position.X), FeetY: ns.Float64(position.Y), Z: ns.Float64(position.Z), Yaw: ns.Float32(rotation.Yaw), Pitch: ns.Float32(rotation.Pitch)}); err != nil {
		return err
	}
	r.bot.Self.update(func(s *SelfState) { s.Position = position; s.Rotation = rotation; s.OnGround = false })
	return nil
}

func (r *Riding) waitVehicle(ctx context.Context, id int32) error {
	if r.bot.Self.State().VehicleID == id {
		return nil
	}
	ch := make(chan struct{}, 1)
	unsub := r.bot.Self.OnChange(func(s SelfState) {
		if s.VehicleID == id {
			select {
			case ch <- struct{}{}:
			default:
			}
		}
	})
	defer unsub()
	t := time.NewTimer(5 * time.Second)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.bot.done:
		return ErrNotConnected
	case <-t.C:
		return fmt.Errorf("minego: server did not confirm passenger state")
	case <-ch:
		return nil
	}
}
