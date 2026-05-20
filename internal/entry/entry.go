package entry

type UI struct {
	KeepOpen func()
}

type Context struct {
	Input string
	UI    UI
}

type Entry struct {
	Label      string
	IsFreeText bool
	Run        func(ctx Context)
}
