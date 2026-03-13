package skills

import (
	"uberlauncher/internal/cache"
	"uberlauncher/internal/types"
)

type Skill interface {
	Name() string
	Init(runtime Runtime) error
	Execute(cmd types.Command) error
}

type Runtime interface {
	ReportError(err error)
	UpsertEntries(entries []types.Entry)
	UpsertEntry(entry types.Entry)
	Notify(message string)
	Cache() *cache.SkillCache
}
