package notifier

type EventType int

const (
	EventMessage EventType = iota
	EventError
)

type Event struct {
	Type    EventType
	Message string
	Err     error
}
