package minego

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestOfflineServer26_2 is opt-in because it requires an official 26.2 server.
// It verifies login and authoritative spawn synchronization without credentials.
func TestOfflineServer26_2(t *testing.T) {
	address := os.Getenv("MINEGO_TEST_SERVER_26_2")
	if address == "" {
		t.Skip("set MINEGO_TEST_SERVER_26_2 to run the dedicated-server test")
	}
	bot, err := New(Config{Address: address, Auth: Offline("MineGoTest")})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = bot.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer bot.Close()
	if err = bot.WaitReady(ctx); err != nil {
		t.Fatal(err)
	}
	if !bot.World.IsLoaded(bot.Self.State().Position.Block()) {
		deadline := time.NewTimer(10 * time.Second)
		defer deadline.Stop()
		for !bot.World.IsLoaded(bot.Self.State().Position.Block()) {
			select {
			case <-deadline.C:
				t.Fatal("spawn chunk did not synchronize")
			case <-time.After(50 * time.Millisecond):
			}
		}
	}
}
