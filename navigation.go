package minego

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

type Goal interface {
	Reached(BlockPos) bool
	Estimate(BlockPos) float64
}

type FollowTarget struct {
	EntityID   int32
	PlayerUUID string
	PlayerName string
}

type FollowOptions struct {
	Distance       float64
	RepathDistance float64
	LostTimeout    time.Duration
	Navigation     NavigationOptions
}

type ExploreOptions struct {
	Origin     BlockPos
	Radius     int
	ChunkLimit int
	Navigation NavigationOptions
}

type ExploreResult struct {
	Chunks    int
	Frontiers []BlockPos
}
type GoalBlock BlockPos

func (g GoalBlock) Reached(p BlockPos) bool     { return BlockPos(g) == p }
func (g GoalBlock) Estimate(p BlockPos) float64 { return distance(BlockPos(g), p) }

type GoalNear struct {
	Position BlockPos
	Radius   float64
}

func (g GoalNear) Reached(p BlockPos) bool     { return distance(g.Position, p) <= g.Radius }
func (g GoalNear) Estimate(p BlockPos) float64 { return math.Max(0, distance(g.Position, p)-g.Radius) }

type GoalAdjacent BlockPos

func (g GoalAdjacent) Reached(p BlockPos) bool {
	q := BlockPos(g)
	return abs(p.X-q.X)+abs(p.Y-q.Y)+abs(p.Z-q.Z) == 1
}
func (g GoalAdjacent) Estimate(p BlockPos) float64 { return math.Max(0, distance(BlockPos(g), p)-1) }

type NavigationOptions struct {
	MaxNodes      int
	MaxDrop       int
	Sprint        bool
	AllowParkour  bool
	MaxParkourGap int
	AllowBreaking bool
	// BreakFilter optionally restricts blocks destroyed to execute a route.
	// Explicit Miner targets are unaffected.
	BreakFilter  func(Block) bool
	AllowPlacing bool
	// TemporaryBlocks lists block items that may be used for route bridges.
	// Names without a namespace are treated as minecraft names.
	TemporaryBlocks []string
	// AcquireTemporary permits mining a TemporaryBlock when the inventory is
	// short before route execution.
	AcquireTemporary bool
	Avoid            []string
}
type MoveKind uint8

const (
	MoveWalk MoveKind = iota
	MoveJump
	MoveDrop
	MoveSwim
	MoveClimb
	MoveDoor
	MoveBreak
	MoveParkour
	MoveBridge
	MovePillar
)

type PathNode struct {
	Position BlockPos
	Move     MoveKind
	Cost     float64
	Break    []BlockPos
	Place    []BlockPos
}
type PathProgress struct {
	Index, Total int
	Position     Vec3
	Replanned    bool
	Target       BlockPos
	Move         MoveKind
}
type NavigationResult struct {
	Path     []PathNode
	Replans  int
	Repairs  int
	Segments int
	Expanded int
}

type navRun struct {
	ctx          context.Context
	goal         Goal
	options      NavigationOptions
	path         []PathNode
	index        int
	done         chan error
	result       NavigationResult
	dirty        bool
	lastDistance float64
	stalled      int
	actionIndex  int
	digging      bool
	pillaring    bool
	replanned    bool
	planner      *dstarPlanner
	changes      []BlockPos
	jumpIndex    int
	expected     map[BlockPos]bool
}
type Navigator struct {
	bot                 *Bot
	mu                  sync.Mutex
	active              *navRun
	sprinting           bool
	horizontalCollision bool
	blockedX, blockedZ  bool
	onProgress          event[PathProgress]
}

func newNavigator(b *Bot) *Navigator {
	n := &Navigator{bot: b}
	b.World.OnBlockChange(func(change BlockChange) {
		n.mu.Lock()
		if n.active != nil {
			if n.active.expected[change.Position] {
				delete(n.active.expected, change.Position)
			} else if routeAffected(n.active.path, n.active.index, change.Position) {
				n.active.dirty = true
				n.active.result.Repairs++
			}
		}
		n.mu.Unlock()
	})
	return n
}

// routeAffected deliberately considers only cells used by a remaining edge.
// Tree decay, item landings, and building updates near (but not on) a route
// must not turn every world packet into a full path search.
func routeAffected(path []PathNode, start int, changed BlockPos) bool {
	for i := max(0, start-1); i < len(path); i++ {
		node := path[i]
		if changed == node.Position || changed == (BlockPos{node.Position.X, node.Position.Y + 1, node.Position.Z}) || changed == (BlockPos{node.Position.X, node.Position.Y - 1, node.Position.Z}) {
			return true
		}
		for _, pos := range node.Break {
			if changed == pos {
				return true
			}
		}
		for _, pos := range node.Place {
			if changed == pos {
				return true
			}
		}
	}
	return false
}
func (n *Navigator) OnProgress(fn func(PathProgress)) func() { return n.onProgress.subscribe(fn) }
func (n *Navigator) Stop()                                   { n.finish(context.Canceled) }
func (n *Navigator) Path(start BlockPos, goal Goal, opt NavigationOptions) ([]PathNode, error) {
	return n.findPath(start, goal, defaultsNav(opt))
}

