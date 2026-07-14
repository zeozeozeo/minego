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
	bot.Chat.OnMessage(func(m minego.ChatMessage) { log.Printf("%s: %s", m.Sender, m.Text) })
	if err = bot.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer bot.Close()
	if err = bot.WaitReady(ctx); err != nil {
		log.Fatal(err)
	}
	<-bot.Done()
}
