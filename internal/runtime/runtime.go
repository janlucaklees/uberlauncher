package runtime

import (
	"fmt"
	"os/exec"
	"sync"

	"uberlauncher/internal/cache"
	"uberlauncher/internal/skills"
	"uberlauncher/internal/store"
	"uberlauncher/internal/types"
)

type Runtime struct {
	Cache         *cache.Cache
	Store         *store.Store
	SkillMap      map[string]skills.Skill
	Channel       chan Event
	wg            sync.WaitGroup
	isInitialized bool
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

// Wait blocks until all background tasks started via Go have completed.
func (r *Runtime) Wait() {
	r.wg.Wait()
}

func (r *Runtime) Go(fn func()) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		fn()
	}()
}

func (r *Runtime) Init(skills []skills.Skill) {
	for _, skill := range skills {
		r.RegisterSkill(skill)
	}
	r.isInitialized = true
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

	if r.isInitialized {
		r.Channel <- Event{Type: EventNewEntry}
	}
}

type skillRuntime struct {
	r         *Runtime
	skillName string
}

func (sr *skillRuntime) ReportError(err error) { sr.r.ReportError(err) }
func (sr *skillRuntime) Notify(message string) {
	sr.r.Channel <- Event{Type: EventMessage, Message: message}
}
func (sr *skillRuntime) UpsertEntries(e []types.Entry) { sr.r.UpsertEntries(e) }
func (sr *skillRuntime) UpsertEntry(e types.Entry)     { sr.r.UpsertEntry(e) }
func (sr *skillRuntime) Cache() *cache.SkillCache      { return sr.r.Cache.ForSkill(sr.skillName) }
func (sr *skillRuntime) Go(fn func())                  { sr.r.Go(fn) }

func (r *Runtime) RegisterSkill(skill skills.Skill) {
	_, ok := r.SkillMap[skill.Name()]

	if ok {
		r.ReportError(fmt.Errorf("skill %q already registered", skill.Name()))
		return
	}

	skill.Init(&skillRuntime{r: r, skillName: skill.Name()})

	r.SkillMap[skill.Name()] = skill
}
