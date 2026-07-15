package minego

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

type RecipeIngredient struct{ Alternatives []string }
type Recipe struct {
	ID            string
	Output        ItemStack
	Width, Height int
	Ingredients   []RecipeIngredient
}
type CraftOptions struct {
	Table     *BlockPos
	Recursive bool
	Timeout   time.Duration
}
type CraftResult struct {
	Item                string
	Requested, Produced int
	Recipe              Recipe
	CraftedDependencies []string
}

type Crafter struct {
	bot     *Bot
	mu      sync.RWMutex
	recipes []Recipe
}

func newCrafter(bot *Bot) *Crafter { return &Crafter{bot: bot, recipes: generatedCraftingRecipes()} }
func (c *Crafter) RecipesFor(output string) []Recipe {
	output = identifier(output)
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []Recipe
	for _, r := range c.recipes {
		if r.Output.Name == output {
			out = append(out, cloneRecipe(r))
		}
	}
	return out
}

// RegisterRecipe makes server/datapack recipes available to the resolver.
func (c *Crafter) RegisterRecipe(recipe Recipe) {
	recipe.Output.Name = identifier(recipe.Output.Name)
	c.mu.Lock()
	c.recipes = append(c.recipes, cloneRecipe(recipe))
	c.mu.Unlock()
}

func (c *Crafter) Craft(ctx context.Context, output string, count int, opt CraftOptions) (CraftResult, error) {
	result := CraftResult{Item: identifier(output), Requested: count}
	if count <= 0 {
		return result, nil
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 30 * time.Second
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, opt.Timeout)
	defer cancel()
	lease, err := c.bot.actions.acquire(ctx, controlMovement|controlView|controlHands|controlInventory|controlWindows, priorityExplicit)
	if err != nil {
		return result, err
	}
	defer lease.Release()
	ctx = lease.Context(ctx)
	recipes := c.RecipesFor(result.Item)
	if len(recipes) == 0 {
		return result, fmt.Errorf("%w: %s", ErrNoRecipe, result.Item)
	}
	var recipe Recipe
	found := false
	bestScore := -1
	available := inventoryCounts(c.bot.Inventory.Slots())
	for _, candidate := range recipes {
		if opt.Table != nil || (candidate.Width <= 2 && candidate.Height <= 2) {
			score := 0
			for _, ing := range candidate.Ingredients {
				for _, name := range ing.Alternatives {
					if available[identifier(name)] > 0 {
						score++
						break
					}
				}
			}
			if score > bestScore {
				recipe = candidate
				found = true
				bestScore = score
			}
		}
	}
	if !found {
		return result, ErrCraftingTableRequired
	}
	result.Recipe = recipe
	batches := (count + int(recipe.Output.Count) - 1) / int(recipe.Output.Count)
	if opt.Recursive {
		if err := c.validateAcyclic(result.Item, map[string]bool{}, map[string]bool{}); err != nil {
			return result, err
		}
		deps, err := c.craftDependencies(ctx, recipe, batches, opt, map[string]bool{result.Item: true})
		if err != nil {
			return result, err
		}
		result.CraftedDependencies = deps
	}
	windowID := int32(0)
	if opt.Table != nil {
		id, err := c.openTable(ctx, *opt.Table)
		if err != nil {
			return result, err
		}
		windowID = id
		defer c.closeWindow(context.Background(), windowID)
	}
	for batch := 0; batch < batches; batch++ {
		if err := c.fillAndTake(ctx, windowID, recipe); err != nil {
			return result, err
		}
		result.Produced += int(recipe.Output.Count)
	}
	return result, nil
}

func cloneRecipe(r Recipe) Recipe {
	r.Output = cloneStack(r.Output)
	r.Ingredients = append([]RecipeIngredient(nil), r.Ingredients...)
	for i := range r.Ingredients {
		r.Ingredients[i].Alternatives = append([]string(nil), r.Ingredients[i].Alternatives...)
	}
	return r
}
func (c *Crafter) validateAcyclic(item string, visiting, done map[string]bool) error {
	item = identifier(item)
	if done[item] {
		return nil
	}
	if visiting[item] {
		return fmt.Errorf("%w: recipe cycle at %s", ErrMissingIngredients, item)
	}
	visiting[item] = true
	for _, recipe := range c.RecipesFor(item) {
		for _, ing := range recipe.Ingredients {
			for _, alt := range ing.Alternatives {
				if len(c.RecipesFor(alt)) > 0 {
					if err := c.validateAcyclic(alt, visiting, done); err != nil {
						return err
					}
				}
			}
		}
	}
	delete(visiting, item)
	done[item] = true
	return nil
}

