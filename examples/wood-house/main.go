package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/zeozeozeo/minego"
)

func main() {
	address := flag.String("address", "127.0.0.1:25565", "Minecraft server address")
	username := flag.String("username", "MineGoHouse", "offline-mode username")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	bot, err := minego.New(minego.Config{Address: *address, Auth: minego.Offline(*username)})
	if err != nil {
		log.Fatal(err)
	}
	if err = bot.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer bot.Close()
	log.Printf("connected using Minecraft %s (protocol %d); waiting for initial world", bot.Version().Name, bot.Version().Protocol)
	if err = bot.WaitReady(ctx); err != nil {
		log.Fatal(err)
	}
	log.Printf("world ready at %+v", bot.Self.State().Position)
	bot.OnDisconnect(func(event minego.DisconnectEvent) {
		log.Printf("disconnected: %s: %v", event.Reason, event.Err)
	})
	announce := func(format string, args ...any) {
		message := fmt.Sprintf(format, args...)
		log.Print(message)
		if err := bot.Chat.Send(ctx, message); err != nil {
			log.Printf("send progress to chat: %v", err)
		}
	}
	started := time.Now()
	phaseStarted := started
	phase := func(name string) {
		now := time.Now()
		log.Printf("phase %s completed in %s (total %s)", name, now.Sub(phaseStarted).Round(time.Millisecond), now.Sub(started).Round(time.Millisecond))
		phaseStarted = now
	}
	var replans, clearedLeaves, routeMaterial atomic.Int64
	bot.Navigator.OnProgress(func(progress minego.PathProgress) {
		log.Printf("path %d/%d position=(%.1f, %.1f, %.1f) target=%+v move=%d replan=%t", progress.Index, progress.Total, progress.Position.X, progress.Position.Y, progress.Position.Z, progress.Target, progress.Move, progress.Replanned)
		if progress.Replanned {
			replans.Add(1)
		}
	})
	bot.Miner.OnProgress(func(progress minego.MiningProgress) {
		log.Printf("mining %s %s at %+v (%d/%d): %v", progress.Kind, progress.Name, progress.Position, progress.Completed, progress.Requested, progress.Err)
		if progress.Kind == "clearing" && strings.HasSuffix(progress.Name, "_leaves") {
			clearedLeaves.Add(1)
		}
		if progress.Kind == "target" && (progress.Name == "minecraft:dirt" || progress.Name == "minecraft:cobblestone") {
			routeMaterial.Add(1)
		}
		if progress.Kind == "target" && isLog(progress.Name) && (progress.Completed == 1 || progress.Completed%4 == 0 || progress.Completed == progress.Requested) {
			announce("Mined logs: %d/%d", progress.Completed, progress.Requested)
		}
	})
	bot.Builder.OnProgress(func(progress minego.BuildProgress) {
		if progress.Completed == 1 || progress.Completed%10 == 0 || progress.Completed == progress.Total {
			announce("Building house: %d/%d blocks", progress.Completed, progress.Total)
		}
	})
	announce("Looking for a tree")
	first, err := bot.Miner.Mine(ctx, minego.Tags("minecraft:logs"), 1, minego.MineOptions{Navigation: collectionNavigation()})
	if err != nil || len(first.Blocks) == 0 {
		log.Fatalf("find wood: %v", err)
	}
	logItem := first.Blocks[0].Name
	phase("tree search")
	wood := woodFamily(logItem)
	planks := "minecraft:" + wood + "_planks"
	door := "minecraft:" + wood + "_door"
	announce("Found %s; collecting enough logs for the house", logItem)
	for attempts := 0; itemCount(bot, logItem) < 33; attempts++ {
		if attempts >= 4 {
			log.Fatalf("could not collect enough %s: have %d/33", logItem, itemCount(bot, logItem))
		}
		missing := 33 - itemCount(bot, logItem)
		if _, err := bot.Miner.Mine(ctx, minego.Blocks(logItem), missing, minego.MineOptions{Navigation: collectionNavigation()}); err != nil {
			log.Fatal(err)
		}
		// Allow the final pickup packets to reach the inventory before deciding
		// whether replacement logs are needed for drops lost down terrain.
		_ = waitForItems(ctx, bot, logItem, 33, time.Second)
	}
	phase("log collection")
	announce("Crafting planks and a crafting table")
	if _, err = bot.Crafter.Craft(ctx, planks, 129, minego.CraftOptions{}); err != nil {
		log.Fatal(err)
	}
	if _, err = bot.Crafter.Craft(ctx, "minecraft:crafting_table", 1, minego.CraftOptions{}); err != nil {
		log.Fatal(err)
	}
	phase("initial crafting")
	siteOptions := minego.FindSiteOptions{Width: 9, Depth: 7, Height: 5, Radius: 64, AllowClearing: true}
	site, err := bot.Builder.FindSite(siteOptions)
	if err != nil {
		log.Fatal(err)
	}
	cleared, err := bot.Builder.ClearSite(ctx, site, siteOptions, constructionNavigation())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("prepared build site by clearing %d blocks", cleared)
	table := minego.BlockPos{X: site.X + 8, Y: site.Y, Z: site.Z}
	announce("Building at %d %d %d", site.X, site.Y, site.Z)
	if _, err = bot.Builder.Build(ctx, table, minego.Blueprint{Name: "table", Blocks: []minego.BlueprintBlock{{Item: "minecraft:crafting_table"}}}, minego.BuildOptions{Navigation: constructionNavigation()}); err != nil {
		log.Fatal(err)
	}
	if _, err = bot.Crafter.Craft(ctx, door, 1, minego.CraftOptions{Table: &table}); err != nil {
		log.Fatal(err)
	}
	phase("site and table")
	result, err := bot.Builder.Build(ctx, site, houseBlueprint(planks, door), minego.BuildOptions{Navigation: constructionNavigation()})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("built %s house at %+v (%d blocks)", wood, site, result.Placed)
	phase("house build")
	log.Printf("run metrics duration=%s replans=%d cleared_leaves=%d route_material=%d", time.Since(started).Round(time.Millisecond), replans.Load(), clearedLeaves.Load(), routeMaterial.Load())
	_ = bot.Chat.Send(ctx, fmt.Sprintf("Built the %s house at %d %d %d", wood, site.X, site.Y, site.Z))
}

