package minego

import "math"

const (
	playerHalfWidth = 0.3
	playerHeight    = 1.8
)

// physicsInput is the movement intent for one 20 Hz client tick. Vertical
// position is deliberately absent: height is determined by gravity and
// collision, never by the navigator.
type physicsInput struct {
	X, Z         float64
	LookX, LookZ float64
	Jump         bool
	Climb        float64
	Sprint       bool
}

type physicsResult struct {
	Position            Vec3
	Velocity            Vec3
	OnGround            bool
	HorizontalCollision bool
	BlockedX, BlockedZ  bool
}

func (n *Navigator) physicsStep(state SelfState, in physicsInput) physicsResult {
	velocity := state.Velocity
	velocity.X, velocity.Z = in.X, in.Z

	feet, feetLoaded := n.bot.World.Block(state.Position.Block())
	inFluid := feetLoaded && isFluid(feet.Name)
	head, headLoaded := n.bot.World.Block(Vec3{X: state.Position.X, Y: state.Position.Y + 1.62, Z: state.Position.Z}.Block())
	headInFluid := headLoaded && isFluid(head.Name)
	onClimbable := feetLoaded && (stringsContainsClimbable(feet.Name))
	grounded := state.OnGround || n.hasSupport(playerBox(state.Position))

	if onClimbable && in.Climb != 0 {
		velocity.Y = math.Max(-0.15, math.Min(0.2, in.Climb))
	} else if inFluid {
		velocity.Y = (velocity.Y - 0.02) * 0.8
		// Keep the bot breathing even when navigation is idle. Once its head is
		// above the surface it resumes normal input-controlled bobbing.
		if in.Jump || headInFluid {
			velocity.Y += 0.04
		}
		velocity.X *= 0.8
		velocity.Z *= 0.8
	} else if in.Jump && grounded {
		velocity.Y = 0.42
	} else {
		velocity.Y = (velocity.Y - 0.08) * 0.98
	}

	box := playerBox(state.Position)
	wantedX, wantedY, wantedZ := velocity.X, velocity.Y, velocity.Z
	region := box.expanded(wantedX, wantedY, wantedZ).expanded(0, 0.6, 0)
	boxes := n.collisionBoxes(region)
	pushX, pushZ := horizontalDepenetration(box, boxes)
	box = box.offset(pushX, 0, pushZ)
	motion := resolveMotion(box, boxes, wantedX, wantedY, wantedZ)
	if motion.horizontalCollision && wantedX != 0 && wantedZ != 0 {
		// Evaluate both horizontal sweep orders. At a block corner one order can
		// clip the component that would otherwise carry the player along the free
		// face. Vanilla resolves using the more productive horizontal candidate,
		// which gives movement its characteristic wall sliding.
		alternate := resolveMotionZX(box, boxes, wantedX, wantedY, wantedZ)
		if horizontalProgress(alternate, wantedX, wantedZ) > horizontalProgress(motion, wantedX, wantedZ)+1e-9 {
			motion = alternate
		}
	}

	// Vanilla can step onto collision shapes up to 0.6 blocks high. Without
	// this alternate candidate, slabs, stairs, path edges, and water banks all
	// look like walls and the client remains pressed into their side.
	if motion.horizontalCollision && grounded && !in.Jump && !inFluid {
		step := resolveStep(box, boxes, wantedX, wantedY, wantedZ, 0.6)
		if step.X*step.X+step.Z*step.Z > motion.X*motion.X+motion.Z*motion.Z+1e-9 {
			motion = step
		}
	}
	dx, dy, dz := motion.X, motion.Y, motion.Z

	verticalCollision := math.Abs(dy-wantedY) > 1e-9
	blockedX := math.Abs(dx-wantedX) > 1e-9
	blockedZ := math.Abs(dz-wantedZ) > 1e-9
	horizontalCollision := blockedX || blockedZ
	onGround := motion.onGround
	if verticalCollision {
		velocity.Y = 0
	}
	if math.Abs(dx-wantedX) > 1e-9 {
		velocity.X = 0
	}
	if math.Abs(dz-wantedZ) > 1e-9 {
		velocity.Z = 0
	}

	return physicsResult{
		Position: Vec3{state.Position.X + pushX + dx, state.Position.Y + dy, state.Position.Z + pushZ + dz},
		Velocity: velocity, OnGround: onGround, HorizontalCollision: horizontalCollision,
		BlockedX: blockedX, BlockedZ: blockedZ,
	}
}

