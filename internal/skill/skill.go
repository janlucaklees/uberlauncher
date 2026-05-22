package skill

import (
	"uberlauncher/internal/entry"
)

type Runtime interface {
	HasCommand(name string) bool
	Go(fn func())
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

type Store interface {
	UpsertEntry(e entry.Entry)
}

type Context struct {
	Runtime  Runtime
	Notifier Notifier
	Store    Store

	Config ConfigMap
	Cache  Cache
}

type Skill interface {
	Id() string
	Init(ctx Context)
}
