package minego

import (
	"context"
	"math"
	"testing"

	"github.com/zeozeozeo/minego/internal/data/chunks"
)

func syntheticBot(t *testing.T) *Bot {
	b, err := New(Config{Address: "localhost", Version: "26.2", Auth: Offline("test")})
	if err != nil {
		t.Fatal(err)
	}
	col := &chunks.ChunkColumn{X: 0, Z: 0}
	for i := range chunks.SectionCount {
		col.Sections[i] = chunks.NewEmptySection()
	}
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			col.SetBlockState(x, 63, z, stone)
		}
	}
	b.World.chunks[chunkKey{0, 0}] = col
	return b
}
func TestPathAroundObstacle(t *testing.T) {
	b := syntheticBot(t)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 64, 1, stone)
	p, err := b.Navigator.Path(BlockPos{1, 64, 1}, GoalBlock(BlockPos{3, 64, 1}), NavigationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(p) < 3 || p[1].Move != MoveJump {
		t.Fatalf("expected a legal jump or detour, got %#v", p)
	}
}
func TestPathCancellation(t *testing.T) {
	b := syntheticBot(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := b.Navigator.Navigate(ctx, GoalBlock(BlockPos{2, 64, 2}), NavigationOptions{})
	if err == nil {
		t.Fatal("expected cancellation")
	}
}

func TestPhysicsStandsOnAuthoritativeCollision(t *testing.T) {
	b := syntheticBot(t)
	state := SelfState{Position: Vec3{1.5, 64, 1.5}, OnGround: true}
	got := b.Navigator.physicsStep(state, physicsInput{})
	if got.Position.Y != 64 || !got.OnGround || got.Velocity.Y != 0 {
		t.Fatalf("standing player moved through support: %#v", got)
	}
}

func TestPhysicsFallsWhenSupportIsRemoved(t *testing.T) {
	b := syntheticBot(t)
	air, _ := b.pack.StateID("minecraft:air", nil)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(1, 63, 1, air)
	state := SelfState{Position: Vec3{1.5, 64, 1.5}, OnGround: true}
	got := b.Navigator.physicsStep(state, physicsInput{})
	if got.Position.Y >= state.Position.Y || got.OnGround || got.Velocity.Y >= 0 {
		t.Fatalf("unsupported player did not fall: %#v", got)
	}
}

func TestPhysicsClipsWall(t *testing.T) {
	b := syntheticBot(t)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 64, 1, stone)
	state := SelfState{Position: Vec3{1.5, 64, 1.5}, OnGround: true}
	got := b.Navigator.physicsStep(state, physicsInput{X: .4})
	if got.Position.X > 1.7000001 || !got.HorizontalCollision {
		t.Fatalf("player crossed wall boundary: %#v", got)
	}
}

func TestPhysicsJumpUsesCollisionSupport(t *testing.T) {
	b := syntheticBot(t)
	// The first authoritative spawn packet does not carry the regular
	// movement on-ground bit. Collision support must still permit a jump.
	state := SelfState{Position: Vec3{1.5, 64, 1.5}, OnGround: false}
	got := b.Navigator.physicsStep(state, physicsInput{Jump: true})
	if got.Position.Y <= state.Position.Y || got.Velocity.Y <= 0 {
		t.Fatalf("supported player failed to jump: %#v", got)
	}
}

func TestPhysicsSwimsUpWhenSubmerged(t *testing.T) {
	b := syntheticBot(t)
	water, _ := b.pack.StateID("minecraft:water", nil)
	for y := 64; y <= 66; y++ {
		b.World.chunks[chunkKey{0, 0}].SetBlockState(1, y, 1, water)
	}
	state := SelfState{Position: Vec3{1.5, 64, 1.5}}
	got := b.Navigator.physicsStep(state, physicsInput{})
	if got.Velocity.Y <= 0 || got.Position.Y <= state.Position.Y {
		t.Fatalf("submerged player did not seek air: %#v", got)
	}
}

func TestPhysicsRecoversTinyFloorOverlap(t *testing.T) {
	b := syntheticBot(t)
	state := SelfState{Position: Vec3{1.5, 63.999999, 1.5}}
	got := b.Navigator.physicsStep(state, physicsInput{})
	if math.Abs(got.Position.Y-64) > 1e-9 || !got.OnGround {
		t.Fatalf("player was not depenetrated onto floor: %#v", got)
	}
}

