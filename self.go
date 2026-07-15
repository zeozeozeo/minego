package minego

import "sync"

type Effect struct {
	ID        int32
	Amplifier int32
	Duration  int32
	Flags     int8
}
type SelfState struct {
	EntityID     int32
	Position     Vec3
	Velocity     Vec3
	Rotation     Rotation
	OnGround     bool
	Health       float32
	Food         int32
	Saturation   float32
	GameMode     uint8
	SelectedSlot int
	Flying       bool
	CanFly       bool
	FlyingSpeed  float32
	VehicleID    int32
	Effects      map[int32]Effect
}
type Self struct {
	mu       sync.RWMutex
	state    SelfState
	onChange event[SelfState]
}

func newSelf() *Self { s := &Self{}; s.state.Effects = map[int32]Effect{}; return s }
func (s *Self) State() SelfState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v := s.state
	v.Effects = cloneEffects(v.Effects)
	return v
}
func (s *Self) OnChange(fn func(SelfState)) func() { return s.onChange.subscribe(fn) }
func cloneEffects(v map[int32]Effect) map[int32]Effect {
	r := make(map[int32]Effect, len(v))
	for k, x := range v {
		r[k] = x
	}
	return r
}
func (s *Self) update(fn func(*SelfState)) {
	s.mu.Lock()
	fn(&s.state)
	v := s.state
	v.Effects = cloneEffects(v.Effects)
	s.mu.Unlock()
	s.onChange.emit(v)
}
