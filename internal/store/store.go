package store

import (
	"fmt"
	"sync"

	"uberlauncher/internal/engines"
	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type Store struct {
	mu      sync.RWMutex
	entries map[string]entry.Entry
	engine  engines.Engine
	Updates chan struct{}
}

func New(engine engines.Engine) *Store {
	return &Store{
		entries: make(map[string]entry.Entry),
		engine:  engine,
		Updates: make(chan struct{}, 1),
	}
}

func (s *Store) UpsertEntry(e entry.Entry) {
	if e.SkillId == "" {
		panic("entry.SkillId must be set before calling UpsertEntry")
	}
	key := fmt.Sprintf("%s:%s", e.SkillId, e.Id())
	s.mu.Lock()
	s.entries[key] = e
	s.mu.Unlock()
	select {
	case s.Updates <- struct{}{}:
	default:
	}
}

func (s *Store) GetMatches(query string) []engines.Match {
	s.mu.RLock()
	entries := make([]entry.Entry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}
	s.mu.RUnlock()
	return s.engine.Rank(entries, query)
}

type SkillStore struct {
	store   *Store
	skillId string
}

func (s *Store) GetForSkill(sk skill.Skill) *SkillStore {
	return &SkillStore{store: s, skillId: sk.Id()}
}

func (ss *SkillStore) UpsertEntry(e entry.Entry) {
	if e.SkillId != "" && e.SkillId != ss.skillId {
		panic("entry.SkillId does not match skill id!")
	}

	e.SkillId = ss.skillId
	ss.store.UpsertEntry(e)
}
