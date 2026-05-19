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
	ReportError(err error)
	ReportMessage(message string)
	SendNotification(message string)
}

type Config interface {
	Get(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	Set(key string, value any)
}

type Store interface {
	UpsertEntry(e entry.Entry)
}

type Context struct {
	Runtime  Runtime
	Notifier Notifier
	Store    Store

	Config Config
	Cache  Cache
}

type Skill interface {
	Id() string
	Init(ctx Context)
}