// Follow keeps the bot within Distance of a moving entity until ctx is
// cancelled. Player names and UUIDs are resolved through Bot.Players.
func (n *Navigator) Follow(ctx context.Context, target FollowTarget, opt FollowOptions) error {
	if opt.Distance <= 0 {
		opt.Distance = 2
	}
	if opt.RepathDistance <= 0 {
		opt.RepathDistance = 1.5
	}
	if opt.LostTimeout <= 0 {
		opt.LostTimeout = 10 * time.Second
	}
	changes := make(chan struct{}, 1)
	unsub := n.bot.Entities.OnChange(func(EntityChange) {
		select {
		case changes <- struct{}{}:
		default:
		}
	})
	defer unsub()
	var lostSince time.Time
	for {
		entity, ok := n.followEntity(target)
		if !ok {
			if lostSince.IsZero() {
				lostSince = time.Now()
			}
			if time.Since(lostSince) >= opt.LostTimeout {
				return ErrTargetLost
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-changes:
				continue
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}
		lostSince = time.Time{}
		if _, err := n.Navigate(ctx, GoalNear{Position: entity.Position.Block(), Radius: opt.Distance}, opt.Navigation); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		start := entity.Position
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-changes:
				current, found := n.followEntity(target)
				if !found || current.Position.Distance(start) >= opt.RepathDistance {
					goto replan
				}
			}
		}
	replan:
	}
}

func (n *Navigator) followEntity(target FollowTarget) (Entity, bool) {
	if target.PlayerName != "" || target.PlayerUUID != "" {
		if n.bot.Players != nil {
			player, ok := n.bot.Players.find(target.PlayerName, target.PlayerUUID)
			if ok && player.EntityID != 0 {
				return n.bot.Entities.Get(player.EntityID)
			}
		}
		if target.PlayerUUID != "" {
			for _, entity := range n.bot.Entities.All() {
				if entity.UUID == target.PlayerUUID {
					return entity, true
				}
			}
		}
		return Entity{}, false
	}
	return n.bot.Entities.Get(target.EntityID)
}

// Explore walks loaded terrain frontiers until ChunkLimit new chunks have
// appeared, the radius is exhausted, or ctx is cancelled.
func (n *Navigator) Explore(ctx context.Context, opt ExploreOptions) (ExploreResult, error) {
	if opt.Radius <= 0 {
		opt.Radius = 256
	}
	if opt.ChunkLimit <= 0 {
		opt.ChunkLimit = 1
	}
	if opt.Origin == (BlockPos{}) {
		opt.Origin = n.bot.Self.State().Position.Block()
	}
	result := ExploreResult{}
	for result.Chunks < opt.ChunkLimit {
		frontier, ok := n.bot.Miner.frontier(opt.Origin, opt.Radius)
		if !ok {
			return result, ErrSearchExhausted
		}
		before := len(n.bot.World.LoadedChunks())
		if _, err := n.Navigate(ctx, GoalNear{Position: frontier, Radius: 2}, opt.Navigation); err != nil {
			return result, err
		}
		deadline := time.NewTimer(3 * time.Second)
		for len(n.bot.World.LoadedChunks()) <= before {
			select {
			case <-ctx.Done():
				deadline.Stop()
				return result, ctx.Err()
			case <-deadline.C:
				return result, ErrSearchExhausted
			case <-time.After(50 * time.Millisecond):
			}
		}
		deadline.Stop()
		result.Chunks += len(n.bot.World.LoadedChunks()) - before
		result.Frontiers = append(result.Frontiers, frontier)
	}
	return result, nil
}
func (n *Navigator) Navigate(ctx context.Context, goal Goal, opt NavigationOptions) (NavigationResult, error) {
	start := n.bot.Self.State().Position.Block()
	if goal.Reached(start) {
		return NavigationResult{Path: []PathNode{{Position: start}}}, nil
	}
	lease, err := n.bot.actions.acquire(ctx, controlMovement|controlView|controlHands|controlInventory, priorityAutomation)
	if err != nil {
		return NavigationResult{}, err
	}
	defer lease.Release()
	ctx = lease.Context(ctx)
	opt = defaultsNav(opt)
	path, segmented, expanded, planner, err := n.planRoute(start, goal, opt)
	if err != nil {
		return NavigationResult{}, err
	}
	path, segmented, expanded, planner = n.preferAvailableRoute(start, goal, opt, path, segmented, expanded, planner)
	if err := n.ensureTemporary(ctx, path, opt); err != nil {
		return NavigationResult{}, err
	}
	// Acquiring route material may move the bot and change the graph.
	newStart := n.bot.Self.State().Position.Block()
	expandedAgain := 0
	if newStart != start {
		path, segmented, expandedAgain, planner, err = n.planRoute(newStart, goal, opt)
		if err != nil {
			return NavigationResult{}, err
		}
	}
	result := NavigationResult{Path: path, Expanded: expanded + expandedAgain}
	if segmented {
		result.Segments = 1
	}
	run := &navRun{ctx: ctx, goal: goal, options: opt, path: path, index: 1, done: make(chan error, 1), result: result, lastDistance: math.MaxFloat64, jumpIndex: -1, planner: planner, expected: make(map[BlockPos]bool)}
	n.mu.Lock()
	if n.active != nil {
		n.active.done <- context.Canceled
	}
	n.active = run
	n.mu.Unlock()
	if run.index < len(run.path) {
		n.onProgress.emit(PathProgress{Index: run.index, Total: len(run.path), Position: n.bot.Self.State().Position, Target: run.path[run.index].Position, Move: run.path[run.index].Move})
	}
	select {
	case <-ctx.Done():
		n.finish(ctx.Err())
		return run.result, ctx.Err()
	case err := <-run.done:
		return run.result, err
	case <-n.bot.done:
		return run.result, ErrNotConnected
	}
}

