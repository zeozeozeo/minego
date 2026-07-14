package minego

import (
	"context"
	"sync"
)

type actionControl uint8

const (
	controlMovement actionControl = 1 << iota
	controlView
	controlHands
	controlInventory
	controlWindows
)

type actionPriority uint8

const (
	priorityBackground actionPriority = iota
	priorityAutomation
	priorityExplicit
)

type actionLease struct {
	coordinator *actionCoordinator
	controls    actionControl
	priority    actionPriority
	cancel      context.CancelFunc
	done        chan struct{}
	once        sync.Once
}

func (l *actionLease) Context(parent context.Context) context.Context {
	ctx, cancel := context.WithCancel(parent)
	go func() {
		select {
		case <-l.done:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx
}

func (l *actionLease) Release() {
	l.once.Do(func() {
		l.cancel()
		close(l.done)
		l.coordinator.release(l)
	})
}

type actionCoordinator struct {
	mu      sync.Mutex
	active  map[*actionLease]struct{}
	changed chan struct{}
}

func newActionCoordinator() *actionCoordinator {
	return &actionCoordinator{active: make(map[*actionLease]struct{}), changed: make(chan struct{})}
}

// acquire atomically claims all controls. Higher-priority actions preempt every
// conflicting lower-priority lease; otherwise acquisition waits for release.
func (c *actionCoordinator) acquire(ctx context.Context, controls actionControl, priority actionPriority) (*actionLease, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		c.mu.Lock()
		blocked := false
		var victims []*actionLease
		for lease := range c.active {
			if lease.controls&controls == 0 {
				continue
			}
			if priority > lease.priority {
				victims = append(victims, lease)
				continue
			}
			blocked = true
		}
		if !blocked {
			for _, victim := range victims {
				delete(c.active, victim)
			}
			if len(victims) > 0 {
				close(c.changed)
				c.changed = make(chan struct{})
			}
			_, cancel := context.WithCancel(context.Background())
			lease := &actionLease{coordinator: c, controls: controls, priority: priority, cancel: cancel, done: make(chan struct{})}
			c.active[lease] = struct{}{}
			c.mu.Unlock()
			for _, victim := range victims {
				victim.Release()
			}
			go func() {
				select {
				case <-ctx.Done():
					lease.Release()
				case <-lease.done:
				}
			}()
			return lease, nil
		}
		changed := c.changed
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-changed:
		}
	}
}

func (c *actionCoordinator) release(lease *actionLease) {
	c.mu.Lock()
	if _, ok := c.active[lease]; ok {
		delete(c.active, lease)
		close(c.changed)
		c.changed = make(chan struct{})
	}
	c.mu.Unlock()
}

func (c *actionCoordinator) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.active)
}
