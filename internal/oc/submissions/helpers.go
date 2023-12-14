package submissions

import (
	"fmt"
	"ocontest/pkg/structs"
)

func getObjectName(userID, problemID, submissionID int64) string {
	return fmt.Sprintf("%d/%d/%d", problemID, userID, submissionID)
}

func calcScore(results []structs.TestResult, userError string) int {
	if userError != "" {
		return 0
	}

	accepeted := 0
	for _, t := range results {
		if t.Verdict == structs.VerdictOK {
			accepeted += 1
		}
	}
	return 100 * accepeted / len(results)
}
