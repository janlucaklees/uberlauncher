package runtime

type EventType int

const (
	EventNewEntry EventType = iota
	EventMessage
	EventError
)

// Todo use general Payload field and implement different kinds of payload
type Event struct {
	Type    EventType
	Message string
	Err     error
}
