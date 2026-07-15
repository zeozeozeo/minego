package minego

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type BlueprintBlock struct {
	Offset BlockPos
	Item   string
}
type Blueprint struct {
	Name   string
	Blocks []BlueprintBlock
}
type FindSiteOptions struct {
	Width, Depth, Height, Radius int
	// AllowClearing permits a site whose floor is solid but whose build volume
	// contains diggable terrain. Call ClearSite before building at such a site.
	AllowClearing bool
}
type BuildOptions struct {
	Navigation    NavigationOptions
	AllowClearing bool
	Timeout       time.Duration
}
type BuildProgress struct {
	Completed, Total int
	Position         BlockPos
	Item             string
}
type BuildResult struct {
	Origin                   BlockPos
	Placed, Skipped, Cleared int
}

func (b *Builder) OnProgress(fn func(BuildProgress)) func() { return b.onProgress.subscribe(fn) }
func (b *Builder) Materials(blueprint Blueprint) map[string]int {
	out := make(map[string]int)
	for _, block := range blueprint.Blocks {
		out[identifier(block.Item)]++
	}
	return out
}

func (b *Builder) FindSite(opt FindSiteOptions) (BlockPos, error) {
	if opt.Width <= 0 {
		opt.Width = 7
	}
	if opt.Depth <= 0 {
		opt.Depth = 7
	}
	if opt.Height <= 0 {
		opt.Height = 5
	}
	if opt.Radius <= 0 {
		opt.Radius = 64
	}
	start := b.bot.Self.State().Position.Block()
	bestCost := math.MaxInt
	var best BlockPos
	found := false
	for r := 0; r <= opt.Radius; r++ {
		for dx := -r; dx <= r; dx++ {
			for dz := -r; dz <= r; dz++ {
				if abs(dx) != r && abs(dz) != r {
					continue
				}
				if opt.AllowClearing {
					origin, ok := b.surfaceSiteOrigin(start, start.X+dx, start.Z+dz, opt)
					if !ok {
						continue
					}
					if b.siteClear(origin, opt) {
						return origin, nil
					}
					if cost, ok := b.siteClearCost(origin, opt); ok && cost < bestCost {
						bestCost, best, found = cost, origin, true
					}
					continue
				}
				for y := start.Y + 8; y >= start.Y-8; y-- {
					origin := BlockPos{start.X + dx, y, start.Z + dz}
					if b.siteClear(origin, opt) {
						return origin, nil
					}
				}
			}
		}
		// Prefer a modest amount of nearby leveling to walking dozens of blocks
		// in search of a mathematically perfect rectangle.
		if found && (bestCost <= opt.Width || r >= 8) {
			return best, nil
		}
	}
	if found {
		return best, nil
	}
	return BlockPos{}, ErrNoBuildSite
}

func (b *Builder) surfaceSiteOrigin(start BlockPos, x0, z0 int, opt FindSiteOptions) (BlockPos, bool) {
	lowest := math.MaxInt
	for x := 0; x < opt.Width; x++ {
		for z := 0; z < opt.Depth; z++ {
			ground := math.MinInt
			for y := start.Y + 8; y >= start.Y-12; y-- {
				block, ok := b.bot.World.Block(BlockPos{x0 + x, y, z0 + z})
				if !ok {
					return BlockPos{}, false
				}
				if len(block.Collision) > 0 && !strings.HasSuffix(block.Name, "_leaves") && !strings.HasSuffix(block.Name, "_log") && !strings.HasSuffix(block.Name, "_wood") {
					ground = y
					break
				}
			}
			if ground == math.MinInt {
				return BlockPos{}, false
			}
			lowest = min(lowest, ground)
		}
	}
	return BlockPos{x0, lowest + 1, z0}, true
}

func (b *Builder) siteClearCost(origin BlockPos, opt FindSiteOptions) (int, bool) {
	cost := 0
	for x := 0; x < opt.Width; x++ {
		for z := 0; z < opt.Depth; z++ {
			below, ok := b.bot.World.Block(BlockPos{origin.X + x, origin.Y - 1, origin.Z + z})
			if !ok || len(below.Collision) == 0 {
				return 0, false
			}
			for y := 0; y < opt.Height; y++ {
				block, ok := b.bot.World.Block(BlockPos{origin.X + x, origin.Y + y, origin.Z + z})
				if !ok || block.Hardness < 0 || isFluid(block.Name) {
					return 0, false
				}
				if !replaceableBlock(block.Name) {
					cost++
				}
			}
		}
	}
	return cost, true
}

