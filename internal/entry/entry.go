package entry

type Context struct {
	Input string
}

type Entry struct {
	Label string
	Run   func(ctx Context)
}
