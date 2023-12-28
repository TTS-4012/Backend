package contests

import (
	"github.com/ocontest/backend/pkg/structs"
	"sort"
)

func calcScore(results []structs.TestResult) int {
	correct := 0
	total := len(results)
	for _, r := range results {
		if r.Verdict == structs.VerdictOK {
			correct += 1
		}
	}
	if len(results) == 0 {
		return 0
	}
	return 100 * correct / total
}

func assoaciateSort(scores []int, standings []structs.ScoreboardUserStanding) []structs.ScoreboardUserStanding {
	compacted := SortByOther{mainSlice: scores, otherSlice: standings}

	sort.Sort(compacted)

	return compacted.otherSlice
}