// ClearSite levels the build volume while retaining the solid floor beneath
// origin. Blocks are removed top-down so every subsequent target is exposed.
func (b *Builder) ClearSite(ctx context.Context, origin BlockPos, opt FindSiteOptions, nav NavigationOptions) (int, error) {
	var targets []BlockPos
	for y := opt.Height - 1; y >= 0; y-- {
		for x := 0; x < opt.Width; x++ {
			for z := 0; z < opt.Depth; z++ {
				pos := BlockPos{origin.X + x, origin.Y + y, origin.Z + z}
				block, ok := b.bot.World.Block(pos)
				if ok && !replaceableBlock(block.Name) {
					targets = append(targets, pos)
				}
			}
		}
	}
	cleared := 0
	for len(targets) > 0 {
		self := b.bot.Self.State().Position
		sort.SliceStable(targets, func(i, j int) bool {
			if targets[i].Y != targets[j].Y {
				return targets[i].Y > targets[j].Y
			}
			a := Vec3{float64(targets[i].X) + .5, float64(targets[i].Y) + .5, float64(targets[i].Z) + .5}
			c := Vec3{float64(targets[j].X) + .5, float64(targets[j].Y) + .5, float64(targets[j].Z) + .5}
			return self.Distance(a) < self.Distance(c)
		})
		pos := targets[0]
		targets = targets[1:]
		block, ok := b.bot.World.Block(pos)
		if !ok || replaceableBlock(block.Name) {
			continue
		}
		if _, err := b.bot.Navigator.Navigate(ctx, digGoal{target: pos}, nav); err != nil {
			return cleared, err
		}
		clearNav := nav
		clearNav.BreakFilter = nil
		if err := b.bot.Miner.clearLineOfSight(ctx, pos, clearNav); err != nil {
			return cleared, err
		}
		if _, err := b.bot.Miner.Dig(ctx, pos, DigOptions{}); err != nil {
			return cleared, err
		}
		cleared++
	}
	return cleared, nil
}
func (b *Builder) siteClear(origin BlockPos, opt FindSiteOptions) bool {
	for x := 0; x < opt.Width; x++ {
		for z := 0; z < opt.Depth; z++ {
			below, ok := b.bot.World.Block(BlockPos{origin.X + x, origin.Y - 1, origin.Z + z})
			if !ok || len(below.Collision) == 0 {
				return false
			}
			for y := 0; y < opt.Height; y++ {
				block, ok := b.bot.World.Block(BlockPos{origin.X + x, origin.Y + y, origin.Z + z})
				if !ok || !replaceableBlock(block.Name) {
					return false
				}
			}
		}
	}
	return true
}

