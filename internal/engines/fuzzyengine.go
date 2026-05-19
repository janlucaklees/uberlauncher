package engines

import (
	"github.com/sahilm/fuzzy"

	"uberlauncher/internal/entry"
)

type FuzzyEngine struct{}

func New() *FuzzyEngine {
	return &FuzzyEngine{}
}

func (fe *FuzzyEngine) Rank(entries []entry.Entry, query string) []entry.Entry {
	if query == "" {
		return entries
	}

	sources := make([]string, len(entries))
	for i, e := range entries {
		sources[i] = e.Label
	}

	matches := fuzzy.Find(query, sources)
	result := make([]entry.Entry, len(matches))
	for i, m := range matches {
		result[i] = entries[m.Index]
	}
	return result
}
