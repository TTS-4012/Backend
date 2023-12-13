// structs for api request and response
package structs

// AUTH
type RegisterUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterUserResponse struct {
	Ok      bool   `json:"ok"`
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

type RequestVerifyEmail struct {
	UserID int64  `json:"user_id"`
	OTP    string `json:"otp"`
}

type AuthenticateResponse struct {
	Ok           bool   `json:"ok"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type RequestLogin struct {
	GrantType string `json:"grant_type"` // base on grant type, we need either password or otp
	Email     string `json:"email"`
	Password  string `json:"password"`
	OTP       string `json:"otp"`
}

type RequestGetOTPLogin struct {
	Email string `json:"email"`
}

type RequestEditUser struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// PROBLEMS
type RequestCreateProblem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ResponseCreateProblem struct {
	ProblemID int64 `json:"problem_Id"`
}

type RequestListProblems struct {
	OrderedBy  string `json:"ordered_by"`
	Descending bool   `json:"descending"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
	GetCount   bool   `json:"get_count"`
}

type ResponseListProblems struct {
	TotalCount int                        `json:"total_count,omitempty"`
	Problems   []ResponseListProblemsItem `json:"problems"`
}

type ResponseListProblemsItem struct {
	ProblemID  int64  `json:"problem_id"`
	Title      string `json:"title"`
	SolveCount int64  `json:"solve_count"`
	Hardness   int64  `json:"hardness"`
}

type ResponseGetProblem struct {
	ProblemID   int64  `json:"problem_Id"`
	Title       string `json:"title"`
	SolveCount  int64  `json:"solve_count"`
	Hardness    int64  `json:"hardness"`
	Description string `json:"description"`
}

type RequestUpdateProblem struct {
	Id          int64
	Title       string `json:"title"`
	Description string `json:"description"`
}

// SUBMISSIONS
type RequestSubmit struct {
	UserID      int64
	ProblemID   int64
	Code        []byte
	FileName    string
	ContentType string
	Language    string
}

type ResponseGetSubmission struct {
	Metadata SubmissionMetadata
	RawCode  []byte `json:"data"`
}

type ResponseGetSubmissionResults struct {
	TestStates []TestState `json:"test_states"`
	Score      int         `json:"score"`
	Message    string      `json:"message"`
}

type RequestListSubmissions struct {
	ProblemID  int64 `json:"problem_id"`
	UserID     int64 `json:"user_id"`
	Descending bool  `json:"descending"`
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
	GetCount   bool  `json:"get_count"`
}

type ResponseListSubmissions struct {
	TotalCount  int                           `json:"total_count,omitempty"`
	Submissions []ResponseListSubmissionsItem `json:"submissions"`
}

type SubmissionListMetadata struct {
	ID        int64  `json:"submission_id"`
	Language  string `json:"language"`
	CreatedAt string `json:"created_at"`
	FileName  string `json:"file_name"`
}

type ResponseListSubmissionsItem struct {
	Metadata SubmissionListMetadata       `json:"metadata"`
	Results  ResponseGetSubmissionResults `json:"results"`
}
