package minego

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/registries"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

type DigOptions struct {
	Reach      float64
	Face       int
	Swing      bool
	Timeout    time.Duration
	SelectTool bool
}
type DigResult struct {
	Position BlockPos
	Block    Block
	Tool     ItemStack
	Duration time.Duration
}
type MiningProgress struct {
	Kind                 string
	Position             BlockPos
	Completed, Requested int
}
type Selector struct {
	Names     []string
	Tags      []string
	Predicate func(Block) bool
}

func Blocks(names ...string) Selector { return Selector{Names: names} }
func Tags(names ...string) Selector   { return Selector{Tags: names} }

type MineOptions struct {
	Navigation        NavigationOptions
	Dig               DigOptions
	ExplorationRadius int
	Timeout           time.Duration
}
type MineStatus uint8

const (
	MineComplete MineStatus = iota
	MineCancelled
	MineExhausted
	MineUnreachable
	MineToolFailure
	MineDisconnected
)

type MineResult struct {
	Status               MineStatus
	Requested, Completed int
	Mined                []BlockPos
	ExploredChunks       int
}

// digGoal keeps the player beside the target. GoalAdjacent also accepts the
// cell directly above a block, which would let the miner remove its own
// support and fall into the hole.
type digGoal struct{ target BlockPos }

func (g digGoal) Reached(p BlockPos) bool {
	dx, dy, dz := abs(p.X-g.target.X), abs(p.Y-g.target.Y), abs(p.Z-g.target.Z)
	horizontal := max(dx, dz)
	return horizontal >= 1 && horizontal <= 2 && dy <= 2
}

func (g digGoal) Estimate(p BlockPos) float64 {
	return math.Max(0, distance(g.target, p)-2)
}

type Miner struct {
	bot        *Bot
	sequence   atomic.Int32
	onProgress event[MiningProgress]
}

func newMiner(b *Bot) *Miner                               { return &Miner{bot: b} }
func (m *Miner) OnProgress(fn func(MiningProgress)) func() { return m.onProgress.subscribe(fn) }

func (m *Miner) Dig(ctx context.Context, pos BlockPos, opt DigOptions) (DigResult, error) {
	lease, err := m.bot.actions.acquire(ctx, controlView|controlHands, priorityExplicit)
	if err != nil {
		return DigResult{}, err
	}
	defer lease.Release()
	ctx = lease.Context(ctx)
	if opt.Reach <= 0 {
		opt.Reach = 4.5
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 30 * time.Second
	}
	if !opt.Swing {
		opt.Swing = true
	}
	if !opt.SelectTool {
		opt.SelectTool = true
	}
	block, ok := m.bot.World.Block(pos)
	if !ok {
		return DigResult{}, fmt.Errorf("minego: block is not loaded")
	}
	if block.Name == "minecraft:air" || block.Hardness < 0 {
		return DigResult{}, fmt.Errorf("minego: block %s cannot be dug", block.Name)
	}
	self := m.bot.Self.State()
	eye := Vec3{self.Position.X, self.Position.Y + 1.62, self.Position.Z}
	center := Vec3{float64(pos.X) + .5, float64(pos.Y) + .5, float64(pos.Z) + .5}
	if eye.Distance(center) > opt.Reach {
		return DigResult{}, fmt.Errorf("minego: block outside reach")
	}
	if !m.lineOfSight(eye, center, pos) {
		return DigResult{}, fmt.Errorf("minego: line of sight is obstructed")
	}
	tool, slot, speed := m.bestTool(block)
	if block.RequiresCorrectTool && slot < 0 {
		return DigResult{}, fmt.Errorf("minego: no suitable tool for %s", block.Name)
	}
	if opt.SelectTool && slot >= 0 {
		if err := m.bot.Inventory.Select(ctx, slot); err != nil {
			return DigResult{}, err
		}
	}
	duration := breakDuration(block, speed, self)
	seq := m.sequence.Add(1)
	face := opt.Face
	if face < 0 || face > 5 {
		face = nearestFace(eye, center)
	}
	changed := make(chan struct{}, 1)
	unsubscribe := m.bot.World.OnBlockChange(func(c BlockChange) {
		if c.Position == pos && c.New.StateID != block.StateID {
			select {
			case changed <- struct{}{}:
			default:
			}
		}
	})
	defer unsubscribe()
	action := func(status int32) error {
		return m.bot.send(ctx, &packets.C2SPlayerAction{Status: ns.VarInt(status), Location: ns.NewPosition(pos.X, pos.Y, pos.Z), Face: ns.Int8(face), Sequence: ns.VarInt(seq)})
	}
	if err := action(0); err != nil {
		return DigResult{}, err
	}
	if opt.Swing {
		_ = m.bot.send(ctx, &packets.C2SSwing{Hand: 0})
	}
	m.onProgress.emit(MiningProgress{Kind: "started", Position: pos})
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		_ = action(1)
		return DigResult{}, ctx.Err()
	case <-m.bot.done:
		_ = action(1)
		return DigResult{}, ErrNotConnected
	case <-timer.C:
	}
	if err := action(2); err != nil {
		return DigResult{}, err
	}
	timeout := time.NewTimer(opt.Timeout)
	defer timeout.Stop()
	select {
	case <-ctx.Done():
		_ = action(1)
		return DigResult{}, ctx.Err()
	case <-m.bot.done:
		return DigResult{}, ErrNotConnected
	case <-timeout.C:
		return DigResult{}, fmt.Errorf("minego: server did not confirm block break")
	case <-changed:
		m.onProgress.emit(MiningProgress{Kind: "completed", Position: pos})
		return DigResult{pos, block, tool, duration}, nil
	}
}

