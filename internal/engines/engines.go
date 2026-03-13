package engines

import "uberlauncher/internal/types"

type Engine interface {
	Rank(entries []types.Entry, query string) []types.Entry
}
