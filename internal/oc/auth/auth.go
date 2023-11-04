package auth

import (
	"context"
	"github.com/sirupsen/logrus"
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
	RegisterUser(ctx context.Context, request structs.RegisterUserRequest) (structs.RegisterUserResponse, int)
	VerifyEmail(ctx context.Context, token string) int
	LoginUser(ctx context.Context, request structs.LoginUserRequest) (structs.LoginUserResponse, int)
	RenewToken(ctx context.Context, oldRefreshToken string) (structs.RenewTokenResponse, int)
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

func (p *AuthHandlerImp) RegisterUser(ctx context.Context, reqData structs.RegisterUserRequest) (ans structs.RegisterUserResponse, status int) {
	logger := pkg.Log.WithField("method", "RegisterUser")

	encryptedPassword, err := p.aesHandler.Encrypt(reqData.Password)
	if err != nil {
		logger.Error("error on encrypting password", err)
		status = 503
		ans.Message = "something went wrong, please try again later."
		return
	}

	var user structs.User
	user, err = p.authRepo.GetByUsername(ctx, reqData.Username)
	if err != nil {

		user := structs.User{
			Username:          reqData.Username,
			EncryptedPassword: encryptedPassword,
			Email:             reqData.Email,
			Verified:          false,
		}
		userID, newErr := p.authRepo.InsertUser(ctx, user)
		if newErr != nil {
			logger.Errorf("couldn't insert user in database, error on get: %v, error on insert: %v", err, newErr)
			status = 503
			ans.Message = "something went wrong, please try again later."
			return
		}
		user.ID = userID
	}

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
func (p *AuthHandlerImp) LoginUser(ctx context.Context, request structs.LoginUserRequest) (structs.LoginUserResponse, int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "LoginUser",
		"module": "auth",
	})

	userInDB, err := p.authRepo.GetByUsername(ctx, request.Username)
	if err != nil {
		logger.Error("error on getting user from db", err)
		return structs.LoginUserResponse{
			Ok:      false,
			Message: "couldn't find user",
		}, http.StatusInternalServerError
	}
	if !userInDB.Verified {
		logger.Warning("unverify user login attempt", userInDB.Username)
		return structs.LoginUserResponse{
			Ok:      false,
			Message: "user is not verified",
		}, http.StatusForbidden
	}
	encPassword, err := p.aesHandler.Encrypt(request.Password)
	if err != nil {
		logger.Error("error on encrypting password")
		return structs.LoginUserResponse{
			Ok:      false,
			Message: "something went wrong",
		}, http.StatusInternalServerError
	}
	if encPassword != userInDB.EncryptedPassword {
		logger.Warning("wrong password")
		return structs.LoginUserResponse{
			Ok:      false,
			Message: "wrong password",
		}, http.StatusForbidden
	}
	accessToken, refreshToken, err := p.genAuthToken(userInDB.ID)
	if err != nil {
		logger.Error("error on creating tokens", err)
		return structs.LoginUserResponse{
			Ok:      false,
			Message: "something went wrong",
		}, http.StatusInternalServerError
	}
	return structs.LoginUserResponse{
		Ok:           true,
		Message:      "success",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, http.StatusOK
}

func (p *AuthHandlerImp) RenewToken(ctx context.Context, oldRefreshToken string) (structs.RenewTokenResponse, int) {
	uid, typ, err := p.jwtHandler.ParseToken(oldRefreshToken)
	if err != nil || typ != "refresh" {
		return structs.RenewTokenResponse{
			Ok:      false,
			Message: "current token is invalid",
		}, http.StatusBadRequest
	}
	accessToken, refreshToken, err := p.genAuthToken(uid)
	if err != nil {
		return structs.RenewTokenResponse{
			Ok:      false,
			Message: "couldn't generate new token",
		}, http.StatusInternalServerError
	}
	return structs.RenewTokenResponse{
		Ok:           true,
		Message:      "success",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, http.StatusOK
}
