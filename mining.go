package minego

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
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
	Kind, Name           string
	Position             BlockPos
	Completed, Requested int
	Err                  error
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
	Blocks               []Block
	ExploredChunks       int
}

// digGoal keeps the player beside the target. GoalAdjacent also accepts the
// cell directly above a block, which would let the miner remove its own
// support and fall into the hole.
type digGoal struct{ target BlockPos }

func (g digGoal) Reached(p BlockPos) bool {
	dx, dy, dz := abs(p.X-g.target.X), abs(p.Y-g.target.Y), abs(p.Z-g.target.Z)
	horizontal := max(dx, dz)
	// Stay in an immediately neighboring column. Besides looking more like a
	// player mining a block, this keeps ordinary block drops inside pickup
	// range instead of leaving them scattered two or three blocks away.
	return horizontal == 1 && dy <= 2
}

func (g digGoal) Estimate(p BlockPos) float64 {
	return math.Max(0, distance(g.target, p)-math.Sqrt2)
}

type Miner struct {
	bot            *Bot
	onProgress     event[MiningProgress]
	rejectedMu     sync.RWMutex
	rejected       map[BlockPos]struct{}
	rejectedChunks map[chunkKey]struct{}
	rejections     map[chunkKey]int
	protected      []BlockPos
}

func newMiner(b *Bot) *Miner {
	return &Miner{bot: b, rejected: make(map[BlockPos]struct{}), rejectedChunks: make(map[chunkKey]struct{}), rejections: make(map[chunkKey]int)}
}
func (m *Miner) OnProgress(fn func(MiningProgress)) func() { return m.onProgress.subscribe(fn) }

func (m *Miner) reject(pos BlockPos) {
	m.rejectedMu.Lock()
	if _, exists := m.rejected[pos]; exists {
		m.rejectedMu.Unlock()
		return
	}
	m.rejected[pos] = struct{}{}
	// One missed block update is not enough evidence to discard a whole tree.
	// Repeated refusals at distinct positions in one chunk are a useful signal
	// for vanilla spawn protection or a claim plugin, though, so stop planning
	// destructive edges through that chunk after three such refusals.
	key := chunkKey{int32(pos.X >> 4), int32(pos.Z >> 4)}
	m.rejections[key]++
	if m.rejections[key] >= 3 {
		m.rejectedChunks[key] = struct{}{}
		if m.rejections[key] == 3 {
			m.protected = append(m.protected, pos)
		}
	}
	m.rejectedMu.Unlock()
}

