package auth

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
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
	VerifyEmail(ctx context.Context, token string) int
	LoginUser(ctx context.Context, request structs.LoginUserRequest) (structs.LoginUserResponse, int, error)
	RenewToken(ctx context.Context, oldRefreshToken string) (structs.RenewTokenResponse, int, error)
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
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			logger.Warn("dup request", err)
			status = 400
			err = pkg.ErrBadRequest
			return
		}
		logger.Error("error on inserting user", err)
		status = 503
		err = pkg.ErrInternalServerError
		return
	}
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

func (p *AuthHandlerImp) VerifyEmail(ctx context.Context, token string) int {
	token, err := p.aesHandler.Decrypt(token)
	if err != nil {
		pkg.Log.Error("error on decrypting token", err)
		return http.StatusBadRequest
	}
	userID, typ, err := p.jwtHandler.ParseToken(token)
	if typ != "verify" {
		err = pkg.ErrBadRequest
	}
	if err != nil {
		pkg.Log.Error("error on parsing jwt", err)
		return http.StatusBadRequest
	}

	err = p.authRepo.VerifyUser(ctx, userID)
	if err != nil {
		pkg.Log.Error("error on verifying user", err)
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
func (p *AuthHandlerImp) LoginUser(ctx context.Context, request structs.LoginUserRequest) (structs.LoginUserResponse, int, error) {
	//TODO implement me
	panic("implement me")
}

func (p *AuthHandlerImp) RenewToken(ctx context.Context, oldRefreshToken string) (structs.RenewTokenResponse, int, error) {
	//TODO implement me
	panic("implement me")
}
