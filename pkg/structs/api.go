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
	GrantType string `json:"grant_type"`
	UserID    int64  `json:"user_id"`
	UserName  string `json:"username"`
	Password  string `json:"password"`
	OTP       string `json:"otp"`
}

type RequestGetOTPLogin struct {
	UserID int64 `json:"user_id"`
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

// FILE
type ResponseUploadFile struct {
	FileName string `json:"file_name"`
}

type ResponseDownloadFile struct {
	ContentDisposition string `header:"Content-Disposition"`
	ContentType        string `json:"ContentType"`
	Data               []byte `json:"data"`
}