func (m *Miner) breakRejected(pos BlockPos) bool {
	m.rejectedMu.RLock()
	_, rejected := m.rejected[pos]
	if !rejected {
		_, rejected = m.rejectedChunks[chunkKey{int32(pos.X >> 4), int32(pos.Z >> 4)}]
	}
	if !rejected {
		for _, center := range m.protected {
			if abs(pos.X-center.X) <= 16 && abs(pos.Z-center.Z) <= 16 {
				rejected = true
				break
			}
		}
	}
	m.rejectedMu.RUnlock()
	return rejected
}

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
		opt.Timeout = 5 * time.Second
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
	if _, obstructed := m.lineOfSightObstruction(eye, center, pos); obstructed {
		return DigResult{}, fmt.Errorf("minego: line of sight is obstructed")
	}
	// A real client faces the block for the entire dig. This also makes the
	// action intelligible to nearby players instead of leaving the bot staring
	// in its previous travel direction while its hand swings.
	if err := m.bot.Interaction.lookAt(ctx, center); err != nil {
		return DigResult{}, err
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
	timingState := self
	timingState.OnGround = m.bot.Navigator.grounded(self)
	duration := breakDuration(block, speed, timingState)
	face := opt.Face
	if face <= 0 || face > 5 {
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
		return m.bot.send(ctx, &packets.C2SPlayerAction{Status: ns.VarInt(status), Location: ns.NewPosition(pos.X, pos.Y, pos.Z), Face: ns.Int8(face), Sequence: ns.VarInt(m.bot.nextSequence())})
	}
	if err := action(0); err != nil {
		return DigResult{}, err
	}
	if opt.Swing {
		_ = m.bot.send(ctx, &packets.C2SSwing{Hand: 0})
	}
	m.onProgress.emit(MiningProgress{Kind: "started", Name: block.Name, Position: pos})
	timer := time.NewTimer(duration)
	defer timer.Stop()
	swing := time.NewTicker(250 * time.Millisecond)
	defer swing.Stop()
	digging := true
	for digging {
		select {
		case <-ctx.Done():
			_ = action(1)
			return DigResult{}, ctx.Err()
		case <-m.bot.done:
			_ = action(1)
			return DigResult{}, ErrNotConnected
		case <-swing.C:
			if opt.Swing {
				_ = m.bot.send(ctx, &packets.C2SSwing{Hand: 0})
			}
		case <-timer.C:
			digging = false
		}
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
		m.reject(pos)
		m.onProgress.emit(MiningProgress{Kind: "rejected", Name: block.Name, Position: pos})
		return DigResult{}, fmt.Errorf("%w at %v", ErrBlockBreakRejected, pos)
	case <-changed:
		m.onProgress.emit(MiningProgress{Kind: "completed", Name: block.Name, Position: pos})
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
	stagnantExploration := 0
	var targets []Block
	var pickup []BlockPos
	for result.Completed < count {
		if len(targets) > 0 {
			self := m.bot.Self.State().Position
			sort.Slice(targets, func(i, j int) bool {
				return miningTargetScore(self, targets[i], pickup) < miningTargetScore(self, targets[j], pickup)
			})
			target := targets[0]
			targets = targets[1:]
			if seen[target.Position] || m.breakRejected(target.Position) {
				continue
			}
			current, loaded := m.bot.World.Block(target.Position)
			if !loaded || current.Name != target.Name {
				continue
			}
			seen[target.Position] = true
			_, err := m.bot.Navigator.Navigate(ctx, digGoal{target.Position}, opt.Navigation)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					result.Status = MineCancelled
					return result, err
				}
				continue
			}
			err = m.clearLineOfSight(ctx, target.Position, opt.Navigation)
			if err == nil {
				_, err = m.Dig(ctx, target.Position, opt.Dig)
			}
			if err != nil {
				m.onProgress.emit(MiningProgress{Kind: "failed", Name: target.Name, Position: target.Position, Completed: result.Completed, Requested: count, Err: err})
				if errors.Is(err, ErrNotConnected) {
					result.Status = MineDisconnected
					return result, err
				}
				continue
			}
			result.Completed++
			result.Mined = append(result.Mined, target.Position)
			result.Blocks = append(result.Blocks, target)
			pickup = append(pickup, target.Position)
			m.onProgress.emit(MiningProgress{Kind: "target", Name: target.Name, Position: target.Position, Completed: result.Completed, Requested: count})
			if len(pickup) >= 6 || result.Completed == count {
				m.collectDrops(ctx, pickup, opt.Navigation)
				pickup = pickup[:0]
			}
			continue
		}
		targets = m.search(selector, seen)
		if len(targets) > 0 {
			continue
		}
		if len(pickup) > 0 {
			m.collectDrops(ctx, pickup, opt.Navigation)
			pickup = pickup[:0]
		}
		frontier, ok := m.frontier(origin, opt.ExplorationRadius)
		if !ok {
			result.Status = MineExhausted
			return result, ErrSearchExhausted
		}
		step, ok := m.explorationStep(m.bot.Self.State().Position.Block(), frontier, 24, opt.Navigation)
		if !ok {
			result.Status = MineExhausted
			return result, ErrSearchExhausted
		}
		before := loadedChunkSet(m.bot.World.LoadedChunks())
		m.onProgress.emit(MiningProgress{Kind: "exploring", Position: step, Completed: result.Completed, Requested: count})
		exploreNav := opt.Navigation
		// Exploration is deliberately incremental. A modest node budget keeps a
		// distant loaded-chunk boundary from freezing the bot in one huge plan.
		if exploreNav.MaxNodes <= 0 || exploreNav.MaxNodes > 6000 {
			exploreNav.MaxNodes = 6000
		}
		_, err := m.bot.Navigator.Navigate(ctx, GoalNear{step, 2}, exploreNav)
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
		added := newLoadedChunks(before, m.bot.World.LoadedChunks())
		if added == 0 {
			stagnantExploration++
			if stagnantExploration >= 8 {
				result.Status = MineExhausted
				return result, ErrSearchExhausted
			}
		} else {
			stagnantExploration = 0
			result.ExploredChunks += added
		}
	}
	result.Status = MineComplete
	return result, nil
}

