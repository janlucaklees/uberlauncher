package ranking

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type UsageRecord struct {
	Count    int       `json:"count"`
	LastUsed time.Time `json:"last_used"`
}

type UsageStore struct {
	mu    sync.Mutex
	path  string
	items map[string]UsageRecord
}

func NewUsageStore(cacheDir string) *UsageStore {
	return &UsageStore{path: filepath.Join(cacheDir, "usage.json"), items: make(map[string]UsageRecord)}
}

func (s *UsageStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	var items map[string]UsageRecord
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	s.items = items
	return nil
}

func (s *UsageStore) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *UsageStore) Get(key string) UsageRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.items[key]
}

func (s *UsageStore) Bump(key string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec := s.items[key]
	rec.Count++
	rec.LastUsed = now
	s.items[key] = rec
}

func UsageKey(skillName, entryID string) string {
	return skillName + "|" + entryID
}