func (n *Navigator) preferAvailableRoute(start BlockPos, goal Goal, opt NavigationOptions, path []PathNode, segmented bool, expanded int, planner *dstarPlanner) ([]PathNode, bool, int, *dstarPlanner) {
	missing := pathTemporaryCount(path) - n.availableTemporary(opt)
	if missing <= 0 || !opt.AllowPlacing {
		return path, segmented, expanded, planner
	}
	withoutPlacement := opt
	withoutPlacement.AllowPlacing = false
	alternative, altSegmented, altExpanded, _, err := n.planRoute(start, goal, withoutPlacement)
	expanded += altExpanded
	if err != nil {
		return path, segmented, expanded, planner
	}
	// Missing route blocks require another search, dig, pickup, and inventory
	// operation. Charge that real overhead before deciding that a pillar or
	// bridge is the faster route.
	if pathCost(alternative) <= pathCost(path)+float64(missing)*8 {
		return alternative, altSegmented, expanded, nil
	}
	return path, segmented, expanded, planner
}

func pathCost(path []PathNode) float64 {
	if len(path) == 0 {
		return math.Inf(1)
	}
	return path[len(path)-1].Cost
}
func defaultsNav(o NavigationOptions) NavigationOptions {
	if o.MaxNodes <= 0 {
		o.MaxNodes = 30000
	}
	if o.MaxDrop <= 0 {
		o.MaxDrop = 3
	}
	if o.AllowParkour && o.MaxParkourGap <= 0 {
		o.MaxParkourGap = 2
	}
	if o.AllowPlacing && len(o.TemporaryBlocks) == 0 {
		o.TemporaryBlocks = []string{"minecraft:dirt", "minecraft:cobblestone"}
	}
	for i, name := range o.TemporaryBlocks {
		if !strings.Contains(name, ":") {
			o.TemporaryBlocks[i] = "minecraft:" + name
		}
	}
	return o
}

