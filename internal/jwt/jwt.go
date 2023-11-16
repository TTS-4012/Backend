package jwt

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"ocontest/pkg"
	"ocontest/pkg/configs"
	"time"
)

type TokenGenerator interface {
	GenToken(ID int64, typ string, expireTime time.Duration) (string, error)
	ParseToken(token string) (ID int64, typ string, err error)
}

type TokenGeneratorImp struct {
	secret     []byte
	accessExp  time.Duration
	refreshExp time.Duration
}

func NewGenerator(conf configs.SectionJWT) TokenGenerator {
	return TokenGeneratorImp{
		secret: []byte(conf.Secret),
	}
}

func (t TokenGeneratorImp) GenToken(userID int64, typ string, expireTime time.Duration) (string, error) {
	if expireTime == 0 {
		return "", pkg.ErrBadRequest
	}

	mapClaims := jwt.MapClaims{}

	mapClaims["iat"] = time.Now().Unix()
	mapClaims["userID"] = userID
	mapClaims["typ"] = typ
	mapClaims["exp"] = time.Now().Add(expireTime).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	return token.SignedString(t.secret)

}

func (t TokenGeneratorImp) ParseToken(token string) (int64, string, error) {
	mapClaims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, mapClaims, func(token *jwt.Token) (interface{}, error) {
		return t.secret, nil
	}, jwt.WithJSONNumber())
	if err != nil {
		return -1, "", err
	}

	pkg.Log.Debug(mapClaims)
	if exp, exists := mapClaims["exp"]; exists {
		expInt, err := exp.(json.Number).Int64()
		if err != nil || expInt < time.Now().Unix() {
			return -1, "", pkg.ErrExpired
		}
		userID, err := mapClaims["userID"].(json.Number).Int64()
		if err != nil {
			return -1, "", pkg.ErrBadRequest
		}
		return userID, mapClaims["typ"].(string), err
	}
	return -1, "", pkg.ErrExpired
}
