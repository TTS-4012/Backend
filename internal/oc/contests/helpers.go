package contests

import (
	"github.com/ocontest/backend/pkg/structs"
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
