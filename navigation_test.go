package minego

import (
	"context"
	"math"
	"strings"
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

func TestPathPrefersJumpOverBreakingOrPillaringOneBlockRise(t *testing.T) {
	b := syntheticBot(t)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 64, 1, stone)
	path, err := b.Navigator.Path(BlockPos{1, 64, 1}, GoalBlock(BlockPos{2, 65, 1}), NavigationOptions{AllowBreaking: true, AllowPlacing: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(path) != 2 || path[1].Move != MoveJump {
		t.Fatalf("one-block terrain rise should be jumped, got %#v", path)
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

func TestGroundedUsesCollisionSupportForMiningAndJumping(t *testing.T) {
	b := syntheticBot(t)
	state := SelfState{Position: Vec3{1.5, 64, 1.5}, OnGround: false}
	if !b.Navigator.grounded(state) {
		t.Fatal("supported player was treated as airborne")
	}
}

func TestObstacleJumpTurnsBeforeTakeoff(t *testing.T) {
	input := physicsInput{LookX: 1, Jump: true}
	state := SelfState{Rotation: Rotation{Yaw: 0}, OnGround: true}
	oriented, yaw, _ := orientMovement(input, state, true)
	if yaw != -30 {
		t.Fatalf("first turn yaw = %v, want -30", yaw)
	}
	if oriented.Jump {
		t.Fatal("jump began before the bot faced the obstacle")
	}
	state.Rotation.Yaw = -90
	oriented, _, _ = orientMovement(input, state, true)
	if !oriented.Jump {
		t.Fatal("jump remained suppressed after the bot faced the obstacle")
	}
}

func TestObstacleJumpCentersDepartureBeforeTakeoff(t *testing.T) {
	state := SelfState{Position: Vec3{1.82, 64, 1.18}, OnGround: true}
	input, centered := centerNodeInput(state, BlockPos{1, 64, 1}, .215)
	if centered || input.X >= 0 || input.Z <= 0 {
		t.Fatalf("off-center jump did not steer toward node center: centered=%t input=%#v", centered, input)
	}
	state.Position = Vec3{1.54, 64, 1.46}
	if _, centered = centerNodeInput(state, BlockPos{1, 64, 1}, .215); !centered {
		t.Fatal("centered player was not allowed to take off")
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

func TestDiagonalJumpRequiresSweptHeadClearance(t *testing.T) {
	b := syntheticBot(t)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	start := BlockPos{1, 64, 1}
	target := BlockPos{2, 65, 2}
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 64, 2, stone)
	options := defaultsNav(NavigationOptions{})
	if !hasMoveTo(b.Navigator.neighbors(start, options), target, MoveJump) {
		t.Fatal("clear diagonal rise was not offered to the planner")
	}
	// This ceiling is beside the departure node. It does not occupy the start
	// or landing column, but the player's head sweeps through it diagonally.
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 66, 1, stone)
	if hasMoveTo(b.Navigator.neighbors(start, options), target, MoveJump) {
		t.Fatal("diagonal jump clipped through side ceiling")
	}
}

func TestDropRequiresClearDepartureColumn(t *testing.T) {
	b := syntheticBot(t)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	// Stand two blocks above the ordinary floor and consider dropping into the
	// neighboring column.
	b.World.chunks[chunkKey{0, 0}].SetBlockState(1, 65, 1, stone)
	start := BlockPos{1, 66, 1}
	landing := BlockPos{2, 64, 1}
	options := defaultsNav(NavigationOptions{})
	if !hasMoveTo(b.Navigator.neighbors(start, options), landing, MoveDrop) {
		t.Fatal("clear two-block drop was not planned")
	}
	// The landing itself remains clear, but this block catches the player's
	// head before it can step off the upper ledge.
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 67, 1, stone)
	if hasMoveTo(b.Navigator.neighbors(start, options), landing, MoveDrop) {
		t.Fatal("drop was planned through an obstructed departure column")
	}
}

func hasMoveTo(nodes []PathNode, position BlockPos, move MoveKind) bool {
	for _, node := range nodes {
		if node.Position == position && node.Move == move {
			return true
		}
	}
	return false
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

func TestNavigationCameraTurnsShortestWayAndLevelsPitch(t *testing.T) {
	if got := approachAngle(170, -170, 30); got != 190 {
		t.Fatalf("wrapped yaw = %v, want 190", got)
	}
	if got := approachAngle(-170, 170, 30); got != -190 {
		t.Fatalf("reverse wrapped yaw = %v, want -190", got)
	}
	if got := approach(45, 0, 12); got != 33 {
		t.Fatalf("leveled pitch = %v, want 33", got)
	}
	if got := angleDelta(90, -90); got != -180 {
		t.Fatalf("opposite-direction delta = %v, want -180", got)
	}
}

func TestPathAvoidsServerRejectedBreak(t *testing.T) {
	b := syntheticBot(t)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	rejected := BlockPos{2, 64, 1}
	b.World.chunks[chunkKey{0, 0}].SetBlockState(rejected.X, rejected.Y, rejected.Z, stone)
	b.Miner.reject(rejected)
	path, err := b.Navigator.Path(BlockPos{1, 64, 1}, GoalBlock(BlockPos{3, 64, 1}), NavigationOptions{AllowBreaking: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range path {
		if node.Position == rejected {
			t.Fatalf("path retained rejected protected block: %#v", path)
		}
	}
}

func TestBridgeAndTwoBlockDigActions(t *testing.T) {
	b := syntheticBot(t)
	air, _ := b.pack.StateID("minecraft:air", nil)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 63, 1, air)
	move, _, ok := b.Navigator.passable(BlockPos{2, 64, 1}, NavigationOptions{AllowPlacing: true})
	if !ok || move != MoveBridge {
		t.Fatalf("expected bridge edge, got move=%v ok=%t", move, ok)
	}
	node := b.Navigator.pathNode(BlockPos{2, 64, 1}, move, 4)
	if len(node.Place) != 1 || node.Place[0] != (BlockPos{2, 63, 1}) {
		t.Fatalf("unexpected bridge action: %#v", node)
	}
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 63, 1, stone)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 64, 1, stone)
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 65, 1, stone)
	move, _, ok = b.Navigator.passable(BlockPos{2, 64, 1}, NavigationOptions{AllowBreaking: true})
	if !ok || move != MoveBreak {
		t.Fatalf("expected break edge, got move=%v ok=%t", move, ok)
	}
	node = b.Navigator.pathNode(BlockPos{2, 64, 1}, move, 4)
	if len(node.Break) != 2 {
		t.Fatalf("expected feet and head dig actions, got %#v", node.Break)
	}
}

type customBlockGoal BlockPos

func (g customBlockGoal) Reached(p BlockPos) bool     { return p == BlockPos(g) }
func (g customBlockGoal) Estimate(p BlockPos) float64 { return distance(p, BlockPos(g)) }

func TestAStarPathRetainsBreakActions(t *testing.T) {
	b := syntheticBot(t)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	for z := 0; z < 16; z++ {
		b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 64, z, stone)
		b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 65, z, stone)
	}
	path, err := b.Navigator.Path(BlockPos{1, 64, 1}, customBlockGoal(BlockPos{3, 64, 1}), NavigationOptions{AllowBreaking: true})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, node := range path {
		if len(node.Break) > 0 {
			found = true
		}
	}
	if !found {
		t.Fatalf("A* path discarded break actions: %#v", path)
	}
}