func survivalNavigation(maxNodes int) minego.NavigationOptions {
	return minego.NavigationOptions{
		MaxNodes: maxNodes, Sprint: true, AllowBreaking: true, AllowPlacing: true, AcquireTemporary: true,
		TemporaryBlocks: []string{"dirt", "cobblestone"},
		BreakFilter:     func(block minego.Block) bool { return strings.HasSuffix(block.Name, "_leaves") },
	}
}
func collectionNavigation() minego.NavigationOptions   { return survivalNavigation(12000) }
func constructionNavigation() minego.NavigationOptions { return survivalNavigation(8000) }
func woodFamily(name string) string {
	s := strings.TrimPrefix(name, "minecraft:")
	s = strings.TrimPrefix(s, "stripped_")
	for _, suffix := range []string{"_log", "_wood", "_stem", "_hyphae"} {
		s = strings.TrimSuffix(s, suffix)
	}
	return s
}
func isLog(name string) bool {
	for _, suffix := range []string{"_log", "_wood", "_stem", "_hyphae"} {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}
func itemCount(bot *minego.Bot, item string) int {
	total := 0
	for _, stack := range bot.Inventory.Slots() {
		if stack.Name == item {
			total += int(stack.Count)
		}
	}
	return total
}
func waitForItems(ctx context.Context, bot *minego.Bot, item string, count int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if itemCount(bot, item) >= count {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
func houseBlueprint(planks, door string) minego.Blueprint {
	bp := minego.Blueprint{Name: "simple_wood_house"}
	for y := 0; y < 3; y++ {
		for x := 0; x < 7; x++ {
			for z := 0; z < 7; z++ {
				wall := x == 0 || x == 6 || z == 0 || z == 6
				if !wall {
					continue
				}
				if z == 0 && x == 3 && (y == 0 || y == 1) {
					continue
				}
				bp.Blocks = append(bp.Blocks, minego.BlueprintBlock{Offset: minego.BlockPos{X: x, Y: y, Z: z}, Item: planks})
			}
		}
	}
	for x := 0; x < 7; x++ {
		for z := 0; z < 7; z++ {
			bp.Blocks = append(bp.Blocks, minego.BlueprintBlock{Offset: minego.BlockPos{X: x, Y: 3, Z: z}, Item: planks})
		}
	}
	bp.Blocks = append(bp.Blocks, minego.BlueprintBlock{Offset: minego.BlockPos{X: 3, Y: 0, Z: 0}, Item: door})
	return bp
}
