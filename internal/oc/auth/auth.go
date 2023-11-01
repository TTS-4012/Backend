package auth

import (
	"context"
	"net/http"
	"ocontest/internal/db"
	"ocontest/internal/jwt"
	"ocontest/pkg"
	"ocontest/pkg/smtp"
)

type AuthHandler interface {
	RegisterUser(ctx context.Context, request pkg.RegisterUserRequest) (pkg.RegisterUserResponse, int, error)
}

type AuthHandlerImp struct {
	authRepo   db.AuthRepo
	jwtHandler jwt.TokenGenerator
	smtpSender smtp.Sender
}

func NewAuthHandler(authRepo db.AuthRepo, jwtHandler jwt.TokenGenerator, smtpSender smtp.Sender) AuthHandler {
	return &AuthHandlerImp{
		authRepo:   authRepo,
		jwtHandler: jwtHandler,
		smtpSender: smtpSender,
	}
}

func (p *AuthHandlerImp) RegisterUser(ctx context.Context, reqData pkg.RegisterUserRequest) (ans pkg.RegisterUserResponse, status int, err error) {
	logger := pkg.Log.WithField("method", "RegisterUser")
	userId, err := p.authRepo.InsertUser(ctx, reqData.Username, reqData.Password, reqData.Email)
	if err != nil {
		logger.Error("error on inserting to database", err)
		status = http.StatusInternalServerError
		return
	}
	accessToken, err := p.jwtHandler.GenAccessToken(userId)
	if err != nil {
		logger.Error("error on creating access token: ", err)
		status = http.StatusInternalServerError
		return
	}
	refreshToken, err := p.jwtHandler.GenRefreshToken(userId)
	if err != nil {
		logger.Error("error on creating refresh token: ", err)
		status = http.StatusInternalServerError
		return
	}

	ans = pkg.RegisterUserResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return
}
