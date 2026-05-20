package entry

type Context struct {
	Input string
}

type Entry struct {
	Label      string
	IsFreeText bool
	Run        func(ctx Context)
}
