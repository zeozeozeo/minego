package minego

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

type ElytraOptions struct {
	CruiseAltitude   int
	Clearance        int
	LandingRadius    int
	RocketInterval   time.Duration
	RocketReserve    int
	AutoEquip        bool
	RestoreEquipment bool
}
type ElytraResult struct {
	LaunchPosition   Vec3
	LandingPosition  Vec3
	RocketsUsed      int
	Replans          int
	EmergencyLanding bool
}
type Elytra struct{ bot *Bot }

func newElytra(bot *Bot) *Elytra { return &Elytra{bot: bot} }

func (e *Elytra) Fly(ctx context.Context, goal Goal, opt ElytraOptions) (ElytraResult, error) {
	var result ElytraResult
	target, ok := elytraTarget(goal)
	if !ok {
		return result, fmt.Errorf("minego: elytra requires a block, near, or adjacent goal")
	}
	if opt.CruiseAltitude <= 0 {
		opt.CruiseAltitude = 32
	}
	if opt.Clearance <= 0 {
		opt.Clearance = 6
	}
	if opt.LandingRadius <= 0 {
		opt.LandingRadius = 8
	}
	if opt.RocketInterval <= 0 {
		opt.RocketInterval = 3 * time.Second
	}
	if !opt.AutoEquip {
		opt.AutoEquip = true
	}
	if !opt.RestoreEquipment {
		opt.RestoreEquipment = true
	}
	lease, err := e.bot.actions.acquire(ctx, controlMovement|controlView|controlHands|controlInventory, priorityExplicit)
	if err != nil {
		return result, err
	}
	defer lease.Release()
	ctx = lease.Context(ctx)
	selected := e.bot.Inventory.Selected()
	elytraSlot := findItem(e.bot.Inventory.Slots(), "minecraft:elytra")
	if elytraSlot < 0 {
		return result, ErrNoElytra
	}
	rocketSlot := findHotbar(e.bot.Inventory.Slots(), "minecraft:firework_rocket")
	rocketInventorySlot := findItem(e.bot.Inventory.Slots(), "minecraft:firework_rocket")
	if rocketInventorySlot < 0 {
		return result, ErrNoFireworks
	}
	var restore func() error = func() error { return nil }
	if elytraSlot != 6 {
		restore, err = e.swapInventory(ctx, elytraSlot, 6)
		if err != nil {
			return result, err
		}
	}
	var restoreRocket func() error = func() error { return nil }
	if rocketSlot < 0 {
		dest := 36 + selected
		if dest >= len(e.bot.Inventory.Slots()) {
			dest = selected
		}
		restoreRocket, err = e.swapInventory(ctx, rocketInventorySlot, dest)
		if err != nil {
			return result, err
		}
		rocketSlot = selected
	}
	if opt.RestoreEquipment {
		defer func() { _ = restoreRocket(); _ = restore(); _ = e.bot.Inventory.Select(context.Background(), selected) }()
	}
	if err = e.bot.Inventory.Select(ctx, rocketSlot); err != nil {
		return result, err
	}
	state := e.bot.Self.State()
	result.LaunchPosition = state.Position
	// A normal jump followed by START_FALL_FLYING works from level ground on
	// modern servers and avoids building a permanent launch tower.
	e.bot.Navigator.move(physicsInput{Jump: true}, state)
	timer := time.NewTimer(100 * time.Millisecond)
	select {
	case <-ctx.Done():
		timer.Stop()
		return result, ctx.Err()
	case <-timer.C:
	}
	state = e.bot.Self.State()
	if err = e.bot.send(ctx, &packets.C2SPlayerCommand{EntityId: ns.VarInt(state.EntityID), ActionId: 8}); err != nil {
		return result, err
	}
	if err = e.useRocket(ctx, state); err != nil {
		return result, err
	}
	result.RocketsUsed++
	landing, landOK := e.safeLanding(target, opt.LandingRadius)
	cruiseY := math.Max(state.Position.Y, float64(target.Y)) + float64(opt.CruiseAltitude)
	lastRocket := time.Now()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-ticker.C:
		}
		state = e.bot.Self.State()
		destination := Vec3{float64(target.X) + .5, cruiseY, float64(target.Z) + .5}
		horizontal := math.Hypot(destination.X-state.Position.X, destination.Z-state.Position.Z)
		if horizontal < 24 && landOK {
			destination = Vec3{float64(landing.X) + .5, float64(landing.Y), float64(landing.Z) + .5}
		}
		if !e.corridorClear(state.Position, destination, opt.Clearance) {
			cruiseY += 8
			destination.Y = cruiseY
			result.Replans++
		}
		dx, dy, dz := destination.X-state.Position.X, destination.Y-state.Position.Y, destination.Z-state.Position.Z
		hd := math.Hypot(dx, dz)
		if hd > 0 {
			dx, dz = dx/hd*.85, dz/hd*.85
		}
		dy = math.Max(-.35, math.Min(.25, dy*.08))
		next := Vec3{state.Position.X + dx, state.Position.Y + dy, state.Position.Z + dz}
		yaw := float32(math.Atan2(-dx, dz) * 180 / math.Pi)
		pitch := float32(-math.Atan2(dy, math.Max(.001, hd)) * 180 / math.Pi)
		flags := ns.Int8(0)
		if state.OnGround {
			flags |= 1
		}
		if err = e.bot.send(ctx, &packets.C2SMovePlayerPosRot{X: ns.Float64(next.X), FeetY: ns.Float64(next.Y), Z: ns.Float64(next.Z), Yaw: ns.Float32(yaw), Pitch: ns.Float32(pitch), Flags: flags}); err != nil {
			return result, err
		}
		e.bot.Self.update(func(s *SelfState) {
			s.Position = next
			s.Velocity = Vec3{dx, dy, dz}
			s.Rotation = Rotation{Yaw: yaw, Pitch: pitch}
			s.Flying = true
		})
		if time.Since(lastRocket) >= opt.RocketInterval && horizontal >= 24 {
			if countItem(e.bot.Inventory.Slots(), "minecraft:firework_rocket") <= opt.RocketReserve {
				result.EmergencyLanding = true
			} else if err = e.useRocket(ctx, state); err == nil {
				result.RocketsUsed++
				lastRocket = time.Now()
			}
		}
		if landOK && next.Distance(Vec3{float64(landing.X) + .5, float64(landing.Y), float64(landing.Z) + .5}) < .8 {
			e.bot.Self.update(func(s *SelfState) { s.OnGround = true; s.Flying = false; s.Velocity = Vec3{} })
			result.LandingPosition = next
			return result, nil
		}
		if !landOK && horizontal < 32 {
			landing, landOK = e.safeLanding(target, opt.LandingRadius)
			if !landOK {
				return result, ErrNoSafeLanding
			}
		}
	}
}
func (e *Elytra) useRocket(ctx context.Context, state SelfState) error {
	seq := e.bot.nextSequence()
	return e.bot.send(ctx, &packets.C2SUseItem{Hand: 0, Sequence: ns.VarInt(seq), Yaw: ns.Float32(state.Rotation.Yaw), Pitch: ns.Float32(state.Rotation.Pitch)})
}
func (e *Elytra) swapInventory(ctx context.Context, a, b int) (func() error, error) {
	w := e.bot.Crafter.window(0)
	for _, slot := range []int{a, b, a} {
		if err := e.bot.Crafter.click(ctx, w, slot, 0, 0); err != nil {
			return nil, err
		}
		w = e.bot.Crafter.window(0)
	}
	return func() error {
		w := e.bot.Crafter.window(0)
		for _, slot := range []int{a, b, a} {
			if err := e.bot.Crafter.click(context.Background(), w, slot, 0, 0); err != nil {
				return err
			}
			w = e.bot.Crafter.window(0)
		}
		return nil
	}, nil
}
func (e *Elytra) corridorClear(from, to Vec3, clearance int) bool {
	distance := from.Distance(to)
	steps := int(distance / 2)
	if steps < 1 {
		steps = 1
	}
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		p := Vec3{from.X + (to.X-from.X)*t, from.Y + (to.Y-from.Y)*t, from.Z + (to.Z-from.Z)*t}.Block()
		for y := -clearance / 2; y <= clearance/2; y++ {
			b, ok := e.bot.World.Block(BlockPos{p.X, p.Y + y, p.Z})
			if ok && len(b.Collision) > 0 {
				return false
			}
		}
	}
	return true
}
func (e *Elytra) safeLanding(target BlockPos, radius int) (BlockPos, bool) {
	for r := 0; r <= radius; r++ {
		for x := -r; x <= r; x++ {
			for z := -r; z <= r; z++ {
				if abs(x) != r && abs(z) != r {
					continue
				}
				for y := target.Y + 16; y >= target.Y-16; y-- {
					p := BlockPos{target.X + x, y, target.Z + z}
					below, ok := e.bot.World.Block(BlockPos{p.X, p.Y - 1, p.Z})
					if ok && len(below.Collision) > 0 && e.bot.Navigator.bodyClear(p) {
						return p, true
					}
				}
			}
		}
	}
	return BlockPos{}, false
}
func elytraTarget(goal Goal) (BlockPos, bool) {
	switch g := goal.(type) {
	case GoalBlock:
		return BlockPos(g), true
	case GoalNear:
		return g.Position, true
	case GoalAdjacent:
		return BlockPos(g), true
	default:
		return BlockPos{}, false
	}
}
func findItem(slots []ItemStack, name string) int {
	for i, s := range slots {
		if s.Name == name && s.Count > 0 {
			return i
		}
	}
	return -1
}
func findHotbar(slots []ItemStack, name string) int {
	for hot := 0; hot < 9; hot++ {
		idx := 36 + hot
		if idx >= len(slots) {
			idx = hot
		}
		if idx < len(slots) && slots[idx].Name == name && slots[idx].Count > 0 {
			return hot
		}
	}
	return -1
}
func countItem(slots []ItemStack, name string) int {
	n := 0
	for _, s := range slots {
		if s.Name == name {
			n += int(s.Count)
		}
	}
	return n
}