func miningTargetScore(self Vec3, target Block, cluster []BlockPos) float64 {
	position := Vec3{float64(target.Position.X) + .5, float64(target.Position.Y) + .5, float64(target.Position.Z) + .5}
	score := self.Distance(position)
	// Stay on the current connected tree/vein when it is still nearby. This
	// prevents alternating between equally distant trunks as their vertical
	// blocks reorder by straight-line distance.
	if len(cluster) > 0 {
		last := cluster[len(cluster)-1]
		if abs(last.X-target.Position.X) <= 1 && abs(last.Z-target.Position.Z) <= 1 && abs(last.Y-target.Position.Y) <= 4 {
			score -= 4
		}
	}
	return score
}

func loadedChunkSet(chunks [][2]int32) map[[2]int32]struct{} {
	set := make(map[[2]int32]struct{}, len(chunks))
	for _, chunk := range chunks {
		set[chunk] = struct{}{}
	}
	return set
}

func newLoadedChunks(before map[[2]int32]struct{}, after [][2]int32) int {
	added := 0
	for _, chunk := range after {
		if _, existed := before[chunk]; !existed {
			added++
		}
	}
	return added
}

// explorationStep turns a potentially view-distance-sized frontier route into
// a short walk. Repeating these steps loads terrain naturally as the player
// approaches the edge of the current world view.
func (m *Miner) explorationStep(from, frontier BlockPos, limit float64, nav NavigationOptions) (BlockPos, bool) {
	dx, dz := float64(frontier.X-from.X), float64(frontier.Z-from.Z)
	distanceXZ := math.Hypot(dx, dz)
	if distanceXZ > limit {
		dx *= limit / distanceXZ
		dz *= limit / distanceXZ
	}
	x, z := from.X+int(math.Round(dx)), from.Z+int(math.Round(dz))
	for offset := 0; offset <= 32; offset++ {
		for _, y := range []int{from.Y + offset, from.Y - offset} {
			candidate := BlockPos{x, y, z}
			if _, _, ok := m.bot.Navigator.passable(candidate, nav); ok {
				return candidate, true
			}
		}
	}
	return BlockPos{}, false
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
		for sectionIndex, section := range col.Sections {
			if section == nil || section.BlockStates == nil {
				continue
			}
			sectionMatches := false
			for _, id := range section.BlockStates.Values() {
				cached, known := states[id]
				if !known {
					b, ok := m.bot.block(BlockPos{}, id)
					cached = cachedState{block: b, ok: ok, match: ok && names[b.Name]}
					if ok && s.Predicate != nil {
						cached.match = cached.match || s.Predicate(b)
					}
					states[id] = cached
				}
				sectionMatches = sectionMatches || cached.match
			}
			if !sectionMatches {
				continue
			}
			baseY := -64 + sectionIndex*16
			for y := 0; y < 16; y++ {
				for z := 0; z < 16; z++ {
					for x := 0; x < 16; x++ {
						pos := BlockPos{int(key.X)*16 + x, baseY + y, int(key.Z)*16 + z}
						if skip[pos] || m.breakRejected(pos) {
							continue
						}
						id := section.GetBlockState(x, y, z)
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
						if cached.match {
							out = append(out, b)
						}
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
	_, obstructed := m.lineOfSightObstruction(from, to, target)
	return !obstructed
}

func (m *Miner) lineOfSightObstruction(from, to Vec3, target BlockPos) (Block, bool) {
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
			if !ok {
				return Block{Position: p, Hardness: -1}, true
			}
			return b, true
		}
	}
	return Block{}, false
}

func (m *Miner) clearLineOfSight(ctx context.Context, pos BlockPos, nav NavigationOptions) error {
	for range 8 {
		self := m.bot.Self.State()
		eye := Vec3{self.Position.X, self.Position.Y + 1.62, self.Position.Z}
		center := Vec3{float64(pos.X) + .5, float64(pos.Y) + .5, float64(pos.Z) + .5}
		blocker, obstructed := m.lineOfSightObstruction(eye, center, pos)
		if !obstructed {
			return nil
		}
		if !nav.AllowBreaking || blocker.Hardness < 0 || m.breakRejected(blocker.Position) || (nav.BreakFilter != nil && !nav.BreakFilter(blocker)) {
			return fmt.Errorf("minego: line of sight to %v is obstructed by %s at %v", pos, blocker.Name, blocker.Position)
		}
		m.onProgress.emit(MiningProgress{Kind: "clearing", Name: blocker.Name, Position: blocker.Position})
		if _, err := m.Dig(ctx, blocker.Position, DigOptions{}); err != nil {
			return fmt.Errorf("clear %s at %v: %w", blocker.Name, blocker.Position, err)
		}
	}
	return fmt.Errorf("minego: too many obstructions in front of %v", pos)
}

func (m *Miner) collectDrops(ctx context.Context, mined []BlockPos, nav NavigationOptions) {
	if len(mined) == 0 {
		return
	}
	select {
	case <-ctx.Done():
		return
	case <-time.After(150 * time.Millisecond):
	}
	if nav.MaxNodes <= 0 || nav.MaxNodes > 2000 {
		nav.MaxNodes = 2000
	}
	for attempts := 0; attempts < 3; attempts++ {
		drop, found := m.nearestBatchDrop(mined, 8)
		if !found {
			return
		}
		m.onProgress.emit(MiningProgress{Kind: "collecting", Position: drop.Position.Block()})
		if _, err := m.bot.Navigator.Navigate(ctx, GoalNear{Position: drop.Position.Block(), Radius: 1.25}, nav); err != nil {
			m.onProgress.emit(MiningProgress{Kind: "drop-unreachable", Position: drop.Position.Block(), Err: err})
			return
		}
		deadline := time.NewTimer(350 * time.Millisecond)
		select {
		case <-ctx.Done():
			deadline.Stop()
			return
		case <-deadline.C:
		}
	}
}

func (m *Miner) nearestBatchDrop(mined []BlockPos, radius float64) (Entity, bool) {
	self := m.bot.Self.State().Position
	best := math.MaxFloat64
	var result Entity
	found := false
	for _, entity := range m.bot.Entities.All() {
		if entity.Type != "minecraft:item" {
			continue
		}
		nearBatch := false
		for _, pos := range mined {
			origin := Vec3{float64(pos.X) + .5, float64(pos.Y) + .5, float64(pos.Z) + .5}
			if entity.Position.Distance(origin) < radius {
				nearBatch = true
				break
			}
		}
		if nearBatch {
			if distance := entity.Position.Distance(self); distance < best {
				best, result, found = distance, entity, true
			}
		}
	}
	return result, found
}

func (m *Miner) nearbyDrop(mined BlockPos, radius float64, skip map[int32]bool) (Entity, bool) {
	origin := Vec3{float64(mined.X) + .5, float64(mined.Y) + .5, float64(mined.Z) + .5}
	best := radius
	var result Entity
	found := false
	for _, entity := range m.bot.Entities.All() {
		if entity.Type != "minecraft:item" || skip[entity.ID] {
			continue
		}
		if d := entity.Position.Distance(origin); d < best {
			best, result, found = d, entity, true
		}
	}
	return result, found
}

func (m *Miner) collectDrop(ctx context.Context, drop Entity, nav NavigationOptions) {
	// Item pickup is a local best-effort task. Keep an awkward drop under a
	// canopy or down a ravine from monopolizing the miner with a full 30k-node
	// search before it can continue to the next block.
	if nav.MaxNodes <= 0 || nav.MaxNodes > 4000 {
		nav.MaxNodes = 4000
	}
	for range 6 {
		current, exists := m.bot.Entities.Get(drop.ID)
		if !exists {
			return
		}
		drop = current
		m.onProgress.emit(MiningProgress{Kind: "collecting", Position: current.Position.Block()})
		if _, err := m.bot.Navigator.Navigate(ctx, GoalNear{Position: current.Position.Block(), Radius: 1.25}, nav); err != nil {
			m.onProgress.emit(MiningProgress{Kind: "drop-unreachable", Position: current.Position.Block(), Err: err})
			select {
			case <-ctx.Done():
				return
			case <-time.After(250 * time.Millisecond):
				continue
			}
		}
		deadline := time.NewTimer(750 * time.Millisecond)
		for {
			if next, exists := m.bot.Entities.Get(drop.ID); !exists {
				deadline.Stop()
				return
			} else {
				drop = next
			}
			select {
			case <-ctx.Done():
				deadline.Stop()
				return
			case <-deadline.C:
				goto retry
			case <-time.After(50 * time.Millisecond):
			}
		}
	retry:
	}
	m.onProgress.emit(MiningProgress{Kind: "drop-uncollected", Position: drop.Position.Block()})
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
	// The server applies the vanilla five-times mining penalty while airborne.
	// This commonly matters for the first block after login, before the client
	// has sent its first grounded movement tick, and while jump-pillaring.
	if !s.OnGround && !s.Flying {
		seconds *= 5
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
