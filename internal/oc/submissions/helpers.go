package submissions

import (
	"fmt"
	"ocontest/pkg/structs"
)

func getObjectName(userID, problemID, submissionID int64) string {
	return fmt.Sprintf("%d/%d/%d", problemID, userID, submissionID)
}

func calcScore(results []structs.TestState, userError string) int {
	if userError != "" {
		return 0
	}
	if len(results) == 0 {
		return 0
	}

	accepeted := 0
	for _, t := range results {
		if t == structs.TestStateSuccess {
			accepeted += 1
		}
	}
	return 100 * accepeted / len(results)
}
