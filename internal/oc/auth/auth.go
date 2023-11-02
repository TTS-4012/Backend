package auth

import (
	"context"
	"ocontest/internal/db"
	"ocontest/internal/jwt"
	"ocontest/pkg"
	"ocontest/pkg/configs"
	"ocontest/pkg/smtp"
	"ocontest/pkg/structs"
)

type AuthHandler interface {
	RegisterUser(ctx context.Context, request structs.RegisterUserRequest) (structs.RegisterUserResponse, int, error)
}

type AuthHandlerImp struct {
	authRepo   db.AuthRepo
	jwtHandler jwt.TokenGenerator
	smtpSender smtp.Sender
	configs    configs.OContestConf
}

func NewAuthHandler(authRepo db.AuthRepo, jwtHandler jwt.TokenGenerator, smtpSender smtp.Sender, config configs.OContestConf) AuthHandler {
	return &AuthHandlerImp{
		authRepo:   authRepo,
		jwtHandler: jwtHandler,
		smtpSender: smtpSender,
		configs:    config,
	}
}

func (p *AuthHandlerImp) RegisterUser(ctx context.Context, reqData structs.RegisterUserRequest) (ans structs.RegisterUserResponse, status int, err error) {
	logger := pkg.Log.WithField("method", "RegisterUser")

	user := structs.User{
		Username:       reqData.Username,
		HashedPassword: reqData.Password,
	}

	validateLink := p.genValidateLink(user)
	p.smtpSender.SendEmail(reqData.Email, "Welcome to OContest")
	ans = structs.RegisterUserResponse{
		Ok:      true,
		Message: "Success",
	}
	return
}