func TestPhysicsStepsOntoBottomSlab(t *testing.T) {
	b := syntheticBot(t)
	slab, ok := b.pack.StateID("minecraft:stone_slab", map[string]string{"type": "bottom", "waterlogged": "false"})
	if !ok {
		t.Fatal("bottom slab state missing")
	}
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 64, 1, slab)
	state := SelfState{Position: Vec3{1.5, 64, 1.5}, OnGround: true}
	got := b.Navigator.physicsStep(state, physicsInput{X: .4})
	if got.Position.X < 1.89 || math.Abs(got.Position.Y-64.5) > 1e-9 || !got.OnGround {
		t.Fatalf("player did not step onto slab: %#v", got)
	}
}

func TestPhysicsJumpClearsFullBlock(t *testing.T) {
	b := syntheticBot(t)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 64, 1, stone)
	state := SelfState{Position: Vec3{1.5, 64, 1.5}, OnGround: true}
	maxX, maxY := state.Position.X, state.Position.Y
	for range 20 {
		got := b.Navigator.physicsStep(state, physicsInput{X: .18, Jump: true})
		state.Position, state.Velocity, state.OnGround = got.Position, got.Velocity, got.OnGround
		maxX, maxY = math.Max(maxX, got.Position.X), math.Max(maxY, got.Position.Y)
	}
	if maxY < 65 || maxX < 2.3 {
		t.Fatalf("jump never cleared full block: maxX=%v maxY=%v state=%#v", maxX, maxY, state)
	}
}

func TestPhysicsSwimsOntoOneBlockBank(t *testing.T) {
	b := syntheticBot(t)
	water, _ := b.pack.StateID("minecraft:water", nil)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(1, 64, 1, water)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(1, 65, 1, water)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 64, 1, stone)
	state := SelfState{Position: Vec3{1.5, 64, 1.5}}
	maxX := state.Position.X
	for range 40 {
		got := b.Navigator.physicsStep(state, physicsInput{X: .13, Jump: true})
		state.Position, state.Velocity, state.OnGround = got.Position, got.Velocity, got.OnGround
		maxX = math.Max(maxX, got.Position.X)
	}
	if maxX < 2.3 {
		t.Fatalf("swimmer never exited onto bank: maxX=%v state=%#v", maxX, state)
	}
}

func TestPathPlansWaterExitOntoBank(t *testing.T) {
	b := syntheticBot(t)
	water, _ := b.pack.StateID("minecraft:water", nil)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(1, 64, 1, water)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 64, 1, stone)
	path, err := b.Navigator.Path(BlockPos{1, 64, 1}, GoalBlock(BlockPos{2, 65, 1}), NavigationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(path) != 2 || path[1].Move != MoveJump {
		t.Fatalf("expected swim-to-bank jump, got %#v", path)
	}
}

func TestPathUsesDiagonalMovement(t *testing.T) {
	b := syntheticBot(t)
	path, err := b.Navigator.Path(BlockPos{1, 64, 1}, GoalBlock(BlockPos{4, 64, 4}), NavigationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(path) != 4 {
		t.Fatalf("expected three diagonal moves, got %#v", path)
	}
}

func TestPathDoesNotCutBlockedDiagonalCorner(t *testing.T) {
	b := syntheticBot(t)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 64, 1, stone)
	path, err := b.Navigator.Path(BlockPos{1, 64, 1}, GoalBlock(BlockPos{2, 64, 2}), NavigationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(path) == 2 {
		t.Fatalf("path cut through a blocked corner: %#v", path)
	}
}

func TestPathPlansBoundedParkourGap(t *testing.T) {
	b := syntheticBot(t)
	air, _ := b.pack.StateID("minecraft:air", nil)
	for x := 2; x <= 3; x++ {
		b.World.chunks[chunkKey{0, 0}].SetBlockState(x, 63, 1, air)
	}
	path, err := b.Navigator.Path(BlockPos{1, 64, 1}, GoalBlock(BlockPos{4, 64, 1}), NavigationOptions{AllowParkour: true, MaxParkourGap: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(path) != 2 || path[1].Move != MoveParkour {
		t.Fatalf("expected direct parkour edge, got %#v", path)
	}
}

func TestPhysicsSprintJumpCrossesTwoBlockGap(t *testing.T) {
	b := syntheticBot(t)
	air, _ := b.pack.StateID("minecraft:air", nil)
	for x := 2; x <= 3; x++ {
		b.World.chunks[chunkKey{0, 0}].SetBlockState(x, 63, 1, air)
	}
	state := SelfState{Position: Vec3{1.5, 64, 1.5}, OnGround: true}
	for tick := 0; tick < 20; tick++ {
		got := b.Navigator.physicsStep(state, physicsInput{X: .28, Jump: tick == 0, Sprint: true})
		state.Position, state.Velocity, state.OnGround = got.Position, got.Velocity, got.OnGround
	}
	if state.Position.X < 4.3 {
		t.Fatalf("sprint jump failed to cross gap: %#v", state)
	}
}
