package minego

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/registries"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

type PlaceOptions struct {
	// Item is a block item identifier. An empty value uses the selected item.
	Item    string
	Reach   float64
	Swing   bool
	Timeout time.Duration
}

type PlaceResult struct {
	Position BlockPos
	Block    Block
	Item     ItemStack
	Support  BlockPos
	Face     int
}

type Builder struct {
	bot        *Bot
	onProgress event[BuildProgress]
}

func newBuilder(bot *Bot) *Builder { return &Builder{bot: bot} }

// Place selects a matching hotbar block, clicks a supporting face, and waits
// for the server's authoritative block update at pos.
func (b *Builder) Place(ctx context.Context, pos BlockPos, opt PlaceOptions) (PlaceResult, error) {
	lease, err := b.bot.actions.acquire(ctx, controlView|controlHands, priorityExplicit)
	if err != nil {
		return PlaceResult{}, err
	}
	defer lease.Release()
	ctx = lease.Context(ctx)
	if opt.Reach <= 0 {
		opt.Reach = 4.5
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 10 * time.Second
	}
	if !opt.Swing {
		opt.Swing = true
	}
	old, ok := b.bot.World.Block(pos)
	if !ok {
		return PlaceResult{}, fmt.Errorf("minego: placement position is not loaded")
	}
	if !replaceableBlock(old.Name) {
		return PlaceResult{}, fmt.Errorf("minego: %s is not replaceable", old.Name)
	}

	item, hotbar, err := b.placementItem(opt.Item)
	if err != nil {
		return PlaceResult{}, err
	}
	if _, ok := b.bot.pack.StateID(item.Name, nil); !ok {
		return PlaceResult{}, fmt.Errorf("%w: %s has no block state", ErrNoPlacementItem, item.Name)
	}
	support, face, cursor, ok := b.supportFace(pos)
	if !ok {
		return PlaceResult{}, ErrNoPlacementFace
	}
	eye := b.bot.Self.State().Position
	eye.Y += 1.62
	hit := Vec3{float64(support.X) + cursor.X, float64(support.Y) + cursor.Y, float64(support.Z) + cursor.Z}
	if eye.Distance(hit) > opt.Reach {
		return PlaceResult{}, fmt.Errorf("minego: placement face is outside reach")
	}
	if !b.bot.Miner.lineOfSight(eye, hit, support) {
		return PlaceResult{}, fmt.Errorf("minego: placement face is obstructed")
	}
	if err := b.bot.Inventory.Select(ctx, hotbar); err != nil {
		return PlaceResult{}, err
	}

	changed := make(chan Block, 1)
	unsubscribe := b.bot.World.OnBlockChange(func(change BlockChange) {
		if change.Position == pos && change.New.StateID != old.StateID {
			select {
			case changed <- change.New:
			default:
			}
		}
	})
	defer unsubscribe()
	sequence := b.bot.Miner.sequence.Add(1)
	packet := &packets.C2SUseItemOn{
		Hand: 0, Location: ns.NewPosition(support.X, support.Y, support.Z), Face: ns.VarInt(face),
		CursorPositionX: ns.Float32(cursor.X), CursorPositionY: ns.Float32(cursor.Y), CursorPositionZ: ns.Float32(cursor.Z),
		Sequence: ns.VarInt(sequence),
	}
	if err := b.bot.send(ctx, packet); err != nil {
		return PlaceResult{}, err
	}
	if opt.Swing {
		_ = b.bot.send(ctx, &packets.C2SSwing{Hand: 0})
	}
	timer := time.NewTimer(opt.Timeout)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return PlaceResult{}, ctx.Err()
	case <-b.bot.done:
		return PlaceResult{}, ErrNotConnected
	case <-timer.C:
		return PlaceResult{}, fmt.Errorf("minego: server did not confirm block placement")
	case block := <-changed:
		return PlaceResult{Position: pos, Block: block, Item: item, Support: support, Face: face}, nil
	}
}

func (b *Builder) placementItem(name string) (ItemStack, int, error) {
	if name != "" && !strings.Contains(name, ":") {
		name = "minecraft:" + name
	}
	slots := b.bot.Inventory.Slots()
	selected := b.bot.Inventory.Selected()
	for hotbar := 0; hotbar < 9; hotbar++ {
		idx := 36 + hotbar
		if idx >= len(slots) {
			idx = hotbar
		}
		item := slots[idx]
		if item.Count <= 0 {
			continue
		}
		if name != "" && item.Name != name {
			continue
		}
		if name == "" && hotbar != selected {
			continue
		}
		return item, hotbar, nil
	}
	return ItemStack{}, -1, ErrNoPlacementItem
}

type placementFace struct {
	dx, dy, dz int
	face       int
	cursor     Vec3
}

var placementFaces = [...]placementFace{
	{0, -1, 0, 1, Vec3{.5, 1, .5}},
	{0, 0, -1, 3, Vec3{.5, .5, 1}},
	{0, 0, 1, 2, Vec3{.5, .5, 0}},
	{-1, 0, 0, 5, Vec3{1, .5, .5}},
	{1, 0, 0, 4, Vec3{0, .5, .5}},
	{0, 1, 0, 0, Vec3{.5, 0, .5}},
}

func (b *Builder) supportFace(pos BlockPos) (BlockPos, int, Vec3, bool) {
	for _, candidate := range placementFaces {
		support := BlockPos{pos.X + candidate.dx, pos.Y + candidate.dy, pos.Z + candidate.dz}
		block, ok := b.bot.World.Block(support)
		if ok && len(block.Collision) > 0 {
			return support, candidate.face, candidate.cursor, true
		}
	}
	return BlockPos{}, 0, Vec3{}, false
}

func replaceableBlock(name string) bool {
	for _, block := range registries.TagData["minecraft:block"]["minecraft:replaceable"] {
		if block == name {
			return true
		}
	}
	return false
}
