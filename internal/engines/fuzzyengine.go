package engines

import (
	"github.com/sahilm/fuzzy"

	"uberlauncher/internal/types"
)

type FuzzyEngine struct{}

func New() *FuzzyEngine {
	return &FuzzyEngine{}
}

func (e *FuzzyEngine) Rank(entries []types.Entry, query string) []types.Entry {
	if query == "" {
		return entries
	}

	sources := make([]string, len(entries))
	for i, entry := range entries {
		sources[i] = entry.DisplayText
	}

	matches := fuzzy.Find(query, sources)
	result := make([]types.Entry, len(matches))
	for i, m := range matches {
		result[i] = entries[m.Index]
	}
	return result
}
