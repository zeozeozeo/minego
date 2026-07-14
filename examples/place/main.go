package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/zeozeozeo/minego"
)

func main() {
	address := flag.String("address", "127.0.0.1:25565", "Minecraft server address")
	username := flag.String("username", "MineGo", "offline-mode username")
	item := flag.String("item", "cobblestone", "hotbar block item")
	x, y, z := flag.Int("x", 0, "target X"), flag.Int("y", 64, "target Y"), flag.Int("z", 0, "target Z")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	bot, err := minego.New(minego.Config{Address: *address, Auth: minego.Offline(*username)})
	if err != nil {
		log.Fatal(err)
	}
	defer bot.Close()
	if err = bot.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	if err = bot.WaitReady(ctx); err != nil {
		log.Fatal(err)
	}
	result, err := bot.Builder.Place(ctx, minego.BlockPos{X: *x, Y: *y, Z: *z}, minego.PlaceOptions{Item: *item})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("placed %s at %+v using face %d of %+v\n", result.Block.Name, result.Position, result.Face, result.Support)
}
