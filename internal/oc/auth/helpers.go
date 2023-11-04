package auth

import (
	"fmt"
	"ocontest/pkg"
	"ocontest/pkg/structs"
)

const verificationMessageTemplate = `
Hello %v! Welcome to the ocontest
please click the link below to verify your email 
http://%v/v1/auth/verify/%v> 
ignore this email if you haven't tried to register to ocontest
`

func (a *AuthHandlerImp) genValidateEmailMessage(user structs.User) (link string, err error) {
	token, err := a.jwtHandler.GenToken(user.ID, "verify", a.configs.Auth.Duration.VerifyEmail)
	if err != nil {
		pkg.Log.Error("error on generating email message, ", err)
		return
	}
	token, err = a.aesHandler.Encrypt(token)
	if err != nil {
		return
	}
	return fmt.Sprintf(verificationMessageTemplate,
		user.Username,
		a.configs.Server.Host+":"+a.configs.Server.Port,
		token), nil

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
