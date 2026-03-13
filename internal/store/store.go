package store

import (
	"sync"

	"uberlauncher/internal/engines"
	"uberlauncher/internal/types"
)

type Store struct {
	mu      sync.RWMutex
	entries map[string]types.Entry
	Engine  engines.Engine
}

func New(engine engines.Engine) *Store {
	return &Store{
		entries: make(map[string]types.Entry),
		Engine:  engine,
	}
}

func BuildGlobalEntryId(entry types.Entry) string {
	return entry.SkillName + "|" + entry.EntryID
}

func (s *Store) UpsertEntry(entry types.Entry) {
	id := BuildGlobalEntryId(entry)
	s.mu.Lock()
	s.entries[id] = entry
	s.mu.Unlock()
}

func (s *Store) GetEntry(id string) (types.Entry, bool) {
	s.mu.RLock()
	entry, ok := s.entries[id]
	s.mu.RUnlock()
	return entry, ok
}

func (s *Store) GetMatches(query string) []types.Entry {
	s.mu.RLock()
	entries := make([]types.Entry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}
	s.mu.RUnlock()
	return s.Engine.Rank(entries, query)
}
