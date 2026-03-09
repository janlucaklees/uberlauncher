package store

import (
	"errors"
	"fmt"
	"sync"

	"uberlauncher/internal/types"
)

type EntryKey struct {
	SkillName string
	EntryID   string
}

type Store struct {
	mu      sync.RWMutex
	bySkill map[string]map[string]types.Entry
}

func New() *Store {
	return &Store{bySkill: make(map[string]map[string]types.Entry)}
}

func (s *Store) Publish(entries []types.Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateDuplicates(entries); err != nil {
		return err
	}

	for _, entry := range entries {
		if _, ok := s.bySkill[entry.SkillName]; !ok {
			s.bySkill[entry.SkillName] = make(map[string]types.Entry)
		}
		s.bySkill[entry.SkillName][entry.EntryID] = entry
	}
	return nil
}

func (s *Store) Upsert(entries []types.Entry) error {
	return s.Publish(entries)
}

func (s *Store) Remove(skillName string, ids []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, ok := s.bySkill[skillName]
	if !ok {
		return
	}
	for _, id := range ids {
		delete(bucket, id)
	}
}

func (s *Store) ListAll() []types.Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]types.Entry, 0)
	for _, bucket := range s.bySkill {
		for _, entry := range bucket {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (s *Store) Exists(key EntryKey) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucket, ok := s.bySkill[key.SkillName]
	if !ok {
		return false
	}
	_, ok = bucket[key.EntryID]
	return ok
}

func (s *Store) Get(key EntryKey) (types.Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucket, ok := s.bySkill[key.SkillName]
	if !ok {
		return types.Entry{}, false
	}
	entry, ok := bucket[key.EntryID]
	return entry, ok
}

func validateDuplicates(entries []types.Entry) error {
	seen := make(map[string]map[string]struct{})
	for _, entry := range entries {
		if entry.SkillName == "" || entry.EntryID == "" {
			return errors.New("entry must have skill name and entry ID")
		}
		if _, ok := seen[entry.SkillName]; !ok {
			seen[entry.SkillName] = make(map[string]struct{})
		}
		if _, ok := seen[entry.SkillName][entry.EntryID]; ok {
			return fmt.Errorf("duplicate entry ID '%s' for skill '%s'", entry.EntryID, entry.SkillName)
		}
		seen[entry.SkillName][entry.EntryID] = struct{}{}
	}
	return nil
}
