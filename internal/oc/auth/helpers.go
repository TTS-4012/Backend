package auth

import (
	"fmt"
	"ocontest/pkg/structs"
	"time"
)

const verificationMessageTemplate = `
Hello %v to ocontest
please click the link below to verify your email 
%v/v1/auth/verify/%v
ignore this email if you haven't tried to register to ocontest
`

func (a *AuthHandlerImp) genValidateEmailMessage(user structs.User) (link string, err error) {
	fields := make(map[string]interface{})
	fields["id"] = user.ID
	fields["exp"] = time.Now().Add(a.configs.VerificationDuration).Unix()
	token, err := a.jwtHandler.GenToken(user.ID, "verify", a.configs.VerificationDuration)
	if err != nil {
		return
	}
	return fmt.Sprintf(verificationMessageTemplate,
		user.Username,
		a.configs.Server.Host+":"+a.configs.Server.Port,
		token), nil

}
