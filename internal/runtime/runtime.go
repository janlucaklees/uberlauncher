package runtime

import (
	"os"
	"path/filepath"

	"uberlauncher/internal/skill"
	"uberlauncher/internal/store"
	"uberlauncher/internal/types"
)

type Manager struct {
	Store  *store.Store
	Events chan Event
}

func NewManager(store *store.Store) *Manager {
	return &Manager{Store: store, Events: make(chan Event, 32)}
}

type SkillRuntime struct {
	skillName string
	manager   *Manager
}

func (m *Manager) ForSkill(skillName string) *SkillRuntime {
	return &SkillRuntime{skillName: skillName, manager: m}
}

var _ skill.Runtime = (*SkillRuntime)(nil)

func (r *SkillRuntime) PublishEntries(entries []types.Entry) {
	if err := r.manager.Store.Publish(entries); err != nil {
		r.manager.Events <- Event{Type: EventError, Err: err}
		return
	}
	r.manager.Events <- Event{Type: EventEntries}
}

func (r *SkillRuntime) UpsertEntries(entries []types.Entry) {
	if err := r.manager.Store.Upsert(entries); err != nil {
		r.manager.Events <- Event{Type: EventError, Err: err}
		return
	}
	r.manager.Events <- Event{Type: EventEntries}
}

func (r *SkillRuntime) RemoveEntries(ids []string) {
	r.manager.Store.Remove(r.skillName, ids)
	r.manager.Events <- Event{Type: EventEntries}
}

func (r *SkillRuntime) Notify(message string) {
	r.manager.Events <- Event{Type: EventNotify, Message: message}
}

func (r *SkillRuntime) ReportError(err error) {
	r.manager.Events <- Event{Type: EventError, Err: err}
}

func (r *SkillRuntime) CacheDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		base = filepath.Join(os.TempDir(), "uberlauncher")
	}
	path := filepath.Join(base, "uberlauncher", r.skillName)
	_ = os.MkdirAll(path, 0o755)
	return path
}
