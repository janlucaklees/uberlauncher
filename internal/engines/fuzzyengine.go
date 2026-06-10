package engines

import (
	"sort"

	"github.com/sahilm/fuzzy"

	"uberlauncher/internal/entry"
)

type FuzzyEngine struct{}

func New() *FuzzyEngine {
	return &FuzzyEngine{}
}

func (fe *FuzzyEngine) Rank(entries []entry.Entry, query string) []Match {

	// Prepare data for processing and create lookup tables
	var terms []string
	var termIndexToTermLookup []entry.Term
	var termIndexToEntryLookup []entry.Entry
	for _, e := range entries {
		for _, t := range e.GetTerms() {
			terms = append(terms, t.Text)
			termIndexToTermLookup = append(termIndexToTermLookup, t)
			termIndexToEntryLookup = append(termIndexToEntryLookup, e)
		}
	}

	var matches []fuzzy.Match
	var result []Match

	// If we do not have query string, we need to build the matches on our own as fuzzy just return nil.
	if query != "" {
		matches = fuzzy.Find(query, terms)
	} else {
		matches = getAllMatches(terms)
	}

	addedEntries := map[string]struct{}{}
	for _, m := range matches {

		// Skip matches that didn't reach their specified min-score.
		t := termIndexToTermLookup[m.Index]
		if t.MinScore > m.Score {
			continue
		}

		// Skip entries we already have in the result array
		e := termIndexToEntryLookup[m.Index]
		if _, ok := addedEntries[e.GlobalId()]; ok {
			continue
		}

		result = append(result, Match{
			Entry: e,
			Score: m.Score,
		})
		addedEntries[e.GlobalId()] = struct{}{}
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Score != result[j].Score {
			return result[i].Score > result[j].Score
		}
		return result[i].Entry.Label < result[j].Entry.Label
	})

	return result
}

func getAllMatches(terms []string) []fuzzy.Match {
	var matches []fuzzy.Match

	for i, t := range terms {
		matches = append(matches, fuzzy.Match{
			Str:   t,
			Index: i,
			Score: 0,
		})
	}

	return matches
}
