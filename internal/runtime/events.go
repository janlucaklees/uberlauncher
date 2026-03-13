package runtime

import "uberlauncher/internal/types"

type EventType int

const (
	EventEntries EventType = iota
	EventNotify
	EventError
)

// Todo use general Payload field and implement different kinds of payload
type Event struct {
	Type    EventType
	Entries []types.Entry
	Message string
	Err     error
}
