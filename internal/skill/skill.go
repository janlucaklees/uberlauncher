package skill

import (
	"context"

	"uberlauncher/internal/types"
)

type Skill interface {
	Name() string
	Init(ctx context.Context, runtime Runtime) error
	Execute(ctx context.Context, cmd types.Command) error
}

type Runtime interface {
	PublishEntries(entries []types.Entry)
	UpsertEntries(entries []types.Entry)
	RemoveEntries(ids []string)
	Notify(message string)
	ReportError(err error)
	CacheDir() string
}
