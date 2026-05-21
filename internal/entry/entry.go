package entry

type UI struct {
	KeepOpen func()
}

type Context struct {
	Input string
	UI    UI
}

type Entry struct {
	Key        string
	Label      string
	IsFreeText bool
	Run        func(ctx Context)
}

func (e Entry) Id() string {
	if e.Key != "" {
		return e.Key
	}
	return e.Label
}
