package minego

import "testing"

func TestRespawnRequestedExactlyOncePerDeath(t *testing.T) {
	b := &Bot{}
	if b.beginRespawn(20) {
		t.Fatal("living player requested respawn")
	}
	if !b.beginRespawn(0) {
		t.Fatal("death did not request respawn")
	}
	if b.beginRespawn(0) || b.beginRespawn(-1) {
		t.Fatal("duplicate health packets requested duplicate respawns")
	}
	b.respawning.Store(false) // authoritative post-respawn position
	if !b.beginRespawn(0) {
		t.Fatal("a later death did not request respawn")
	}
}
