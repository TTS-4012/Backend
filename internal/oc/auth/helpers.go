package auth

import (
	"fmt"
	"ocontest/pkg/structs"
)

const registerMessageTemplate = `
Hello %v! Welcome to the ocontest
please enter the code below to verify your email address
Code: %v
ignore this email if you haven't tried to register to ocontest
`

const loginMessageTemplate = `
Hello %v! Welcome to the ocontest
please enter the code below to login to your account
Code: %v
`

type Operation int

const (
	Register Operation = iota
	Login
)

func (a *AuthHandlerImp) genEmailMessage(user structs.User, otpCode string, operation Operation) string {
	var template string
	switch operation {
	case Register:
		template = registerMessageTemplate
	case Login:
		template = loginMessageTemplate
	default:
		return ""
	}
	return fmt.Sprintf(template,
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
