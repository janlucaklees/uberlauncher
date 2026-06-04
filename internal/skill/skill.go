package skill

import (
	"os/exec"
	"time"
	"uberlauncher/internal/entry"
)

type Runtime interface {
	HasCommand(name string) bool
	Go(fn func())
	Done() <-chan struct{}
	Command(name string, args ...string) *exec.Cmd
}

type Cache interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte) error
}

type Notifier interface {
	Debug(message string)
	ReportError(err error)
	ReportWarning(message string)
	ReportMessage(message string)
	SendNotification(message string)
}

type ConfigMap map[string]any

func (m ConfigMap) IsEnabled() bool {
	enabled, ok := m["enabled"].(bool)
	return !ok || enabled
}

type Store interface {
	UpsertEntry(e entry.Entry)
}

type StatusUpdater interface {
	Register(key string, interval time.Duration, fn func() string)
}

type Context struct {
	Runtime  Runtime
	Notifier Notifier
	Store    Store
	Status   StatusUpdater

	Config ConfigMap
	Cache  Cache
}

type Skill interface {
	Id() string
	Init(ctx Context)
}
