package main

import (
	"context"
	"flag"
	"log"
	"strings"
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
	if err = bot.WaitReady(ctx); err != nil {
		log.Fatal(err)
	}
	first, err := bot.Miner.Mine(ctx, minego.Tags("minecraft:logs"), 1, minego.MineOptions{Navigation: survivalNavigation()})
	if err != nil || len(first.Blocks) == 0 {
		log.Fatalf("find wood: %v", err)
	}
	logItem := first.Blocks[0].Name
	wood := woodFamily(logItem)
	planks := "minecraft:" + wood + "_planks"
	door := "minecraft:" + wood + "_door"
	more, err := bot.Miner.Mine(ctx, minego.Blocks(logItem), 32, minego.MineOptions{Navigation: survivalNavigation()})
	if err != nil {
		log.Fatal(err)
	}
	pickup := first.Mined[0]
	for _, p := range more.Mined {
		if p.Y < pickup.Y {
			pickup = p
		}
	}
	_, _ = bot.Navigator.Navigate(ctx, minego.GoalNear{Position: pickup, Radius: 1.5}, survivalNavigation())
	if err = waitForItems(ctx, bot, logItem, 33); err != nil {
		log.Fatal(err)
	}
	if _, err = bot.Crafter.Craft(ctx, planks, 129, minego.CraftOptions{}); err != nil {
		log.Fatal(err)
	}
	if _, err = bot.Crafter.Craft(ctx, "minecraft:crafting_table", 1, minego.CraftOptions{}); err != nil {
		log.Fatal(err)
	}
	site, err := bot.Builder.FindSite(minego.FindSiteOptions{Width: 9, Depth: 7, Height: 5, Radius: 64})
	if err != nil {
		log.Fatal(err)
	}
	table := minego.BlockPos{X: site.X + 8, Y: site.Y, Z: site.Z}
	if _, err = bot.Builder.Build(ctx, table, minego.Blueprint{Name: "table", Blocks: []minego.BlueprintBlock{{Item: "minecraft:crafting_table"}}}, minego.BuildOptions{Navigation: survivalNavigation()}); err != nil {
		log.Fatal(err)
	}
	if _, err = bot.Crafter.Craft(ctx, door, 1, minego.CraftOptions{Table: &table}); err != nil {
		log.Fatal(err)
	}
	result, err := bot.Builder.Build(ctx, site, houseBlueprint(planks, door), minego.BuildOptions{Navigation: survivalNavigation()})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("built %s house at %+v (%d blocks)", wood, site, result.Placed)
}

func survivalNavigation() minego.NavigationOptions {
	return minego.NavigationOptions{Sprint: true, AllowBreaking: true, AllowPlacing: true, AcquireTemporary: true, TemporaryBlocks: []string{"dirt", "cobblestone"}}
}
func woodFamily(name string) string {
	s := strings.TrimPrefix(name, "minecraft:")
	s = strings.TrimPrefix(s, "stripped_")
	for _, suffix := range []string{"_log", "_wood", "_stem", "_hyphae"} {
		s = strings.TrimSuffix(s, suffix)
	}
	return s
}
func waitForItems(ctx context.Context, bot *minego.Bot, item string, count int) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		total := 0
		for _, stack := range bot.Inventory.Slots() {
			if stack.Name == item {
				total += int(stack.Count)
			}
		}
		if total >= count {
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
