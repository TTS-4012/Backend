package auth

import (
	"context"
	"ocontest/internal/db"
	"ocontest/internal/jwt"
	"ocontest/pkg"
	"ocontest/pkg/aes"
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
	configs    *configs.OContestConf
	aesHandler aes.AESHandler
}

func NewAuthHandler(authRepo db.AuthRepo, jwtHandler jwt.TokenGenerator, smtpSender smtp.Sender, config *configs.OContestConf, aesHandler aes.AESHandler) AuthHandler {
	return &AuthHandlerImp{
		authRepo:   authRepo,
		jwtHandler: jwtHandler,
		smtpSender: smtpSender,
		configs:    config,
		aesHandler: aesHandler,
	}
}

func (p *AuthHandlerImp) RegisterUser(ctx context.Context, reqData structs.RegisterUserRequest) (ans structs.RegisterUserResponse, status int, err error) {
	logger := pkg.Log.WithField("method", "RegisterUser")

	encryptedPassword, err := p.aesHandler.Encrypt(reqData.Password)
	if err != nil {
		logger.Error("error on encrypting password", err)
		status = 503
		err = pkg.ErrBadRequest
		return
	}

	user := structs.User{
		Username:          reqData.Username,
		EncryptedPassword: encryptedPassword,
		Email:             reqData.Email,
		Verified:          false,
	}
	userID, err := p.authRepo.InsertUser(ctx, user)
	user.ID = userID

	validateEmailMessage, err := p.genValidateEmailMessage(user)
	if err != nil {
		logger.Error("error on creating verify email message", err)
		status = 503
		err = pkg.ErrInternalServerError
		return
	}
	err = p.smtpSender.SendEmail(reqData.Email, "Welcome to OContest", validateEmailMessage)
	if err != nil {
		logger.Error("error on sending email", err)
		status = 503
		err = pkg.ErrInternalServerError
		return
	}

	ans = structs.RegisterUserResponse{
		Ok:      true,
		Message: "Sent Verification email",
	}
	return
}
