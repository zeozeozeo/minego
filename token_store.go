package minego

import (
	"errors"
	"sort"
	"sync"
)

type memoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*CachedSession
}

func newMemoryStore() *memoryStore { return &memoryStore{sessions: map[string]*CachedSession{}} }
func (s *memoryStore) SaveSession(v *CachedSession) error {
	if v == nil || v.Username == "" {
		return errors.New("session username is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c := *v
	s.sessions[v.Username] = &c
	return nil
}
func (s *memoryStore) LoadSession(name string) (*CachedSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v := s.sessions[name]
	if v == nil {
		return nil, nil
	}
	c := *v
	return &c, nil
}
func (s *memoryStore) Clear(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, name)
	return nil
}
func (s *memoryStore) ListAccounts() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.sessions))
	for k := range s.sessions {
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}