func (m *Miner) Mine(ctx context.Context, selector Selector, count int, opt MineOptions) (MineResult, error) {
	result := MineResult{Requested: count}
	if count <= 0 {
		result.Status = MineComplete
		return result, nil
	}
	if opt.ExplorationRadius <= 0 {
		opt.ExplorationRadius = 256
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, opt.Timeout)
	defer cancel()
	origin := m.bot.Self.State().Position.Block()
	seen := map[BlockPos]bool{}
	for result.Completed < count {
		targets := m.search(selector, seen)
		mined := false
		for _, target := range targets {
			seen[target.Position] = true
			_, err := m.bot.Navigator.Navigate(ctx, digGoal{target.Position}, opt.Navigation)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					result.Status = MineCancelled
					return result, err
				}
				continue
			}
			_, err = m.Dig(ctx, target.Position, opt.Dig)
			if err != nil {
				if errors.Is(err, ErrNotConnected) {
					result.Status = MineDisconnected
					return result, err
				}
				continue
			}
			result.Completed++
			result.Mined = append(result.Mined, target.Position)
			m.onProgress.emit(MiningProgress{Kind: "target", Position: target.Position, Completed: result.Completed, Requested: count})
			mined = true
			break
		}
		if mined {
			continue
		}
		frontier, ok := m.frontier(origin, opt.ExplorationRadius)
		if !ok {
			result.Status = MineExhausted
			return result, ErrSearchExhausted
		}
		before := len(m.bot.World.LoadedChunks())
		_, err := m.bot.Navigator.Navigate(ctx, GoalNear{frontier, 2}, opt.Navigation)
		if err != nil {
			result.Status = MineUnreachable
			return result, err
		}
		wait := time.NewTimer(3 * time.Second)
		select {
		case <-ctx.Done():
			wait.Stop()
			result.Status = MineCancelled
			return result, ctx.Err()
		case <-wait.C:
		}
		if len(m.bot.World.LoadedChunks()) <= before {
			result.Status = MineExhausted
			return result, ErrSearchExhausted
		}
		result.ExploredChunks++
	}
	result.Status = MineComplete
	return result, nil
}

