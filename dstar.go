package minego

import (
	"container/heap"
	"math"
)

type dstarKey struct{ first, second float64 }
type dstarEntry struct {
	pos    BlockPos
	key    dstarKey
	serial uint64
}
type dstarHeap []dstarEntry

func (h dstarHeap) Len() int { return len(h) }
func (h dstarHeap) Less(i, j int) bool {
	return keyLess(h[i].key, h[j].key) || (!keyLess(h[j].key, h[i].key) && h[i].serial < h[j].serial)
}
func (h dstarHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *dstarHeap) Push(v any)   { *h = append(*h, v.(dstarEntry)) }
func (h *dstarHeap) Pop() any     { old := *h; v := old[len(old)-1]; *h = old[:len(old)-1]; return v }

type dstarPlanner struct {
	nav         *Navigator
	goal        Goal
	opt         NavigationOptions
	start, last BlockPos
	km          float64
	g, rhs      map[BlockPos]float64
	goals       map[BlockPos]bool
	open        dstarHeap
	queued      map[BlockPos]dstarKey
	serial      uint64
}

func (n *Navigator) newDStar(start BlockPos, goal Goal, opt NavigationOptions) (*dstarPlanner, bool) {
	goals, ok := n.goalPositions(goal, opt)
	if !ok || len(goals) == 0 {
		return nil, false
	}
	p := &dstarPlanner{nav: n, goal: goal, opt: opt, start: start, last: start, g: make(map[BlockPos]float64), rhs: make(map[BlockPos]float64), goals: make(map[BlockPos]bool), queued: make(map[BlockPos]dstarKey)}
	heap.Init(&p.open)
	for _, position := range goals {
		p.goals[position] = true
		p.rhs[position] = 0
		p.enqueue(position)
	}
	return p, true
}
func (n *Navigator) goalPositions(goal Goal, opt NavigationOptions) ([]BlockPos, bool) {
	var candidates []BlockPos
	switch g := goal.(type) {
	case GoalBlock:
		candidates = []BlockPos{BlockPos(g)}
	case GoalAdjacent:
		q := BlockPos(g)
		candidates = []BlockPos{{q.X + 1, q.Y, q.Z}, {q.X - 1, q.Y, q.Z}, {q.X, q.Y + 1, q.Z}, {q.X, q.Y - 1, q.Z}, {q.X, q.Y, q.Z + 1}, {q.X, q.Y, q.Z - 1}}
	case GoalNear:
		r := int(math.Ceil(g.Radius))
		if r > 8 {
			return nil, false
		}
		for x := -r; x <= r; x++ {
			for y := -r; y <= r; y++ {
				for z := -r; z <= r; z++ {
					p := BlockPos{g.Position.X + x, g.Position.Y + y, g.Position.Z + z}
					if g.Reached(p) {
						candidates = append(candidates, p)
					}
				}
			}
		}
	default:
		return nil, false
	}
	out := candidates[:0]
	for _, p := range candidates {
		if goal.Reached(p) {
			if _, _, ok := n.passable(p, opt); ok {
				out = append(out, p)
			}
		}
	}
	return out, true
}
func (p *dstarPlanner) value(m map[BlockPos]float64, pos BlockPos) float64 {
	if v, ok := m[pos]; ok {
		return v
	}
	return math.Inf(1)
}
func (p *dstarPlanner) key(pos BlockPos) dstarKey {
	m := math.Min(p.value(p.g, pos), p.value(p.rhs, pos))
	return dstarKey{m + distance(p.start, pos) + p.km, m}
}
func keyLess(a, b dstarKey) bool {
	return a.first < b.first-1e-9 || (math.Abs(a.first-b.first) <= 1e-9 && a.second < b.second-1e-9)
}
func keyEqual(a, b dstarKey) bool {
	return math.Abs(a.first-b.first) <= 1e-9 && math.Abs(a.second-b.second) <= 1e-9
}
func (p *dstarPlanner) enqueue(pos BlockPos) {
	key := p.key(pos)
	p.serial++
	p.queued[pos] = key
	heap.Push(&p.open, dstarEntry{pos: pos, key: key, serial: p.serial})
}
func (p *dstarPlanner) remove(pos BlockPos) { delete(p.queued, pos) }
func (p *dstarPlanner) top() (dstarEntry, bool) {
	for p.open.Len() > 0 {
		entry := p.open[0]
		key, ok := p.queued[entry.pos]
		if !ok || !keyEqual(key, entry.key) {
			heap.Pop(&p.open)
			continue
		}
		return entry, true
	}
	return dstarEntry{}, false
}
func (p *dstarPlanner) pop() (dstarEntry, bool) {
	entry, ok := p.top()
	if !ok {
		return dstarEntry{}, false
	}
	heap.Pop(&p.open)
	delete(p.queued, entry.pos)
	return entry, true
}
func (p *dstarPlanner) successors(pos BlockPos) []PathNode { return p.nav.neighbors(pos, p.opt) }
func (p *dstarPlanner) predecessors(target BlockPos) []BlockPos {
	horizontal := 1
	if p.opt.AllowParkour {
		horizontal = max(horizontal, p.opt.MaxParkourGap+1)
	}
	vertical := max(2, p.opt.MaxDrop)
	seen := make(map[BlockPos]bool)
	var out []BlockPos
	for x := -horizontal; x <= horizontal; x++ {
		for y := -vertical; y <= vertical; y++ {
			for z := -horizontal; z <= horizontal; z++ {
				candidate := BlockPos{target.X + x, target.Y + y, target.Z + z}
				if seen[candidate] || !p.nav.bot.World.IsLoaded(candidate) {
					continue
				}
				for _, edge := range p.successors(candidate) {
					if edge.Position == target {
						seen[candidate] = true
						out = append(out, candidate)
						break
					}
				}
			}
		}
	}
	return out
}
func (p *dstarPlanner) update(pos BlockPos) {
	if p.goals[pos] {
		if _, _, ok := p.nav.passable(pos, p.opt); ok {
			p.rhs[pos] = 0
		} else {
			delete(p.rhs, pos)
		}
	} else {
		best := math.Inf(1)
		for _, edge := range p.successors(pos) {
			best = math.Min(best, edge.Cost+p.value(p.g, edge.Position))
		}
		if math.IsInf(best, 1) {
			delete(p.rhs, pos)
		} else {
			p.rhs[pos] = best
		}
	}
	p.remove(pos)
	if math.Abs(p.value(p.g, pos)-p.value(p.rhs, pos)) > 1e-9 {
		p.enqueue(pos)
	}
}
func (p *dstarPlanner) compute() (int, error) {
	expanded := 0
	for {
		top, ok := p.top()
		startKey := p.key(p.start)
		if (!ok || !keyLess(top.key, startKey)) && math.Abs(p.value(p.rhs, p.start)-p.value(p.g, p.start)) <= 1e-9 {
			break
		}
		if !ok {
			return expanded, ErrUnreachable
		}
		if expanded >= p.opt.MaxNodes {
			return expanded, ErrUnreachable
		}
		u, _ := p.pop()
		newKey := p.key(u.pos)
		if keyLess(u.key, newKey) {
			p.enqueue(u.pos)
		} else if p.value(p.g, u.pos) > p.value(p.rhs, u.pos) {
			p.g[u.pos] = p.value(p.rhs, u.pos)
			for _, pred := range p.predecessors(u.pos) {
				p.update(pred)
			}
		} else {
			delete(p.g, u.pos)
			p.update(u.pos)
			for _, pred := range p.predecessors(u.pos) {
				p.update(pred)
			}
		}
		expanded++
	}
	if math.IsInf(p.value(p.g, p.start), 1) {
		return expanded, ErrUnreachable
	}
	return expanded, nil
}
func (p *dstarPlanner) path() (path []PathNode, err error) {
	position := p.start
	path = []PathNode{{Position: position}}
	seen := map[BlockPos]bool{position: true}
	cost := 0.0
	for !p.goal.Reached(position) {
		best := math.Inf(1)
		var chosen PathNode
		found := false
		for _, edge := range p.successors(position) {
			score := edge.Cost + p.value(p.g, edge.Position)
			if score < best {
				best = score
				chosen = edge
				found = true
			}
		}
		if !found || math.IsInf(best, 1) || seen[chosen.Position] {
			return nil, ErrUnreachable
		}
		cost += chosen.Cost
		chosen.Cost = cost
		path = append(path, chosen)
		position = chosen.Position
		seen[position] = true
		if len(path) > p.opt.MaxNodes {
			return nil, ErrUnreachable
		}
	}
	return path, nil
}
func (p *dstarPlanner) Plan() (path []PathNode, expanded int, err error) {
	expanded, err = p.compute()
	if err != nil {
		return nil, expanded, err
	}
	path, err = p.path()
	return
}
func (p *dstarPlanner) Repair(start BlockPos, changes []BlockPos) (path []PathNode, expanded int, err error) {
	p.km += distance(p.last, start)
	p.last = start
	p.start = start
	affected := make(map[BlockPos]bool)
	for _, change := range changes {
		for x := -2; x <= 2; x++ {
			for y := -2; y <= 2; y++ {
				for z := -2; z <= 2; z++ {
					affected[BlockPos{change.X + x, change.Y + y, change.Z + z}] = true
				}
			}
		}
	}
	for pos := range affected {
		p.update(pos)
		for _, pred := range p.predecessors(pos) {
			p.update(pred)
		}
	}
	return p.Plan()
}
