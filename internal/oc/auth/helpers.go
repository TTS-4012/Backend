package auth

import (
	"fmt"
	"ocontest/pkg/structs"
	"time"
)

const verificationMessageTemplate = `
Hello %v! Welcome to the ocontest
please click the link below to verify your email 
http://%v/v1/auth/verify/%v> 
ignore this email if you haven't tried to register to ocontest
`

func (a *AuthHandlerImp) genValidateEmailMessage(user structs.User) (link string, err error) {
	token, err := a.jwtHandler.GenToken(user.ID, "verify", time.Duration(a.configs.VerificationDuration))
	if err != nil {
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
