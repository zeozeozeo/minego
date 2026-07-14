package minego

import (
	"testing"
	"time"
)

func TestBreakDuration(t *testing.T) {
	b := Block{Hardness: 3, RequiresCorrectTool: true}
	if d := breakDuration(b, 6, SelfState{Effects: map[int32]Effect{}}); d != 750*time.Millisecond {
		t.Fatalf("duration %v", d)
	}
}
func TestSelectorTags(t *testing.T) {
	b := syntheticBot(t)
	found := b.Miner.search(Tags("minecraft:mineable/pickaxe"), map[BlockPos]bool{})
	if len(found) == 0 {
		t.Fatal("expected tagged stone")
	}
}

func TestDigGoalNeverAcceptsStandingOnTarget(t *testing.T) {
	g := digGoal{BlockPos{10, 64, 10}}
	if g.Reached(BlockPos{10, 65, 10}) {
		t.Fatal("dig goal accepted the block directly below the player")
	}
	if !g.Reached(BlockPos{11, 65, 10}) {
		t.Fatal("dig goal rejected a safe neighboring face")
	}
}
