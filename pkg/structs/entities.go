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
}

type Testcase struct {
	Input  string `json:"input"`
	Answer string `json:"answer"`
}

type JudgeSubmissions struct {
	ID         string     `json:"id"`
	Code       []byte     `json:"code"`
	Testcases  []Testcase `json:"testcases"`
	TestStates []string   `json:"testStates"`
}
