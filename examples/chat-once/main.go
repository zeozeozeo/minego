package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/zeozeozeo/minego"
)

func main() {
	address := flag.String("address", "127.0.0.1:25565", "Minecraft server address")
	username := flag.String("username", "MineGo", "offline-mode username")
	message := flag.String("message", "Hello from MineGo!", "chat message")
	trace := flag.Bool("trace", false, "print raw packet metadata")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bot, err := minego.New(minego.Config{
		Address: *address,
		Auth:    minego.Offline(*username),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *trace {
		bot.OnPacket(func(packet minego.RawPacket) {
			fmt.Printf("packet state=%d id=0x%02x bytes=%d\n", packet.State, packet.ID, len(packet.Data))
		})
	}
	bot.OnDisconnect(func(event minego.DisconnectEvent) {
		fmt.Printf("disconnect: %s (%v)\n", event.Reason, event.Err)
	})
	if err := bot.Connect(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer bot.Close()

	if err := bot.WaitReady(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("joined %s as %s at %+v\n", *address, *username, bot.Self.State().Position)

	if err := bot.Chat.Send(ctx, *message); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("sent chat message: %q\n", *message)

	// Give the connection loop enough time to flush the message before closing.
	time.Sleep(500 * time.Millisecond)
}
