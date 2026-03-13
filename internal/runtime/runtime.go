package runtime

import (
	"fmt"

	"uberlauncher/internal/cache"
	"uberlauncher/internal/skills"
	"uberlauncher/internal/store"
	"uberlauncher/internal/types"
)

type Runtime struct {
	Cache    *cache.Cache
	Store    *store.Store
	SkillMap map[string]skills.Skill
	Channel  chan Event
}

func New(cache *cache.Cache, store *store.Store) *Runtime {
	r := Runtime{
		Cache:    cache,
		Store:    store,
		SkillMap: make(map[string]skills.Skill),
		Channel:  make(chan Event, 32),
	}

	return &r
}

func (r *Runtime) ReportError(err error) {
	r.Channel <- Event{Type: EventError, Err: err}
}

func (r *Runtime) UpsertEntries(entries []types.Entry) {
	for _, entry := range entries {
		r.UpsertEntry(entry)
	}
}

func (r *Runtime) UpsertEntry(entry types.Entry) {
	r.Store.UpsertEntry(entry)
	r.Channel <- Event{Type: EventEntries}
}

type skillRuntime struct {
	r         *Runtime
	skillName string
}

func (sr *skillRuntime) ReportError(err error) { sr.r.ReportError(err) }
func (sr *skillRuntime) Notify(message string) {
	sr.r.Channel <- Event{Type: EventNotify, Message: message}
}
func (sr *skillRuntime) UpsertEntries(e []types.Entry) { sr.r.UpsertEntries(e) }
func (sr *skillRuntime) UpsertEntry(e types.Entry)     { sr.r.UpsertEntry(e) }
func (sr *skillRuntime) Cache() *cache.SkillCache      { return sr.r.Cache.ForSkill(sr.skillName) }

func (r *Runtime) RegisterSkill(skill skills.Skill) {
	_, ok := r.SkillMap[skill.Name()]

	if ok {
		r.ReportError(fmt.Errorf("skill %q already registered", skill.Name()))
		return
	}

	err := skill.Init(&skillRuntime{r: r, skillName: skill.Name()})

	if err != nil {
		r.ReportError(fmt.Errorf("skill %q init failed: %w", skill.Name(), err))
		return
	}

	r.SkillMap[skill.Name()] = skill
}
