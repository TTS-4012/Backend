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
	Testcases   []string
}

type Problem struct {
	CreatedBy   int64
	ID          int64
	Title       string
	DocumentID  string
	SolvedCount int64
	Hardness    int64
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
}

type Testcase struct {
	Input  string `json:"input"`
	Answer string `json:"answer"`
}

type JudgeRequest struct {
	Code      []byte   `json:"code"`
	Testcases []string `json:"testcases"`
}

type TestState int64

const (
	TestStateSuccess     TestState = iota
	TestStateWrong       TestState = iota
	TestStateTimeLimit   TestState = iota
	TestStateMemoryLimit TestState = iota
)

type JudgeResponse struct {
	ServerError error       `json:"server_error"` // for example, a database failure
	UserError   error       `json:"user_error"`   // for example, a compile error on user code
	TestStates  []TestState `json:"testStates"`   // 'Wrong', 'Success', 'Timelimit', 'Memorylimit'
}