func (m *Miner) search(s Selector, skip map[BlockPos]bool) []Block {
	names := map[string]bool{}
	for _, x := range s.Names {
		names[x] = true
	}
	for _, tag := range s.Tags {
		if !strings.Contains(tag, ":") {
			tag = "minecraft:" + tag
		}
		for _, x := range registries.TagData["minecraft:block"][tag] {
			names[x] = true
		}
	}
	var out []Block
	type cachedState struct {
		block Block
		ok    bool
		match bool
	}
	// A loaded view contains millions of positions but normally only a few
	// hundred distinct state IDs. Decode names, properties, hardness, and
	// collision once per state rather than once per position.
	states := make(map[int32]cachedState, 512)
	m.bot.World.mu.RLock()
	for key, col := range m.bot.World.chunks {
		for y := -64; y < 320; y++ {
			for z := 0; z < 16; z++ {
				for x := 0; x < 16; x++ {
					pos := BlockPos{int(key.X)*16 + x, y, int(key.Z)*16 + z}
					if skip[pos] {
						continue
					}
					id := col.GetBlockState(x, y, z)
					cached, known := states[id]
					if !known {
						b, ok := m.bot.block(BlockPos{}, id)
						cached = cachedState{block: b, ok: ok, match: ok && names[b.Name]}
						states[id] = cached
					}
					if !cached.ok {
						continue
					}
					b := cached.block
					b.Position = pos
					if cached.match || (s.Predicate != nil && s.Predicate(b)) {
						out = append(out, b)
					}
				}
			}
		}
	}
	m.bot.World.mu.RUnlock()
	self := m.bot.Self.State().Position
	sort.Slice(out, func(i, j int) bool {
		return self.Distance(Vec3{float64(out[i].Position.X), float64(out[i].Position.Y), float64(out[i].Position.Z)}) < self.Distance(Vec3{float64(out[j].Position.X), float64(out[j].Position.Y), float64(out[j].Position.Z)})
	})
	return out
}
func (m *Miner) frontier(origin BlockPos, radius int) (BlockPos, bool) {
	chunks := m.bot.World.LoadedChunks()
	loaded := map[[2]int32]bool{}
	for _, c := range chunks {
		loaded[c] = true
	}
	self := m.bot.Self.State().Position.Block()
	best := math.MaxFloat64
	var result BlockPos
	found := false
	for _, c := range chunks {
		for _, d := range [][2]int32{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			if loaded[[2]int32{c[0] + d[0], c[1] + d[1]}] {
				continue
			}
			x, z := int(c[0]*16+8+d[0]*7), int(c[1]*16+8+d[1]*7)
			if abs(x-origin.X) > radius || abs(z-origin.Z) > radius {
				continue
			}
			for y := 319; y >= -63; y-- {
				p := BlockPos{x, y, z}
				if _, _, ok := m.bot.Navigator.passable(p, NavigationOptions{}); ok {
					dist := distance(self, p)
					if dist < best {
						best = dist
						result = p
						found = true
					}
					break
				}
			}
		}
	}
	return result, found
}
func (m *Miner) lineOfSight(from, to Vec3, target BlockPos) bool {
	dist := from.Distance(to)
	steps := int(dist / .1)
	for i := 1; i < steps; i++ {
		t := float64(i) / float64(steps)
		p := Vec3{from.X + (to.X-from.X)*t, from.Y + (to.Y-from.Y)*t, from.Z + (to.Z-from.Z)*t}.Block()
		if p == target {
			continue
		}
		b, ok := m.bot.World.Block(p)
		if !ok || len(b.Collision) > 0 {
			return false
		}
	}
	return true
}
func (m *Miner) bestTool(b Block) (ItemStack, int, float64) {
	slots := m.bot.Inventory.Slots()
	want := ""
	switch {
	case strings.Contains(b.Name, "stone") || strings.Contains(b.Name, "ore"):
		want = "pickaxe"
	case strings.Contains(b.Name, "log") || strings.Contains(b.Name, "wood"):
		want = "axe"
	case strings.Contains(b.Name, "dirt") || strings.Contains(b.Name, "sand") || strings.Contains(b.Name, "gravel"):
		want = "shovel"
	case strings.Contains(b.Name, "leaves") || strings.Contains(b.Name, "wool"):
		want = "shears"
	}
	bestSlot := -1
	speed := 1.0
	var best ItemStack
	for hot := 0; hot < 9; hot++ {
		idx := 36 + hot
		if idx >= len(slots) {
			idx = hot
		}
		s := slots[idx]
		if want != "" && strings.Contains(s.Name, want) {
			v := toolSpeed(s.Name)
			if v > speed {
				speed = v
				best = s
				bestSlot = hot
			}
		}
	}
	return best, bestSlot, speed
}
func toolSpeed(name string) float64 {
	switch {
	case strings.Contains(name, "netherite_"):
		return 9
	case strings.Contains(name, "diamond_"):
		return 8
	case strings.Contains(name, "golden_"):
		return 12
	case strings.Contains(name, "iron_"):
		return 6
	case strings.Contains(name, "stone_"):
		return 4
	case strings.Contains(name, "wooden_"):
		return 2
	case strings.Contains(name, "shears"):
		return 5
	}
	return 1
}
func breakDuration(b Block, speed float64, s SelfState) time.Duration {
	if b.Hardness <= 0 {
		return 50 * time.Millisecond
	}
	seconds := float64(b.Hardness) * 1.5 / speed
	if b.RequiresCorrectTool && speed <= 1 {
		seconds = float64(b.Hardness) * 5
	}
	if e, ok := s.Effects[3]; ok {
		seconds /= 1 + .2*float64(e.Amplifier+1)
	}
	if e, ok := s.Effects[4]; ok {
		seconds *= math.Pow(.3, float64(e.Amplifier+1))
	}
	// The server applies destroy progress on tick boundaries. Keep mining for
	// one extra tick so STOP_DESTROY_BLOCK cannot arrive before the final
	// damage tick is credited.
	ticks := math.Ceil(seconds*20) + 1
	return time.Duration(ticks) * 50 * time.Millisecond
}
func nearestFace(eye, center Vec3) int {
	dx, dy, dz := eye.X-center.X, eye.Y-center.Y, eye.Z-center.Z
	if math.Abs(dy) >= math.Abs(dx) && math.Abs(dy) >= math.Abs(dz) {
		if dy > 0 {
			return 1
		}
		return 0
	}
	if math.Abs(dx) >= math.Abs(dz) {
		if dx > 0 {
			return 5
		}
		return 4
	}
	if dz > 0 {
		return 3
	}
	return 2
}
