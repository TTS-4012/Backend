// structs for api request and response
package structs

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

type LoginUserRequest struct {
	Username string
	Password string
}

type AuthenticateResponse struct {
	Ok           bool   `json:"ok"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type RequestWithOTPCreds struct {
	UserID int64  `json:"user_id"`
	OTP    string `json:"otp"`
}

type RequestEditUser struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

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
	ProblemID  int    `json:"problem_Id"`
	Title      string `json:"title"`
	SolveCount int    `json:"solve_count"`
	Hardness   int    `json:"hardness"`
}

type ResponseGetProblem struct {
	ProblemID   int    `json:"problem_Id"`
	Title       string `json:"title"`
	SolveCount  int    `json:"solve_count"`
	Hardness    int    `json:"hardness"`
	Description string `json:"description"`
}
