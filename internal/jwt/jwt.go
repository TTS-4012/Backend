package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"ocontest/pkg/configs"
	"time"
)

type TokenGenerator interface {
	GenToken(ID int64, jid string, expireTime time.Duration) (string, error)
}

type TokenGeneratorImp struct {
	secret     []byte
	accessExp  time.Duration
	refreshExp time.Duration
}

func NewGenerator(conf configs.SectionJWT) TokenGenerator {
	return TokenGeneratorImp{
		secret:     []byte(conf.Secret),
		accessExp:  conf.AccessDuration,
		refreshExp: conf.RefreshDuration,
	}
}

func (t TokenGeneratorImp) GenToken(userID int64, typ string, expireTime time.Duration) (string, error) {
	mapClaims := jwt.MapClaims{}

	mapClaims["iat"] = time.Now().Unix()
	mapClaims["userID"] = userID
	mapClaims["typ"] = typ
	mapClaims["exp"] = time.Now().Add(expireTime).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	return token.SignedString(t.secret)

}