func horizontalDepenetration(player AABB, obstacles []AABB) (float64, float64) {
	best := math.MaxFloat64
	dx, dz := 0.0, 0.0
	for _, block := range obstacles {
		if !overlaps(player.MinX, player.MaxX, block.MinX, block.MaxX) || !overlaps(player.MinY, player.MaxY, block.MinY, block.MaxY) || !overlaps(player.MinZ, player.MaxZ, block.MinZ, block.MaxZ) {
			continue
		}
		candidates := [...]struct{ x, z float64 }{
			{block.MinX - player.MaxX - 1e-5, 0},
			{block.MaxX - player.MinX + 1e-5, 0},
			{0, block.MinZ - player.MaxZ - 1e-5},
			{0, block.MaxZ - player.MinZ + 1e-5},
		}
		for _, candidate := range candidates {
			distance := math.Abs(candidate.x) + math.Abs(candidate.z)
			if distance < best {
				best, dx, dz = distance, candidate.x, candidate.z
			}
		}
	}
	dx = math.Max(-.2, math.Min(.2, dx))
	dz = math.Max(-.2, math.Min(.2, dz))
	return dx, dz
}

func resolveMotionZX(box AABB, obstacles []AABB, x, y, z float64) resolvedMotion {
	wantedX, wantedY, wantedZ := x, y, z
	for _, obstacle := range obstacles {
		y = clipY(box, obstacle, y)
	}
	box = box.offset(0, y, 0)
	for _, obstacle := range obstacles {
		z = clipZ(box, obstacle, z)
	}
	box = box.offset(0, 0, z)
	for _, obstacle := range obstacles {
		x = clipX(box, obstacle, x)
	}
	return resolvedMotion{
		X: x, Y: y, Z: z,
		onGround:            wantedY < 0 && math.Abs(y-wantedY) > 1e-9,
		horizontalCollision: math.Abs(x-wantedX) > 1e-9 || math.Abs(z-wantedZ) > 1e-9,
	}
}

func horizontalProgress(motion resolvedMotion, wantedX, wantedZ float64) float64 {
	return motion.X*wantedX + motion.Z*wantedZ
}

type resolvedMotion struct {
	X, Y, Z             float64
	onGround            bool
	horizontalCollision bool
}

func resolveMotion(box AABB, obstacles []AABB, x, y, z float64) resolvedMotion {
	wantedX, wantedY, wantedZ := x, y, z
	for _, obstacle := range obstacles {
		y = clipY(box, obstacle, y)
	}
	box = box.offset(0, y, 0)
	for _, obstacle := range obstacles {
		x = clipX(box, obstacle, x)
	}
	box = box.offset(x, 0, 0)
	for _, obstacle := range obstacles {
		z = clipZ(box, obstacle, z)
	}
	return resolvedMotion{
		X: x, Y: y, Z: z,
		onGround:            wantedY < 0 && math.Abs(y-wantedY) > 1e-9,
		horizontalCollision: math.Abs(x-wantedX) > 1e-9 || math.Abs(z-wantedZ) > 1e-9,
	}
}

func resolveStep(box AABB, obstacles []AABB, x, y, z, height float64) resolvedMotion {
	wantedX, wantedY, wantedZ := x, y, z
	up := height
	for _, obstacle := range obstacles {
		up = clipY(box, obstacle, up)
	}
	stepped := box.offset(0, up, 0)
	for _, obstacle := range obstacles {
		x = clipX(stepped, obstacle, x)
	}
	stepped = stepped.offset(x, 0, 0)
	for _, obstacle := range obstacles {
		z = clipZ(stepped, obstacle, z)
	}
	stepped = stepped.offset(0, 0, z)
	downWanted := wantedY - up
	down := downWanted
	for _, obstacle := range obstacles {
		down = clipY(stepped, obstacle, down)
	}
	totalY := up + down
	return resolvedMotion{
		X: x, Y: totalY, Z: z,
		onGround:            downWanted < 0 && math.Abs(down-downWanted) > 1e-9,
		horizontalCollision: math.Abs(x-wantedX) > 1e-9 || math.Abs(z-wantedZ) > 1e-9,
	}
}

