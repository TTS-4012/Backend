package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"ocontest/pkg"
	"ocontest/pkg/configs"
	"time"
)

type TokenGenerator interface {
	GenAccessToken(userID int) (string, error)
	GenRefreshToken(userID int) (string, error)
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

func (t TokenGeneratorImp) genToken(fields map[string]interface{}) (string, error) {
	mapClaims := jwt.MapClaims{}
	for k, v := range fields {
		mapClaims[k] = v
	}
	mapClaims["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	pkg.Log.Debug(t.secret)
	return token.SignedString(t.secret)

}

func (t TokenGeneratorImp) GenAccessToken(userID int) (string, error) {
	return t.genToken(map[string]interface{}{
		"userID": userID,
		"exp":    time.Now().Add(t.accessExp).Unix(),
	})
}

func (t TokenGeneratorImp) GenRefreshToken(userID int) (string, error) {
	return t.genToken(map[string]interface{}{
		"userID": userID,
		"exp":    time.Now().Add(t.refreshExp).Unix(),
	})
}
