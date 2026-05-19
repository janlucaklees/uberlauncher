package engines

import "uberlauncher/internal/entry"

type Engine interface {
	Rank(entries []entry.Entry, query string) []entry.Entry
}
