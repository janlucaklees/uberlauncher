package notifier

type Notifier struct {
	Events chan Event
}

func New() *Notifier {
	return &Notifier{
		Events: make(chan Event, 64),
	}
}

func (n *Notifier) ReportError(err error) {
	n.Events <- Event{Type: EventError, Err: err}
}

func (n *Notifier) ReportMessage(message string) {
	n.Events <- Event{Type: EventMessage, Message: message}
}

func (n *Notifier) SendNotification(message string) {
	_ = send("uberlauncher", message)
}
