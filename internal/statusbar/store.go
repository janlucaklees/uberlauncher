package statusbar

import "sync"

type Store struct {
	mu      sync.RWMutex
	values  map[string]string
	Updates chan struct{}
}

func NewStore() *Store {
	return &Store{
		values:  make(map[string]string),
		Updates: make(chan struct{}, 1),
	}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	s.values[key] = value
	s.mu.Unlock()
	select {
	case s.Updates <- struct{}{}:
	default:
	}
}

func (s *Store) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.values[key]
}
