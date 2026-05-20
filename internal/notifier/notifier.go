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
	n.Events <- NewError(err)
}

func (n *Notifier) ReportWarning(message string) {
	n.Events <- NewWarning(message)
}

func (n *Notifier) ReportMessage(message string) {
	n.Events <- NewInfo(message)
}

func (n *Notifier) SendNotification(message string) {
	_ = send("uberlauncher", message)
}
