package minego

import (
	"container/heap"
	"context"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

type Goal interface {
	Reached(BlockPos) bool
	Estimate(BlockPos) float64
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
	Avoid         []string
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
)

type PathNode struct {
	Position BlockPos
	Move     MoveKind
	Cost     float64
}
type PathProgress struct {
	Index, Total int
	Position     Vec3
	Replanned    bool
	Target       BlockPos
	Move         MoveKind
}
type NavigationResult struct {
	Path    []PathNode
	Replans int
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
	jumpIndex    int
}
type Navigator struct {
	bot        *Bot
	mu         sync.Mutex
	active     *navRun
	sprinting  bool
	onProgress event[PathProgress]
}

func newNavigator(b *Bot) *Navigator {
	n := &Navigator{bot: b}
	b.World.OnBlockChange(func(BlockChange) {
		n.mu.Lock()
		if n.active != nil {
			n.active.dirty = true
		}
		n.mu.Unlock()
	})
	return n
}
func (n *Navigator) OnProgress(fn func(PathProgress)) func() { return n.onProgress.subscribe(fn) }
func (n *Navigator) Stop()                                   { n.finish(context.Canceled) }
func (n *Navigator) Path(start BlockPos, goal Goal, opt NavigationOptions) ([]PathNode, error) {
	return n.findPath(start, goal, defaultsNav(opt))
}
func (n *Navigator) Navigate(ctx context.Context, goal Goal, opt NavigationOptions) (NavigationResult, error) {
	lease, err := n.bot.actions.acquire(ctx, controlMovement|controlView, priorityAutomation)
	if err != nil {
		return NavigationResult{}, err
	}
	defer lease.Release()
	ctx = lease.Context(ctx)
	opt = defaultsNav(opt)
	start := n.bot.Self.State().Position.Block()
	path, err := n.findPath(start, goal, opt)
	if err != nil {
		return NavigationResult{}, err
	}
	run := &navRun{ctx: ctx, goal: goal, options: opt, path: path, index: 1, done: make(chan error, 1), result: NavigationResult{Path: path}, lastDistance: math.MaxFloat64, jumpIndex: -1}
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
	return o
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
		path, err := n.findPath(at, r.goal, r.options)
		if err != nil {
			n.active = nil
			n.mu.Unlock()
			r.done <- err
			return
		}
		r.path = path
		r.result.Path = path
		r.index = 1
		r.result.Replans++
		r.dirty = false
	}
	if r.index >= len(r.path) {
		n.mu.Unlock()
		return
	}
	target := r.path[r.index]
	if target.Move == MoveBreak && !r.digging {
		r.digging = true
		idx := r.index
		n.mu.Unlock()
		go func() {
			_, err := n.bot.Miner.Dig(r.ctx, target.Position, DigOptions{})
			n.mu.Lock()
			defer n.mu.Unlock()
			if n.active != r {
				return
			}
			r.digging = false
			if err != nil {
				n.active = nil
				r.done <- err
				return
			}
			if r.index == idx {
				r.dirty = true
			}
		}()
		n.move(physicsInput{}, state)
		return
	}
	if target.Move == MoveDoor && r.actionIndex != r.index {
		r.actionIndex = r.index
		seq := n.bot.Miner.sequence.Add(1)
		_ = n.bot.send(context.Background(), &packets.C2SUseItemOn{Hand: 0, Location: ns.NewPosition(target.Position.X, target.Position.Y, target.Position.Z), Face: 1, CursorPositionX: .5, CursorPositionY: .5, CursorPositionZ: .5, Sequence: ns.VarInt(seq)})
	}
	dx := float64(target.Position.X) + .5 - state.Position.X
	dz := float64(target.Position.Z) + .5 - state.Position.Z
	dist := math.Hypot(dx, dz)
	if dist < .18 && math.Abs(state.Position.Y-float64(target.Position.Y)) < .55 {
		r.index++
		n.mu.Unlock()
		progress := PathProgress{Index: r.index, Total: len(r.path), Position: state.Position}
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
	input := physicsInput{X: dx, Z: dz}
	input.Sprint = r.options.Sprint
	if target.Move == MoveJump && target.Position.Y > at.Y {
		input.Jump = true
		// Clear the obstacle vertically before translating across its upper
		// edge. Sending a simultaneous edge-touching rise is repeatedly
		// corrected by Paper and leaves the player frozen at jump-frame one.
		if state.Position.Y < float64(target.Position.Y)-0.05 {
			input.X, input.Z = 0, 0
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

func (n *Navigator) move(input physicsInput, state SelfState) {
	result := n.physicsStep(state, input)
	yaw := state.Rotation.Yaw
	if input.X != 0 || input.Z != 0 {
		yaw = float32(math.Atan2(-input.X, input.Z) * 180 / math.Pi)
	}
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
	if err := n.bot.send(context.Background(), &packets.C2SMovePlayerPosRot{X: ns.Float64(result.Position.X), FeetY: ns.Float64(result.Position.Y), Z: ns.Float64(result.Position.Z), Yaw: ns.Float32(yaw), Pitch: ns.Float32(state.Rotation.Pitch), Flags: flags}); err != nil {
		return
	}
	n.bot.Self.update(func(s *SelfState) {
		s.Position, s.Velocity, s.OnGround = result.Position, result.Velocity, result.OnGround
		s.Rotation.Yaw = yaw
	})
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
	open := &pathHeap{}
	heap.Init(open)
	root := &pathEntry{pos: start, f: goal.Estimate(start)}
	heap.Push(open, root)
	best := map[BlockPos]float64{start: 0}
	explored := 0
	for open.Len() > 0 {
		cur := heap.Pop(open).(*pathEntry)
		if goal.Reached(cur.pos) {
			return buildPath(cur), nil
		}
		explored++
		if explored > opt.MaxNodes {
			return nil, fmt.Errorf("%w: node limit %d", ErrUnreachable, opt.MaxNodes)
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
	return nil, ErrUnreachable
}
func buildPath(end *pathEntry) []PathNode {
	var p []PathNode
	for x := end; x != nil; x = x.parent {
		p = append(p, PathNode{x.pos, x.move, x.g})
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
		if move, cost, ok := n.passable(q, opt); ok {
			out = append(out, PathNode{q, move, cost * multiplier})
			continue
		}
		q.Y++
		if _, cost, ok := n.passable(q, opt); ok && n.clear(BlockPos{p.X, p.Y + 2, p.Z}) {
			out = append(out, PathNode{q, MoveJump, (cost + 1.5) * multiplier})
			continue
		}
		for drop := 1; drop <= opt.MaxDrop; drop++ {
			q.Y = p.Y - drop
			if _, cost, ok := n.passable(q, opt); ok {
				out = append(out, PathNode{q, MoveDrop, (cost + float64(drop)*.4) * multiplier})
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
					out = append(out, PathNode{landing, MoveParkour, cost * float64(gap+1) * .9})
					break
				}
			}
		}
	}
	if n.climbable(p) {
		for _, dy := range []int{-1, 1} {
			q := BlockPos{p.X, p.Y + dy, p.Z}
			if n.clear(q) {
				out = append(out, PathNode{q, MoveClimb, 1.5})
			}
		}
	}
	if b, ok := n.bot.World.Block(p); ok && isFluid(b.Name) {
		for _, dy := range []int{-1, 1} {
			q := BlockPos{p.X, p.Y + dy, p.Z}
			if move, cost, ok := n.passable(q, opt); ok && move == MoveSwim {
				out = append(out, PathNode{q, MoveSwim, cost + .5})
			}
		}
	}
	return out
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
		if opt.AllowBreaking && feet.Hardness >= 0 {
			return MoveBreak, 2 + float64(feet.Hardness)*2, true
		}
		return 0, 0, false
	}
	if !n.blockClear(head) && !isFluid(head.Name) {
		return 0, 0, false
	}
	below, ok := n.bot.World.Block(BlockPos{p.X, p.Y - 1, p.Z})
	if !ok {
		return 0, 0, false
	}
	if !n.blockSolid(below) && !fluid && !n.climbable(p) {
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
