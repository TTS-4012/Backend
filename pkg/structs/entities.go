package structs

type User struct {
	ID                int64
	Username          string
	EncryptedPassword string
	Email             string
	Verified          bool
}

type ProblemDescription struct {
	ID          string
	Description string
}

type Problem struct {
	CreatedBy   int64
	ID          int64
	Title       string
	DocumentID  string
	SolvedCount int64
	Hardness    int64
	IsPrivate   bool
}

type SubmissionMetadata struct {
	ID            int64  `json:"id"`
	ProblemID     int64  `json:"problem_id"`
	UserID        int64  `json:"user_id"`
	FileName      string `json:"file_name"`
	JudgeResultID string `json:"judge_result_id"`
	Status        string `json:"status"`   // either 'new', 'processing', 'processed'
	Language      string `json:"language"` // just 'python' for now
	Public        bool   `json:"public"`
	CreatedAT     string `json:"created_at"`
}

type Testcase struct {
	ProblemID int64 `json:"problem_id"`
	ID        int64 `json:"id"`

	Input          string `json:"input,omitempty"`
	ExpectedOutput string `json:"expected_output,omitempty"`
}

type TestResult struct {
	SubmissionID int64 `json:"submission_id"`
	TestcaseID   int64 `json:"id"`

	RunnerOutput string `json:"runner_output"`
	RunnerError  string `json:"runner_error"`
	Verdict
}

type JudgeRequest struct {
	SubmissionID int64      `json:"submission_id"`
	Code         string     `json:"code"`
	Testcases    []Testcase `json:"testcases"`
}

type JudgeResponse struct {
	ServerError string       `json:"server_error" bson:"server_error"` // for example, a database failure
	TestResults []TestResult `json:"test_results" bson:"test_results"` // 'Wrong', 'Success', 'Timelimit', 'Memorylimit'
}

type ContestProblem struct {
	ID    int64
	Title string
}

type Contest struct {
	CreatedBy int64
	ID        int64
	Title     string
	StartTime int64
	Duration  int
	Problems  []ContestProblem
}
