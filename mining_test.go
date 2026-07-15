package minego

import (
	"io"
	"log/slog"
	"testing"
	"time"

	jp "github.com/zeozeozeo/minego/internal/protocol/java_protocol"
)

func TestBreakDuration(t *testing.T) {
	b := Block{Hardness: 3, RequiresCorrectTool: true}
	if d := breakDuration(b, 6, SelfState{OnGround: true, Effects: map[int32]Effect{}}); d != 800*time.Millisecond {
		t.Fatalf("duration %v", d)
	}
}

func TestBreakDurationAccountsForAirbornePenalty(t *testing.T) {
	b := Block{Hardness: .2}
	grounded := breakDuration(b, 1, SelfState{OnGround: true, Effects: map[int32]Effect{}})
	airborne := breakDuration(b, 1, SelfState{Effects: map[int32]Effect{}})
	if grounded != 400*time.Millisecond || airborne != 1600*time.Millisecond {
		t.Fatalf("grounded=%v airborne=%v", grounded, airborne)
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
	if g.Reached(BlockPos{12, 64, 10}) {
		t.Fatal("dig goal accepted a position too far away to reliably collect drops")
	}
}

func TestRejectedBlocksRequireRepeatedEvidenceBeforeRejectingChunk(t *testing.T) {
	b := syntheticBot(t)
	positions := []BlockPos{{32, 64, 48}, {33, 64, 48}, {34, 64, 48}}
	b.Miner.reject(positions[0])
	if !b.Miner.breakRejected(positions[0]) {
		t.Fatal("explicitly rejected block was forgotten")
	}
	if b.Miner.breakRejected(positions[1]) {
		t.Fatal("one rejection discarded the whole chunk")
	}
	b.Miner.reject(positions[1])
	if b.Miner.breakRejected(positions[2]) {
		t.Fatal("two rejections discarded the whole chunk")
	}
	b.Miner.reject(positions[2])
	if !b.Miner.breakRejected(BlockPos{35, 70, 49}) {
		t.Fatal("three distinct rejections did not identify a protected chunk")
	}
	if !b.Miner.breakRejected(BlockPos{48, 70, 48}) {
		t.Fatal("protected-area evidence did not cover the neighboring spawn-protection blocks")
	}
	if b.Miner.breakRejected(BlockPos{64, 70, 48}) {
		t.Fatal("protected-area evidence spread beyond its conservative radius")
	}
}

func TestRejectedBlockDoesNotCountTwice(t *testing.T) {
	b := syntheticBot(t)
	first := BlockPos{48, 64, 64}
	b.Miner.reject(first)
	b.Miner.reject(first)
	b.Miner.reject(first)
	if b.Miner.breakRejected(BlockPos{49, 64, 64}) {
		t.Fatal("retries of one block identified the whole chunk as protected")
	}
}

func TestNewLoadedChunksTracksIdentityNotCount(t *testing.T) {
	before := loadedChunkSet([][2]int32{{0, 0}, {1, 0}})
	if added := newLoadedChunks(before, [][2]int32{{1, 0}, {2, 0}}); added != 1 {
		t.Fatalf("new chunks = %d, want 1", added)
	}
}

func TestMiningTargetScoreKeepsCurrentTreeTogether(t *testing.T) {
	self := Vec3{10.5, 64, 10.5}
	cluster := []BlockPos{{10, 64, 10}}
	sameTree := Block{Position: BlockPos{10, 67, 10}}
	otherTree := Block{Position: BlockPos{12, 64, 10}}
	if miningTargetScore(self, sameTree, cluster) >= miningTargetScore(self, otherTree, cluster) {
		t.Fatal("nearby second trunk interrupted current connected tree")
	}
}

func TestUnsupportedEntityMetadataIsNonFatal(t *testing.T) {
	b := syntheticBot(t)
	b.cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	// Entity ID 1, then metadata index 38 with an unsupported serializer 85.
	// The packet is already framed, so dropping this observational update is
	// safe and must not disconnect the play session.
	wire := &jp.WirePacket{Data: []byte{1, 38, 85, 0}}
	if err := b.handleEntityData(wire); err != nil {
		t.Fatalf("unsupported metadata was fatal: %v", err)
	}
}