func (n *Navigator) ensureTemporary(ctx context.Context, path []PathNode, opt NavigationOptions) error {
	need := pathTemporaryCount(path)
	if need == 0 {
		return nil
	}
	have := 0
	allowed := make(map[string]bool, len(opt.TemporaryBlocks))
	for _, name := range opt.TemporaryBlocks {
		allowed[name] = true
	}
	for _, stack := range n.bot.Inventory.Slots() {
		if allowed[stack.Name] {
			have += int(stack.Count)
		}
	}
	if have >= need {
		return nil
	}
	if !opt.AcquireTemporary {
		return fmt.Errorf("%w: need %d route blocks, have %d", ErrNoTemporaryBlocks, need, have)
	}
	_, err := n.bot.Miner.Mine(ctx, Blocks(opt.TemporaryBlocks...), need-have, MineOptions{Navigation: NavigationOptions{AllowBreaking: opt.AllowBreaking}})
	if err != nil {
		return fmt.Errorf("acquire route blocks: %w", err)
	}
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	for {
		collected := 0
		for _, stack := range n.bot.Inventory.Slots() {
			if allowed[stack.Name] {
				collected += int(stack.Count)
			}
		}
		if collected >= need {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("%w: mined route material was not collected", ErrNoTemporaryBlocks)
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func pathTemporaryCount(path []PathNode) int {
	need := 0
	for _, node := range path {
		need += len(node.Place)
	}
	return need
}

func (n *Navigator) availableTemporary(opt NavigationOptions) int {
	allowed := make(map[string]bool, len(opt.TemporaryBlocks))
	for _, name := range opt.TemporaryBlocks {
		allowed[name] = true
	}
	have := 0
	for _, stack := range n.bot.Inventory.Slots() {
		if allowed[stack.Name] {
			have += int(stack.Count)
		}
	}
	return have
}
func (n *Navigator) finish(err error) {
	n.mu.Lock()
	r := n.active
	n.active = nil
	n.mu.Unlock()
	if r != nil {
		select {
		case r.done <- err:
		default:
		}
	}
}

func (n *Navigator) tick() {
	if n.bot.respawning.Load() {
		return
	}
	n.mu.Lock()
	r := n.active
	if r == nil {
		n.mu.Unlock()
		n.move(physicsInput{}, n.bot.Self.State())
		return
	}
	select {
	case <-r.ctx.Done():
		n.active = nil
		n.mu.Unlock()
		r.done <- r.ctx.Err()
		return
	default:
	}
	state := n.bot.Self.State()
	at := state.Position.Block()
	if r.goal.Reached(at) {
		n.active = nil
		n.mu.Unlock()
		r.done <- nil
		n.move(physicsInput{}, state)
		return
	}
	if r.dirty || r.index >= len(r.path) {
		var path []PathNode
		var segmented bool
		var expanded int
		var planner *dstarPlanner
		var err error
		path, segmented, expanded, planner, err = n.planRoute(at, r.goal, r.options)
		// A live replan can introduce a bridge or pillar that was absent from
		// the original route. Acquiring material here would recursively replace
		// the active navigation run, so prefer a jump/walk detour whenever the
		// inventory cannot execute the newly planned placements.
		if err == nil && pathTemporaryCount(path) > n.availableTemporary(r.options) {
			withoutPlacement := r.options
			withoutPlacement.AllowPlacing = false
			var extra int
			path, segmented, extra, planner, err = n.planRoute(at, r.goal, withoutPlacement)
			expanded += extra
		}
		if err != nil {
			n.active = nil
			n.mu.Unlock()
			r.done <- err
			return
		}
		r.path = path
		r.result.Path = path
		r.index = 1
		r.actionIndex = 0
		r.jumpIndex = -1
		r.result.Replans++
		r.result.Expanded += expanded
		r.replanned = true
		if segmented {
			r.result.Segments++
		}
		r.dirty = false
		r.changes = nil
		r.planner = planner
	}
	if r.index >= len(r.path) {
		n.mu.Unlock()
		return
	}
	target := r.path[r.index]
	if len(target.Break) > 0 && !r.digging {
		r.digging = true
		idx := r.index
		breaks := append([]BlockPos(nil), target.Break...)
		for _, pos := range breaks {
			r.expected[pos] = true
		}
		n.mu.Unlock()
		go func() {
			var err error
			for _, pos := range breaks {
				if _, err = n.bot.Miner.Dig(r.ctx, pos, DigOptions{}); err != nil {
					break
				}
			}
			n.mu.Lock()
			defer n.mu.Unlock()
			if n.active != r {
				return
			}
			r.digging = false
			if err != nil {
				if errors.Is(err, ErrBlockBreakRejected) && r.index == idx {
					r.dirty = true
					return
				}
				n.active = nil
				r.done <- err
				return
			}
			if r.index == idx {
				r.path[idx].Break = nil
			}
		}()
		n.move(physicsInput{}, state)
		return
	}
	if r.digging {
		n.mu.Unlock()
		n.move(physicsInput{Jump: r.pillaring}, state)
		return
	}
	if len(target.Place) > 0 && r.actionIndex != r.index {
		r.actionIndex = r.index
		r.digging = true
		r.pillaring = target.Move == MovePillar
		places := append([]BlockPos(nil), target.Place...)
		for _, pos := range places {
			r.expected[pos] = true
		}
		n.mu.Unlock()
		go func() {
			var err error
			if target.Move == MovePillar {
				clearance := time.NewTimer(time.Second)
				defer clearance.Stop()
				for n.bot.Self.State().Position.Y < float64(places[0].Y)+1.01 {
					select {
					case <-r.ctx.Done():
						err = r.ctx.Err()
					case <-clearance.C:
						err = fmt.Errorf("minego: pillar jump did not clear placement cell")
					case <-time.After(25 * time.Millisecond):
					}
					if err != nil {
						break
					}
				}
			}
			for _, pos := range places {
				if err != nil {
					break
				}
				item := ""
				for _, candidate := range r.options.TemporaryBlocks {
					if e := n.bot.Builder.ensureHotbar(r.ctx, candidate); e == nil {
						item = candidate
						break
					}
				}
				if item == "" {
					err = ErrNoTemporaryBlocks
					break
				}
				if _, err = n.bot.Builder.Place(r.ctx, pos, PlaceOptions{Item: item}); err != nil {
					break
				}
			}
			n.mu.Lock()
			defer n.mu.Unlock()
			if n.active != r {
				return
			}
			r.digging = false
			r.pillaring = false
			if err != nil {
				n.active = nil
				r.done <- err
				return
			}
		}()
		n.move(physicsInput{}, state)
		return
	}
	if target.Move == MoveDoor && r.actionIndex != r.index {
		r.actionIndex = r.index
		seq := n.bot.nextSequence()
		_ = n.bot.send(context.Background(), &packets.C2SUseItemOn{Hand: 0, Location: ns.NewPosition(target.Position.X, target.Position.Y, target.Position.Z), Face: 1, CursorPositionX: .5, CursorPositionY: .5, CursorPositionZ: .5, Sequence: ns.VarInt(seq)})
	}
	dx := float64(target.Position.X) + .5 - state.Position.X
	dz := float64(target.Position.Z) + .5 - state.Position.Z
	dist := math.Hypot(dx, dz)
	if at == target.Position || dist < .30 && math.Abs(state.Position.Y-float64(target.Position.Y)) < .55 || waypointPassed(r, state.Position) {
		r.index++
		n.mu.Unlock()
		progress := PathProgress{Index: r.index, Total: len(r.path), Position: state.Position, Replanned: r.replanned}
		r.replanned = false
		if r.index < len(r.path) {
			progress.Target, progress.Move = r.path[r.index].Position, r.path[r.index].Move
		}
		n.onProgress.emit(progress)
		n.move(physicsInput{}, state)
		return
	}
	if math.Abs(dist-r.lastDistance) < .002 {
		r.stalled++
	} else {
		r.stalled = 0
	}
	r.lastDistance = dist
	if r.stalled > 30 {
		r.dirty = true
		r.stalled = 0
	}
	speed := .215
	if r.options.Sprint {
		speed = .28
	}
	if dist > 0 {
		dx = dx / dist * math.Min(speed, dist)
		dz = dz / dist * math.Min(speed, dist)
	}
	input := physicsInput{X: dx, Z: dz, LookX: dx, LookZ: dz}
	input.Sprint = r.options.Sprint
	// A server correction can leave the player a few millimetres across the
	// corner of a collision shape. Retrying the same diagonal vector keeps both
	// components pinned. Back away to the previous cell center for one or more
	// ticks, then resume the diagonal from clean clearance.
	if n.horizontalCollision && r.index > 0 && target.Move == MoveWalk {
		previous := r.path[r.index-1].Position
		if recovery, recoverable := collisionRecoveryInput(state, previous, target.Position, n.blockedX, n.blockedZ); recoverable {
			input = recovery
			input.Sprint = false
		} else {
			r.dirty = true
		}
	}
	if target.Move == MoveJump && target.Position.Y > at.Y {
		centered := true
		if n.grounded(state) && r.index > 0 {
			centering, atCenter := centerNodeInput(state, r.path[r.index-1].Position, speed)
			centered = atCenter
			if !centered {
				input = centering
				input.Sprint = false
			}
		}
		if centered {
			input.Jump = true
			// Clear the obstacle vertically before translating across its upper
			// edge. Sending a simultaneous edge-touching rise is repeatedly
			// corrected by Paper and leaves the player frozen at jump-frame one.
			if state.Position.Y < float64(target.Position.Y)-0.05 {
				input.X, input.Z = 0, 0
			}
		}
	}
	if target.Move == MoveSwim {
		input.Jump = true
	}
	if target.Move == MoveClimb {
		input.Climb = math.Max(-.15, math.Min(.2, float64(target.Position.Y)-state.Position.Y))
	}
	if target.Move == MoveParkour {
		input.Sprint = true
		if r.jumpIndex != r.index && (state.OnGround || n.hasSupport(playerBox(state.Position))) {
			input.Jump = true
			r.jumpIndex = r.index
		}
	}
	n.mu.Unlock()
	n.move(input, state)
}

func collisionRecoveryInput(state SelfState, previous, target BlockPos, blockedX, blockedZ bool) (physicsInput, bool) {
	dx, dz := target.X-previous.X, target.Z-previous.Z
	if dx != 0 && dz != 0 {
		input, centered := centerNodeInput(state, previous, .18)
		return input, !centered
	}
	if dx == 0 && blockedX && !blockedZ {
		lateral := float64(previous.X) + .5 - state.Position.X
		if math.Abs(lateral) <= .01 {
			return physicsInput{}, false
		}
		lateral = math.Copysign(math.Min(.18, math.Abs(lateral)), lateral)
		return physicsInput{X: lateral, LookX: lateral}, true
	}
	if dz == 0 && blockedZ && !blockedX {
		lateral := float64(previous.Z) + .5 - state.Position.Z
		if math.Abs(lateral) <= .01 {
			return physicsInput{}, false
		}
		lateral = math.Copysign(math.Min(.18, math.Abs(lateral)), lateral)
		return physicsInput{Z: lateral, LookZ: lateral}, true
	}
	return physicsInput{}, false
}

func waypointPassed(r *navRun, position Vec3) bool {
	if r.index+1 >= len(r.path) || r.path[r.index].Move != MoveWalk || r.path[r.index+1].Move != MoveWalk {
		return false
	}
	current := r.path[r.index].Position
	next := r.path[r.index+1].Position
	here := Vec3{float64(current.X) + .5, float64(current.Y), float64(current.Z) + .5}
	there := Vec3{float64(next.X) + .5, float64(next.Y), float64(next.Z) + .5}
	return position.Distance(there)+.05 < position.Distance(here)
}

func centerNodeInput(state SelfState, node BlockPos, speed float64) (physicsInput, bool) {
	dx := float64(node.X) + .5 - state.Position.X
	dz := float64(node.Z) + .5 - state.Position.Z
	dist := math.Hypot(dx, dz)
	if dist <= .08 {
		return physicsInput{}, true
	}
	step := math.Min(speed, dist)
	dx, dz = dx/dist*step, dz/dist*step
	return physicsInput{X: dx, Z: dz, LookX: dx, LookZ: dz}, false
}

func (n *Navigator) move(input physicsInput, state SelfState) {
	grounded := n.grounded(state)
	input, yaw, pitch := orientMovement(input, state, grounded)
	result := n.physicsStep(state, input)
	n.horizontalCollision = result.HorizontalCollision
	n.blockedX, n.blockedZ = result.BlockedX, result.BlockedZ
	flags := ns.Int8(0)
	if result.OnGround {
		flags |= 1
	}
	if result.HorizontalCollision {
		flags |= 2
	}
	inputFlags := ns.Uint8(0)
	if input.X != 0 || input.Z != 0 {
		inputFlags |= 1 // forward (yaw already faces the desired world vector)
	}
	if input.Jump || input.Climb > 0 {
		inputFlags |= 1 << 4
	}
	if input.Sprint {
		inputFlags |= 1 << 6
	}
	if input.Sprint != n.sprinting {
		action := int32(4) // stop sprinting
		if input.Sprint {
			action = 3
		}
		if err := n.bot.send(context.Background(), &packets.C2SPlayerCommand{EntityId: ns.VarInt(state.EntityID), ActionId: ns.VarInt(action)}); err != nil {
			return
		}
		n.sprinting = input.Sprint
	}
	if err := n.bot.send(context.Background(), &packets.C2SPlayerInput{Flags: inputFlags}); err != nil {
		return
	}
	if err := n.bot.send(context.Background(), &packets.C2SMovePlayerPosRot{X: ns.Float64(result.Position.X), FeetY: ns.Float64(result.Position.Y), Z: ns.Float64(result.Position.Z), Yaw: ns.Float32(yaw), Pitch: ns.Float32(pitch), Flags: flags}); err != nil {
		return
	}
	n.bot.Self.update(func(s *SelfState) {
		s.Position, s.Velocity, s.OnGround = result.Position, result.Velocity, result.OnGround
		s.Rotation = Rotation{Yaw: yaw, Pitch: pitch}
	})
}

func (n *Navigator) grounded(state SelfState) bool {
	return state.OnGround || n.hasSupport(playerBox(state.Position))
}

func orientMovement(input physicsInput, state SelfState, grounded bool) (physicsInput, float32, float32) {
	yaw := state.Rotation.Yaw
	pitch := state.Rotation.Pitch
	lookX, lookZ := input.LookX, input.LookZ
	if lookX == 0 && lookZ == 0 {
		lookX, lookZ = input.X, input.Z
	}
	if lookX != 0 || lookZ != 0 {
		desiredYaw := float32(math.Atan2(-lookX, lookZ) * 180 / math.Pi)
		// Turn over several client ticks and bring the view back toward the
		// horizon. This keeps movement camera-driven without the instantaneous
		// head snaps produced by assigning the destination angle directly.
		yaw = approachAngle(yaw, desiredYaw, 30)
		pitch = approach(pitch, 0, 12)
		// Modern servers validate the forward-input bit against the reported
		// camera direction. Turn in place first when the requested direction is
		// far outside the current view instead of sending sideways coordinates
		// while claiming to move forward.
		if math.Abs(float64(angleDelta(state.Rotation.Yaw, desiredYaw))) > 5 {
			input.X, input.Z = 0, 0
			// Do not spend the useful airborne part of an obstacle jump turning
			// in place. Face the target while grounded, then take off.
			if grounded {
				input.Jump = false
			}
		}
	}
	return input, yaw, pitch
}

func approach(value, target, limit float32) float32 {
	delta := target - value
	if delta > limit {
		delta = limit
	} else if delta < -limit {
		delta = -limit
	}
	return value + delta
}

func approachAngle(value, target, limit float32) float32 {
	return value + approach(0, angleDelta(value, target), limit)
}

func angleDelta(value, target float32) float32 {
	delta := float32(math.Mod(float64(target-value+180), 360))
	if delta < 0 {
		delta += 360
	}
	return delta - 180
}

type pathEntry struct {
	pos    BlockPos
	g, f   float64
	move   MoveKind
	parent *pathEntry
	index  int
}
type pathHeap []*pathEntry

func (h pathHeap) Len() int           { return len(h) }
func (h pathHeap) Less(i, j int) bool { return h[i].f < h[j].f }
func (h pathHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *pathHeap) Push(x any)        { *h = append(*h, x.(*pathEntry)) }
func (h *pathHeap) Pop() any          { o := *h; x := o[len(o)-1]; *h = o[:len(o)-1]; return x }
func (n *Navigator) findPath(start BlockPos, goal Goal, opt NavigationOptions) ([]PathNode, error) {
	path, _, err := n.findPathCount(start, goal, opt)
	return path, err
}
func (n *Navigator) findPathCount(start BlockPos, goal Goal, opt NavigationOptions) ([]PathNode, int, error) {
	open := &pathHeap{}
	heap.Init(open)
	root := &pathEntry{pos: start, f: goal.Estimate(start)}
	heap.Push(open, root)
	best := map[BlockPos]float64{start: 0}
	explored := 0
	for open.Len() > 0 {
		cur := heap.Pop(open).(*pathEntry)
		if goal.Reached(cur.pos) {
			return n.buildPath(cur), explored, nil
		}
		explored++
		if explored > opt.MaxNodes {
			return nil, explored, fmt.Errorf("%w: node limit %d", ErrUnreachable, opt.MaxNodes)
		}
		for _, next := range n.neighbors(cur.pos, opt) {
			g := cur.g + next.Cost
			if old, ok := best[next.Position]; ok && old <= g {
				continue
			}
			best[next.Position] = g
			heap.Push(open, &pathEntry{pos: next.Position, g: g, f: g + goal.Estimate(next.Position), move: next.Move, parent: cur})
		}
	}
	return nil, explored, ErrUnreachable
}

// planPath falls back to a safe loaded-chunk frontier when the final goal is
// outside the known graph. Execution replans from that frontier as chunks
// arrive, bounding search memory and latency for long trips.
func (n *Navigator) planPath(start BlockPos, goal Goal, opt NavigationOptions) ([]PathNode, bool, int, error) {
	path, expanded, err := n.findPathCount(start, goal, opt)
	if err == nil {
		return path, false, expanded, nil
	}
	frontier, ok := n.segmentFrontier(start, goal)
	if !ok {
		return nil, false, expanded, err
	}
	path, segmentExpanded, segmentErr := n.findPathCount(start, GoalNear{Position: frontier, Radius: 2}, opt)
	if segmentErr != nil {
		return nil, false, expanded + segmentExpanded, err
	}
	return path, true, expanded + segmentExpanded, nil
}
func (n *Navigator) planRoute(start BlockPos, goal Goal, opt NavigationOptions) ([]PathNode, bool, int, *dstarPlanner, error) {
	// A* expands from the player and is substantially cheaper for the short,
	// repeated goals used by mining and building. D*'s reverse predecessor
	// discovery made even tiny initial routes and repairs scan large 3-D areas.
	path, segmented, expanded, err := n.planPath(start, goal, opt)
	return path, segmented, expanded, nil, err
}
func (n *Navigator) segmentFrontier(start BlockPos, goal Goal) (BlockPos, bool) {
	chunks := n.bot.World.LoadedChunks()
	loaded := make(map[[2]int32]bool, len(chunks))
	for _, c := range chunks {
		loaded[c] = true
	}
	best := math.MaxFloat64
	var result BlockPos
	found := false
	for _, c := range chunks {
		for _, d := range [][2]int32{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			if loaded[[2]int32{c[0] + d[0], c[1] + d[1]}] {
				continue
			}
			x, z := int(c[0]*16+8+d[0]*7), int(c[1]*16+8+d[1]*7)
			for y := 319; y >= -63; y-- {
				p := BlockPos{x, y, z}
				if _, _, ok := n.passable(p, NavigationOptions{}); ok {
					score := goal.Estimate(p)
					if distance(start, p) > 3 && score < best {
						best = score
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
func (n *Navigator) buildPath(end *pathEntry) []PathNode {
	var p []PathNode
	for x := end; x != nil; x = x.parent {
		// Rehydrate execution actions that are derived from the selected move.
		// A* entries store only graph costs and move kinds; omitting this step
		// turns MoveBreak/MoveBridge/MovePillar into impossible plain movement.
		p = append(p, n.pathNode(x.pos, x.move, x.g))
	}
	for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
		p[i], p[j] = p[j], p[i]
	}
	return p
}
func (n *Navigator) neighbors(p BlockPos, opt NavigationOptions) []PathNode {
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	out := make([]PathNode, 0, 24)
	for _, d := range dirs {
		diagonal := d[0] != 0 && d[1] != 0
		if diagonal && !n.diagonalClear(p, d[0], d[1]) {
			continue
		}
		multiplier := 1.0
		if diagonal {
			multiplier = math.Sqrt2
		}
		q := BlockPos{p.X + d[0], p.Y, p.Z + d[1]}
		move, sameCost, sameLevel := n.passable(q, opt)
		if sameLevel && move != MoveBreak {
			out = append(out, n.pathNode(q, move, sameCost*multiplier))
			continue
		}
		q.Y++
		jumpClear := n.clear(BlockPos{p.X, p.Y + 2, p.Z})
		if diagonal {
			jumpClear = jumpClear && n.diagonalClear(BlockPos{p.X, p.Y + 1, p.Z}, d[0], d[1])
		}
		if _, cost, ok := n.passable(q, opt); ok && jumpClear {
			out = append(out, n.pathNode(q, MoveJump, (cost+1.5)*multiplier))
			continue
		}
		if sameLevel {
			q.Y--
			out = append(out, n.pathNode(q, move, sameCost*multiplier))
			continue
		}
		for drop := 1; drop <= opt.MaxDrop; drop++ {
			q.Y = p.Y - drop
			// The player first has to move horizontally into the destination
			// column at the departure height. Checking only the eventual landing
			// cells can plan an impossible drop beneath an overhanging log/leaf,
			// where the player's head remains wedged against the upper block.
			departure := BlockPos{q.X, p.Y, q.Z}
			if !n.bodyClear(departure) {
				continue
			}
			if _, cost, ok := n.passable(q, opt); ok {
				out = append(out, n.pathNode(q, MoveDrop, (cost+float64(drop)*.4)*multiplier))
				break
			}
			if !n.clear(q) {
				break
			}
		}
	}
	if opt.AllowParkour {
		for _, d := range dirs[:4] {
			for gap := 1; gap <= opt.MaxParkourGap; gap++ {
				air := BlockPos{p.X + d[0]*gap, p.Y, p.Z + d[1]*gap}
				if !n.bodyClear(air) {
					break
				}
				below, ok := n.bot.World.Block(BlockPos{air.X, air.Y - 1, air.Z})
				if !ok || n.blockSolid(below) {
					break
				}
				landing := BlockPos{p.X + d[0]*(gap+1), p.Y, p.Z + d[1]*(gap+1)}
				if _, cost, ok := n.passable(landing, opt); ok {
					out = append(out, n.pathNode(landing, MoveParkour, cost*float64(gap+1)*.9))
					break
				}
			}
		}
	}
	if n.climbable(p) {
		for _, dy := range []int{-1, 1} {
			q := BlockPos{p.X, p.Y + dy, p.Z}
			if n.clear(q) {
				out = append(out, n.pathNode(q, MoveClimb, 1.5))
			}
		}
	}
	if b, ok := n.bot.World.Block(p); ok && isFluid(b.Name) {
		for _, dy := range []int{-1, 1} {
			q := BlockPos{p.X, p.Y + dy, p.Z}
			if move, cost, ok := n.passable(q, opt); ok && move == MoveSwim {
				out = append(out, n.pathNode(q, MoveSwim, cost+.5))
			}
		}
	}
	if opt.AllowPlacing {
		q := BlockPos{p.X, p.Y + 1, p.Z}
		if n.bodyClear(q) {
			node := n.pathNode(q, MovePillar, 5)
			node.Place = []BlockPos{p}
			out = append(out, node)
		}
	}
	return out
}

func (n *Navigator) pathNode(pos BlockPos, move MoveKind, cost float64) PathNode {
	node := PathNode{Position: pos, Move: move, Cost: cost}
	if move != MoveDoor {
		for _, p := range []BlockPos{pos, {pos.X, pos.Y + 1, pos.Z}} {
			if b, ok := n.bot.World.Block(p); ok && !n.blockClear(b) {
				node.Break = append(node.Break, p)
			}
		}
	}
	if move == MoveBridge {
		node.Place = []BlockPos{{pos.X, pos.Y - 1, pos.Z}}
	}
	if move == MovePillar {
		node.Place = []BlockPos{{pos.X, pos.Y - 1, pos.Z}}
	}
	return node
}

func (n *Navigator) bodyClear(p BlockPos) bool {
	return n.clear(p) && n.clear(BlockPos{p.X, p.Y + 1, p.Z})
}

func (n *Navigator) diagonalClear(p BlockPos, dx, dz int) bool {
	return n.bodyClear(BlockPos{p.X + dx, p.Y, p.Z}) && n.bodyClear(BlockPos{p.X, p.Y, p.Z + dz})
}
func (n *Navigator) passable(p BlockPos, opt NavigationOptions) (MoveKind, float64, bool) {
	feet, ok := n.bot.World.Block(p)
	if !ok {
		return 0, 0, false
	}
	head, ok := n.bot.World.Block(BlockPos{p.X, p.Y + 1, p.Z})
	if !ok {
		return 0, 0, false
	}
	for _, avoid := range opt.Avoid {
		if feet.Name == avoid || head.Name == avoid {
			return 0, 0, false
		}
	}
	fluid := isFluid(feet.Name)
	door := strings.HasSuffix(feet.Name, "_door")
	if !n.blockClear(feet) && !fluid && !door {
		if opt.AllowBreaking && feet.Hardness >= 0 && !n.bot.Miner.breakRejected(p) && (opt.BreakFilter == nil || opt.BreakFilter(feet)) {
			return MoveBreak, 2 + float64(feet.Hardness)*2, true
		}
		return 0, 0, false
	}
	if !n.blockClear(head) && !isFluid(head.Name) {
		if !(opt.AllowBreaking && head.Hardness >= 0 && !n.bot.Miner.breakRejected(BlockPos{p.X, p.Y + 1, p.Z}) && (opt.BreakFilter == nil || opt.BreakFilter(head))) {
			return 0, 0, false
		}
	}
	below, ok := n.bot.World.Block(BlockPos{p.X, p.Y - 1, p.Z})
	if !ok {
		return 0, 0, false
	}
	if !n.blockSolid(below) && !fluid && !n.climbable(p) {
		if opt.AllowPlacing && n.bodyClear(p) {
			return MoveBridge, 4, true
		}
		return 0, 0, false
	}
	if door {
		return MoveDoor, 2, true
	}
	if fluid {
		return MoveSwim, 2.5, true
	}
	return MoveWalk, 1, true
}
func (n *Navigator) clear(p BlockPos) bool {
	b, ok := n.bot.World.Block(p)
	return ok && n.blockClear(b)
}
func (n *Navigator) blockClear(b Block) bool { return len(b.Collision) == 0 }
func (n *Navigator) blockSolid(b Block) bool { return len(b.Collision) > 0 }
func (n *Navigator) climbable(p BlockPos) bool {
	b, ok := n.bot.World.Block(p)
	return ok && (strings.Contains(b.Name, "ladder") || strings.Contains(b.Name, "vine"))
}
func isFluid(name string) bool {
	return name == "minecraft:water" || name == "minecraft:lava" || strings.Contains(name, "bubble_column")
}
func distance(a, b BlockPos) float64 {
	x, y, z := float64(a.X-b.X), float64(a.Y-b.Y), float64(a.Z-b.Z)
	return math.Sqrt(x*x + y*y + z*z)
}
func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
