package notifier

type Severity int

const (
	Debug Severity = iota
	Info
	Warning
	Error
)

type Event struct {
	Severity Severity
	Text     string
}

func NewDebug(text string) Event {
	return Event{Severity: Debug, Text: text}
}

func NewInfo(text string) Event {
	return Event{Severity: Info, Text: text}
}

func NewWarning(text string) Event {
	return Event{Severity: Warning, Text: text}
}

func NewError(err error) Event {
	return Event{Severity: Error, Text: err.Error()}
}
