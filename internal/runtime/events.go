package runtime

import "uberlauncher/internal/types"

type EventType int

const (
	EventEntries EventType = iota
	EventNotify
	EventError
)

type Event struct {
	Type    EventType
	Entries []types.EntryDTO
	Message string
	Err     error
}
