package minego

import "sync"

type event[T any] struct {
	mu       sync.RWMutex
	next     uint64
	handlers map[uint64]func(T)
}

func (e *event[T]) subscribe(fn func(T)) func() {
	e.mu.Lock()
	if e.handlers == nil {
		e.handlers = map[uint64]func(T){}
	}
	e.next++
	id := e.next
	e.handlers[id] = fn
	e.mu.Unlock()
	var once sync.Once
	return func() { once.Do(func() { e.mu.Lock(); delete(e.handlers, id); e.mu.Unlock() }) }
}
func (e *event[T]) emit(v T) {
	e.mu.RLock()
	fns := make([]func(T), 0, len(e.handlers))
	for _, fn := range e.handlers {
		fns = append(fns, fn)
	}
	e.mu.RUnlock()
	for _, fn := range fns {
		fn(v)
	}
}