func (c *Crafter) craftDependencies(ctx context.Context, r Recipe, batches int, opt CraftOptions, visiting map[string]bool) ([]string, error) {
	need := make(map[string]int)
	available := inventoryCounts(c.bot.Inventory.Slots())
	for _, ing := range r.Ingredients {
		choice := ""
		for _, a := range ing.Alternatives {
			a = identifier(a)
			if available[a] > 0 {
				choice = a
				break
			}
		}
		if choice == "" && len(ing.Alternatives) > 0 {
			choice = identifier(ing.Alternatives[0])
		}
		need[choice] += batches
	}
	var made []string
	for item, n := range need {
		if available[item] >= n {
			continue
		}
		missing := n - available[item]
		if visiting[item] {
			return made, fmt.Errorf("%w: cycle at %s", ErrMissingIngredients, item)
		}
		visiting[item] = true
		sub := opt
		sub.Recursive = true
		res, err := c.Craft(ctx, item, missing, sub)
		delete(visiting, item)
		if err != nil {
			return made, fmt.Errorf("%w: need %d %s", ErrMissingIngredients, missing, item)
		}
		made = append(made, res.Item)
		made = append(made, res.CraftedDependencies...)
	}
	return made, nil
}

func (c *Crafter) openTable(ctx context.Context, pos BlockPos) (int32, error) {
	if _, err := c.bot.Navigator.Navigate(ctx, GoalAdjacent(pos), NavigationOptions{}); err != nil {
		return 0, err
	}
	seq := c.bot.Miner.sequence.Add(1)
	if err := c.bot.send(ctx, &packets.C2SUseItemOn{Hand: 0, Location: ns.NewPosition(pos.X, pos.Y, pos.Z), Face: 1, CursorPositionX: .5, CursorPositionY: .5, CursorPositionZ: .5, Sequence: ns.VarInt(seq)}); err != nil {
		return 0, err
	}
	return c.waitWindow(ctx, func(w WindowSnapshot) bool { return w.ID != 0 && w.Type == 12 })
}
func (c *Crafter) waitWindow(ctx context.Context, predicate func(WindowSnapshot) bool) (int32, error) {
	if w := c.bot.Inventory.Window(); predicate(w) {
		return w.ID, nil
	}
	ch := make(chan WindowSnapshot, 1)
	unsub := c.bot.Inventory.OnWindowChange(func(e WindowChange) {
		if predicate(e.Window) {
			select {
			case ch <- e.Window:
			default:
			}
		}
	})
	defer unsub()
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case w := <-ch:
		return w.ID, nil
	}
}
func (c *Crafter) closeWindow(ctx context.Context, id int32) {
	_ = c.bot.send(ctx, &packets.C2SContainerClose{WindowId: ns.VarInt(id)})
}

func (c *Crafter) fillAndTake(ctx context.Context, windowID int32, r Recipe) error {
	gridWidth := 2
	gridSlots := 4
	inventoryStart := 9
	if windowID != 0 {
		gridWidth = 3
		gridSlots = 9
		inventoryStart = 10
	}
	for cell, ing := range r.Ingredients {
		if len(ing.Alternatives) == 0 {
			continue
		}
		row, col := cell/r.Width, cell%r.Width
		grid := 1 + row*gridWidth + col
		if grid > gridSlots {
			return ErrCraftingTableRequired
		}
		w := c.window(windowID)
		source := -1
		for slot := inventoryStart; slot < len(w.Slots); slot++ {
			for _, a := range ing.Alternatives {
				if w.Slots[slot].Name == identifier(a) && w.Slots[slot].Count > 0 {
					source = slot
					break
				}
			}
			if source >= 0 {
				break
			}
		}
		if source < 0 {
			return fmt.Errorf("%w: %v", ErrMissingIngredients, ing.Alternatives)
		}
		if err := c.click(ctx, w, source, 0, 0); err != nil {
			return err
		}
		w = c.window(windowID)
		if err := c.click(ctx, w, grid, 1, 0); err != nil {
			return err
		}
		w = c.window(windowID)
		if err := c.click(ctx, w, source, 0, 0); err != nil {
			return err
		}
	}
	deadline := time.NewTicker(25 * time.Millisecond)
	defer deadline.Stop()
	for {
		w := c.window(windowID)
		if len(w.Slots) > 0 && w.Slots[0].Count > 0 {
			return c.click(ctx, w, 0, 0, 1)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
		}
	}
}
func (c *Crafter) window(id int32) WindowSnapshot {
	w := c.bot.Inventory.Window()
	if id == 0 && len(w.Slots) == 0 {
		w = WindowSnapshot{ID: 0, Slots: c.bot.Inventory.Slots()}
	}
	return w
}
func (c *Crafter) click(ctx context.Context, w WindowSnapshot, slot int, button int8, mode int32) error {
	before := w.StateID
	packet := &packets.C2SContainerClick{WindowId: ns.VarInt(w.ID), StateId: ns.VarInt(w.StateID), Slot: ns.Int16(slot), Button: ns.Int8(button), Mode: ns.VarInt(mode), CarriedItem: w.Carried.hashed()}
	if err := c.bot.send(ctx, packet); err != nil {
		return err
	}
	_, err := c.waitWindow(ctx, func(next WindowSnapshot) bool { return next.ID == w.ID && next.StateID != before })
	return err
}
func identifier(s string) string {
	if s != "" && !strings.Contains(s, ":") {
		return "minecraft:" + s
	}
	return s
}
func inventoryCounts(slots []ItemStack) map[string]int {
	m := make(map[string]int)
	for _, s := range slots {
		m[s.Name] += int(s.Count)
	}
	return m
}