func (n *Navigator) hasSupport(box AABB) bool {
	const probe = -0.001
	for _, obstacle := range n.collisionBoxes(box.expanded(0, probe, 0)) {
		if clipY(box, obstacle, probe) > probe {
			return true
		}
	}
	return false
}

func stringsContainsClimbable(name string) bool {
	return len(name) != 0 && (contains(name, "ladder") || contains(name, "vine"))
}

// contains is kept local to avoid making the collision hot path allocate.
func contains(s, sub string) bool {
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func playerBox(p Vec3) AABB {
	return AABB{p.X - playerHalfWidth, p.Y, p.Z - playerHalfWidth, p.X + playerHalfWidth, p.Y + playerHeight, p.Z + playerHalfWidth}
}

func (b AABB) offset(x, y, z float64) AABB {
	return AABB{b.MinX + x, b.MinY + y, b.MinZ + z, b.MaxX + x, b.MaxY + y, b.MaxZ + z}
}

func (b AABB) expanded(x, y, z float64) AABB {
	r := b
	if x < 0 {
		r.MinX += x
	} else {
		r.MaxX += x
	}
	if y < 0 {
		r.MinY += y
	} else {
		r.MaxY += y
	}
	if z < 0 {
		r.MinZ += z
	} else {
		r.MaxZ += z
	}
	return r
}

func (n *Navigator) collisionBoxes(region AABB) []AABB {
	minX, maxX := int(math.Floor(region.MinX)), int(math.Floor(region.MaxX-1e-9))
	minY, maxY := int(math.Floor(region.MinY)), int(math.Floor(region.MaxY-1e-9))
	minZ, maxZ := int(math.Floor(region.MinZ)), int(math.Floor(region.MaxZ-1e-9))
	result := make([]AABB, 0, 16)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				pos := BlockPos{x, y, z}
				block, ok := n.bot.World.Block(pos)
				if !ok {
					continue
				}
				for _, shape := range block.Collision {
					result = append(result, shape.offset(float64(x), float64(y), float64(z)))
				}
			}
		}
	}
	return result
}

func overlaps(a1, a2, b1, b2 float64) bool { return a2 > b1 && a1 < b2 }

func clipY(player, block AABB, delta float64) float64 {
	if !overlaps(player.MinX, player.MaxX, block.MinX, block.MaxX) || !overlaps(player.MinZ, player.MaxZ, block.MinZ, block.MaxZ) {
		return delta
	}
	const epsilon = 1e-5
	if delta > 0 && player.MaxY <= block.MinY+epsilon {
		return math.Min(delta, block.MinY-player.MaxY)
	}
	if delta < 0 && player.MinY >= block.MaxY-epsilon {
		return math.Max(delta, block.MaxY-player.MinY)
	}
	return delta
}
func clipX(player, block AABB, delta float64) float64 {
	if !overlaps(player.MinY, player.MaxY, block.MinY, block.MaxY) || !overlaps(player.MinZ, player.MaxZ, block.MinZ, block.MaxZ) {
		return delta
	}
	const epsilon = 1e-5
	if delta > 0 && player.MaxX <= block.MinX+epsilon {
		return math.Min(delta, block.MinX-player.MaxX)
	}
	if delta < 0 && player.MinX >= block.MaxX-epsilon {
		return math.Max(delta, block.MaxX-player.MinX)
	}
	return delta
}
func clipZ(player, block AABB, delta float64) float64 {
	if !overlaps(player.MinX, player.MaxX, block.MinX, block.MaxX) || !overlaps(player.MinY, player.MaxY, block.MinY, block.MaxY) {
		return delta
	}
	const epsilon = 1e-5
	if delta > 0 && player.MaxZ <= block.MinZ+epsilon {
		return math.Min(delta, block.MinZ-player.MaxZ)
	}
	if delta < 0 && player.MinZ >= block.MaxZ-epsilon {
		return math.Max(delta, block.MaxZ-player.MinZ)
	}
	return delta
}
