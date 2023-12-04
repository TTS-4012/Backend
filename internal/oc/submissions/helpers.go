package submissions

import "fmt"

func getObjectName(userID, problemID, submissionID int64) string {
	return fmt.Sprintf("%d/%d/%d", problemID, userID, submissionID)
}
