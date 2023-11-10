// structs for api request and response
package structs

type RegisterUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterUserResponse struct {
	Ok      bool
	UserID  int64
	Message string
}

type LoginUserRequest struct {
	Username string
	Password string
}

type LoginUserResponse struct {
	Ok           bool   `json:"ok"`
	Message      string `json:"message"`
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type RenewTokenResponse struct {
	Ok           bool
	Message      string
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type RequestVerifyUser struct {
	UserID int64  `json:"user_id"`
	OTP    string `json:"otp"`
}
