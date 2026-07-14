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
	_, err = bot.Navigator.Navigate(ctx, minego.GoalNear{Position: minego.BlockPos{X: 20, Y: 64, Z: 20}, Radius: 1}, minego.NavigationOptions{Sprint: true})
	if err != nil {
		log.Fatal(err)
	}
}
