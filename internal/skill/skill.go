package skill

import (
	"context"

	"uberlauncher/internal/types"
)

type Manifest struct {
	Name             string
	SupportsFreeText bool
}

type Skill interface {
	Manifest() Manifest
	Start(ctx context.Context, runtime Runtime) error
	Execute(ctx context.Context, cmd types.RunCommandDTO) error
	Stop(ctx context.Context) error
}

type Runtime interface {
	PublishEntries(entries []types.EntryDTO)
	UpsertEntries(entries []types.EntryDTO)
	RemoveEntries(ids []string)
	Notify(message string)
	ReportError(err error)
	CacheDir() string
}
