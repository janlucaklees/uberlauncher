package skills

import (
	"uberlauncher/internal/cache"
	"uberlauncher/internal/types"
)

type Skill interface {
	Name() string
	Init(runtime Runtime)
	Execute(cmd types.Command)
}

type Runtime interface {
	ReportError(err error)
	UpsertEntries(entries []types.Entry)
	UpsertEntry(entry types.Entry)
	Notify(message string)
	Cache() *cache.SkillCache
	Go(fn func())
}
