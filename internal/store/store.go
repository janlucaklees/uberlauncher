package store

import (
	"sync"

	"uberlauncher/internal/engines"
	"uberlauncher/internal/entry"
)

type Store struct {
	mu      sync.RWMutex
	entries map[string]entry.Entry
	Engine  engines.Engine
}

func New(engine engines.Engine) *Store {
	return &Store{
		entries: make(map[string]entry.Entry),
		Engine:  engine,
	}
}

func (s *Store) UpsertEntry(e entry.Entry) {
	s.mu.Lock()
	s.entries[e.Id()] = e
	s.mu.Unlock()
}

func (s *Store) GetMatches(query string) []entry.Entry {
	s.mu.RLock()
	entries := make([]entry.Entry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}
	s.mu.RUnlock()
	return s.Engine.Rank(entries, query)
}