func TestAStarBreakFilterRejectsGroundAndAllowsLeaves(t *testing.T) {
	b := syntheticBot(t)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	leaves, _ := b.pack.StateID("minecraft:spruce_leaves", nil)
	for z := 0; z < 16; z++ {
		b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 64, z, stone)
		b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 65, z, stone)
	}
	filter := func(block Block) bool { return strings.HasSuffix(block.Name, "_leaves") }
	goal := customBlockGoal(BlockPos{3, 64, 1})
	if _, err := b.Navigator.Path(BlockPos{1, 64, 1}, goal, NavigationOptions{AllowBreaking: true, BreakFilter: filter}); err == nil {
		t.Fatal("filtered path broke non-leaf terrain")
	}
	for y := 64; y <= 65; y++ {
		b.World.chunks[chunkKey{0, 0}].SetBlockState(2, y, 1, leaves)
	}
	path, err := b.Navigator.Path(BlockPos{1, 64, 1}, goal, NavigationOptions{AllowBreaking: true, BreakFilter: filter})
	if err != nil {
		t.Fatal(err)
	}
	if len(path) < 2 || len(path[1].Break) == 0 {
		t.Fatalf("leaf opening was not planned as a break: %#v", path)
	}
}

func TestAStarPathRetainsPillarActions(t *testing.T) {
	b := syntheticBot(t)
	path, err := b.Navigator.Path(BlockPos{1, 64, 1}, customBlockGoal(BlockPos{1, 66, 1}), NavigationOptions{AllowPlacing: true})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, node := range path {
		if node.Move == MovePillar && len(node.Place) == 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("A* path discarded pillar placement: %#v", path)
	}
}

func TestDStarRepairsChangedEdge(t *testing.T) {
	b := syntheticBot(t)
	start := BlockPos{1, 64, 1}
	goal := GoalBlock(BlockPos{5, 64, 1})
	planner, ok := b.Navigator.newDStar(start, goal, defaultsNav(NavigationOptions{MaxNodes: 5000}))
	if !ok {
		t.Fatal("standard goal did not create D* planner")
	}
	path, _, err := planner.Plan()
	if err != nil {
		t.Fatal(err)
	}
	if len(path) < 2 {
		t.Fatalf("short path: %#v", path)
	}
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	changed := BlockPos{3, 64, 1}
	b.World.chunks[chunkKey{0, 0}].SetBlockState(changed.X, changed.Y, changed.Z, stone)
	repaired, _, err := planner.Repair(start, []BlockPos{changed})
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range repaired {
		if node.Position == changed {
			t.Fatalf("repaired path retained blocked edge: %#v", repaired)
		}
	}
}
