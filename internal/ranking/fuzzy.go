package ranking

import (
	"math"
	"sort"
	"time"

	"github.com/sahilm/fuzzy"
	"uberlauncher/internal/types"
)

type RankedEntry struct {
	Entry types.EntryDTO
	Score float64
}

func RankEntries(query string, entries []types.EntryDTO, usage *UsageStore) []RankedEntry {
	if len(entries) == 0 {
		return nil
	}

	if query == "" {
		ranked := make([]RankedEntry, 0, len(entries))
		for _, entry := range entries {
			score := usageScore(entry, usage)
			ranked = append(ranked, RankedEntry{Entry: entry, Score: score})
		}
		sortRanked(ranked)
		return ranked
	}

	inputs := make([]string, len(entries))
	for i, entry := range entries {
		inputs[i] = entry.DisplayText
	}

	matches := fuzzy.Find(query, inputs)
	ranked := make([]RankedEntry, 0, len(matches))
	for _, match := range matches {
		entry := entries[match.Index]
		score := float64(match.Score) + usageScore(entry, usage)
		ranked = append(ranked, RankedEntry{Entry: entry, Score: score})
	}
	sortRanked(ranked)
	return ranked
}

func usageScore(entry types.EntryDTO, usage *UsageStore) float64 {
	if usage == nil {
		return 0
	}
	key := UsageKey(entry.SkillName, entry.EntryID)
	rec := usage.Get(key)
	return usageWeight(rec.Count) + recencyWeight(rec.LastUsed)
}

func usageWeight(count int) float64 {
	if count <= 0 {
		return 0
	}
	return math.Min(20, math.Log2(float64(count)+1)*5)
}

func recencyWeight(last time.Time) float64 {
	if last.IsZero() {
		return 0
	}
	hours := time.Since(last).Hours()
	weight := 10 - (hours / 24)
	if weight < 0 {
		return 0
	}
	if weight > 10 {
		return 10
	}
	return weight
}

func sortRanked(ranked []RankedEntry) {
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return ranked[i].Entry.DisplayText < ranked[j].Entry.DisplayText
		}
		return ranked[i].Score > ranked[j].Score
	})
}
