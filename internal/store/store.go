package store

import (
	"uberlauncher/internal/engines"
	"uberlauncher/internal/types"
)

type Store struct {
	Entries map[string]types.Entry
	Engine  engines.Engine
}

func New(engine engines.Engine) *Store {
	return &Store{
		Entries: make(map[string]types.Entry),
		Engine:  engine,
	}
}

func BuildGlobalEntryId(entry types.Entry) string {
	return entry.SkillName + "|" + entry.EntryID
}

func (s *Store) UpsertEntry(entry types.Entry) {
	id := BuildGlobalEntryId(entry)
	s.Entries[id] = entry
}

func (s *Store) GetMatches(query string) []types.Entry {
	entries := make([]types.Entry, 0, len(s.Entries))
	for _, e := range s.Entries {
		entries = append(entries, e)
	}
	return s.Engine.Rank(entries, query)
}
