package minego

import "sync"

type Entity struct {
	ID       int32
	UUID     string
	Type     string
	Position Vec3
	Velocity Vec3
	Rotation Rotation
	Metadata map[int32]any
}
type Entities struct {
	mu       sync.RWMutex
	values   map[int32]Entity
	onChange event[EntityChange]
}

func newEntities() *Entities { return &Entities{values: map[int32]Entity{}} }
func (e *Entities) Get(id int32) (Entity, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	v, ok := e.values[id]
	return cloneEntity(v), ok
}
func (e *Entities) All() []Entity {
	e.mu.RLock()
	defer e.mu.RUnlock()
	r := make([]Entity, 0, len(e.values))
	for _, v := range e.values {
		r = append(r, cloneEntity(v))
	}
	return r
}
func (e *Entities) OnChange(fn func(EntityChange)) func() { return e.onChange.subscribe(fn) }
func cloneEntity(v Entity) Entity {
	if v.Metadata != nil {
		m := make(map[int32]any, len(v.Metadata))
		for k, x := range v.Metadata {
			m[k] = x
		}
		v.Metadata = m
	}
	return v
}
