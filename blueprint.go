package minego

import (
	"context"
	"fmt"
	"sort"
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
type FindSiteOptions struct{ Width, Depth, Height, Radius int }
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
	for r := 0; r <= opt.Radius; r++ {
		for dx := -r; dx <= r; dx++ {
			for dz := -r; dz <= r; dz++ {
				if abs(dx) != r && abs(dz) != r {
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
	}
	return BlockPos{}, ErrNoBuildSite
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
	sort.SliceStable(blocks, func(i, j int) bool {
		if blocks[i].Offset.Y != blocks[j].Offset.Y {
			return blocks[i].Offset.Y < blocks[j].Offset.Y
		}
		return blocks[i].Offset.X+blocks[i].Offset.Z < blocks[j].Offset.X+blocks[j].Offset.Z
	})
	for _, entry := range blocks {
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
		if err := b.ensureHotbar(ctx, item); err != nil {
			return result, err
		}
		if _, err := b.bot.Navigator.Navigate(ctx, GoalNear{Position: pos, Radius: 3.5}, opt.Navigation); err != nil {
			return result, err
		}
		if _, err := b.Place(ctx, pos, PlaceOptions{Item: item}); err != nil {
			return result, err
		}
		result.Placed++
		b.onProgress.emit(BuildProgress{Completed: result.Placed + result.Skipped, Total: len(blocks), Position: pos, Item: item})
	}
	return result, nil
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
