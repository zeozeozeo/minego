package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/zeozeozeo/minego"
)

type controller struct {
	bot  *minego.Bot
	root context.Context

	mu      sync.Mutex
	cancel  context.CancelFunc
	request uint64
}

func (c *controller) requestBlock(name string) {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "_")
	if name == "" || strings.ContainsAny(name, "\t\r\n/") {
		return
	}
	if !strings.Contains(name, ":") {
		name = "minecraft:" + name
	}
	if !c.bot.HasBlock(name) {
		return // Ordinary conversation is not interpreted as a mining request.
	}

	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
		c.bot.Navigator.Stop()
	}
	ctx, cancel := context.WithCancel(c.root)
	c.cancel = cancel
	c.request++
	id := c.request
	c.mu.Unlock()

	log.Printf("request %d: mining one %s", id, name)
	go c.mine(ctx, id, name)
}

func (c *controller) mine(ctx context.Context, id uint64, name string) {
	result, err := c.bot.Miner.Mine(ctx, minego.Blocks(name), 1, minego.MineOptions{
		Navigation: minego.NavigationOptions{Sprint: true, MaxDrop: 3, AllowParkour: true, MaxParkourGap: 2},
	})

	c.mu.Lock()
	current := id == c.request
	c.mu.Unlock()
	if !current {
		log.Printf("request %d: superseded", id)
		return
	}
	if err != nil {
		if ctx.Err() != nil {
			log.Printf("request %d: cancelled", id)
		} else {
			log.Printf("request %d: failed: %v", id, err)
		}
		return
	}
	log.Printf("request %d: mined %s at %+v", id, name, result.Mined[0])
}

func main() {
	address := flag.String("address", "127.0.0.1:25565", "Minecraft server address")
	username := flag.String("username", "MineGo", "offline-mode username")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bot, err := minego.New(minego.Config{Address: *address, Auth: minego.Offline(*username)})
	if err != nil {
		log.Fatal(err)
	}
	defer bot.Close()

	controller := &controller{bot: bot, root: ctx}
	bot.Chat.OnMessage(func(message minego.ChatMessage) {
		if message.Kind != minego.ChatPlayer {
			return
		}
		log.Printf("chat from %s: %q", message.Sender, message.Text)
		controller.requestBlock(message.Text)
	})
	bot.OnDisconnect(func(event minego.DisconnectEvent) {
		log.Printf("disconnected: %s: %v", event.Reason, event.Err)
	})
	bot.Navigator.OnProgress(func(progress minego.PathProgress) {
		log.Printf("path %d/%d position=(%.3f, %.3f, %.3f) target=%+v move=%d replan=%t", progress.Index, progress.Total, progress.Position.X, progress.Position.Y, progress.Position.Z, progress.Target, progress.Move, progress.Replanned)
	})
	bot.Miner.OnProgress(func(progress minego.MiningProgress) {
		log.Printf("mining %s at %+v (%d/%d)", progress.Kind, progress.Position, progress.Completed, progress.Requested)
	})

	if err := bot.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	if err := bot.WaitReady(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("chat miner joined %s as %s using %s; send a block name such as 'sand' in chat\n", *address, *username, bot.Version().Name)

	select {
	case <-ctx.Done():
	case <-bot.Done():
	}
}
