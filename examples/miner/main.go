package main

import (
	"context"
	"github.com/zeozeozeo/minego"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	bot, err := minego.New(minego.Config{Address: os.Getenv("MINECRAFT_ADDRESS"), Auth: minego.Offline("MineGo")})
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
	result, err := bot.Miner.Mine(ctx, minego.Tags("minecraft:diamond_ores"), 1, minego.MineOptions{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("mined %d block(s)", result.Completed)
}
