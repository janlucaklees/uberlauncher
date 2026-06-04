package engines

import "uberlauncher/internal/entry"

type Match struct {
	Entry entry.Entry
	Score int
}

type Engine interface {
	Rank(entries []entry.Entry, query string) []Match
}