func (b *Builder) Build(ctx context.Context, origin BlockPos, blueprint Blueprint, opt BuildOptions) (BuildResult, error) {
	result := BuildResult{Origin: origin}
	if opt.Timeout <= 0 {
		opt.Timeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, opt.Timeout)
	defer cancel()
	blocks := append([]BlueprintBlock(nil), blueprint.Blocks...)
	total := len(blocks)
	sort.SliceStable(blocks, func(i, j int) bool {
		if blocks[i].Offset.Y != blocks[j].Offset.Y {
			return blocks[i].Offset.Y < blocks[j].Offset.Y
		}
		return blocks[i].Offset.X+blocks[i].Offset.Z < blocks[j].Offset.X+blocks[j].Offset.Z
	})
	for len(blocks) > 0 {
		entryIndex := b.nextBuildBlock(origin, blocks, opt.AllowClearing)
		if entryIndex < 0 {
			return result, fmt.Errorf("%w: no remaining block has a usable support face", ErrBuildObstructed)
		}
		entry := blocks[entryIndex]
		blocks = append(blocks[:entryIndex], blocks[entryIndex+1:]...)
		pos := BlockPos{origin.X + entry.Offset.X, origin.Y + entry.Offset.Y, origin.Z + entry.Offset.Z}
		item := identifier(entry.Item)
		existing, ok := b.bot.World.Block(pos)
		if !ok {
			return result, fmt.Errorf("minego: blueprint position %v is not loaded", pos)
		}
		if existing.Name == item {
			result.Skipped++
			continue
		}
		if !replaceableBlock(existing.Name) {
			if !opt.AllowClearing {
				return result, fmt.Errorf("%w: %s at %v", ErrBuildObstructed, existing.Name, pos)
			}
			if _, err := b.bot.Navigator.Navigate(ctx, GoalNear{Position: pos, Radius: 3.5}, opt.Navigation); err != nil {
				return result, err
			}
			if _, err := b.bot.Miner.Dig(ctx, pos, DigOptions{}); err != nil {
				return result, err
			}
			result.Cleared++
		}
		if _, err := b.bot.Navigator.Navigate(ctx, placementGoal{builder: b, target: pos, reach: 4.5}, opt.Navigation); err != nil {
			return result, err
		}
		// Navigation may bridge or pillar with a temporary block and can replace
		// the selected hotbar slot. Stage the requested item only after movement,
		// immediately before Place consumes it.
		if err := b.ensureHotbar(ctx, item); err != nil {
			return result, err
		}
		if _, err := b.Place(ctx, pos, PlaceOptions{Item: item}); err != nil {
			return result, err
		}
		result.Placed++
		b.onProgress.emit(BuildProgress{Completed: result.Placed + result.Skipped, Total: total, Position: pos, Item: item})
	}
	return result, nil
}

type placementGoal struct {
	builder *Builder
	target  BlockPos
	reach   float64
}

func (g placementGoal) Reached(position BlockPos) bool {
	eye := Vec3{float64(position.X) + .5, float64(position.Y) + 1.62, float64(position.Z) + .5}
	_, _, _, ok := g.builder.supportFaceFrom(g.target, eye, g.reach)
	return ok
}

func (g placementGoal) Estimate(position BlockPos) float64 {
	return math.Max(0, distance(position, g.target)-3)
}

func (b *Builder) nextBuildBlock(origin BlockPos, blocks []BlueprintBlock, allowClearing bool) int {
	self := b.bot.Self.State().Position
	bestScore := math.MaxFloat64
	best := -1
	for i, entry := range blocks {
		pos := BlockPos{origin.X + entry.Offset.X, origin.Y + entry.Offset.Y, origin.Z + entry.Offset.Z}
		existing, ok := b.bot.World.Block(pos)
		if !ok {
			continue
		}
		if existing.Name != identifier(entry.Item) && !replaceableBlock(existing.Name) && !allowClearing {
			continue
		}
		if existing.Name != identifier(entry.Item) {
			if _, _, _, supported := b.supportFace(pos); !supported && replaceableBlock(existing.Name) {
				continue
			}
		}
		center := Vec3{float64(pos.X) + .5, float64(pos.Y) + .5, float64(pos.Z) + .5}
		score := self.Distance(center)
		eye := self
		eye.Y += 1.62
		if _, _, _, reachable := b.supportFaceFrom(pos, eye, 4.5); reachable {
			score -= 1000
		}
		if score < bestScore {
			bestScore, best = score, i
		}
	}
	return best
}
func (b *Builder) ensureHotbar(ctx context.Context, item string) error {
	if _, _, err := b.placementItem(item); err == nil {
		return nil
	}
	slots := b.bot.Inventory.Slots()
	source := findItem(slots, item)
	if source < 0 {
		return fmt.Errorf("%w: %s", ErrNoPlacementItem, item)
	}
	dest := -1
	for hot := 0; hot < 9; hot++ {
		idx := 36 + hot
		if idx >= len(slots) {
			idx = hot
		}
		if slots[idx].Count <= 0 {
			dest = idx
			break
		}
	}
	if dest < 0 {
		dest = 36 + b.bot.Inventory.Selected()
		if dest >= len(slots) {
			dest = b.bot.Inventory.Selected()
		}
	}
	w := b.bot.Crafter.window(0)
	for _, slot := range []int{source, dest, source} {
		if err := b.bot.Crafter.click(ctx, w, slot, 0, 0); err != nil {
			return err
		}
		w = b.bot.Crafter.window(0)
	}
	return nil
}
