package auth

import (
	"fmt"
	"ocontest/pkg/structs"
)

const verificationMessageTemplate = `
Hello %v! Welcome to the ocontest
please enter the code below to verify your email address
Code: %v
ignore this email if you haven't tried to register to ocontest
`

func (a *AuthHandlerImp) genValidateEmailMessage(user structs.User, otpCode string) string {
	return fmt.Sprintf(verificationMessageTemplate,
		user.Username,
		otpCode,
	)

}

// genAuthToken will just try to generate tokens, it doesn't do any authentication and they should be done before calling this method
func (a *AuthHandlerImp) genAuthToken(userID int64) (string, string, error) {
	accessToken, err := a.jwtHandler.GenToken(userID, "access", a.configs.Auth.Duration.AccessToken)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := a.jwtHandler.GenToken(userID, "refresh", a.configs.Auth.Duration.RefreshToken)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}
