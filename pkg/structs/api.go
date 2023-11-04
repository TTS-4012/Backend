// structs for api request and response
package structs

type RegisterUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterUserResponse struct {
	Ok      bool
	Message string
}

type LoginUserRequest struct {
	Username string
	Password string
}

type LoginUserResponse struct {
	Ok           bool
	Message      string
	AccessToken  string
	RefreshToken string
}

type RenewTokenResponse struct {
	Ok           bool
	Message      string
	AccessToken  string
	RefreshToken string
}
