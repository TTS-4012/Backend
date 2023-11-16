package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var iv = []byte{34, 35, 35, 57, 68, 4, 35, 36, 7, 8, 35, 23, 35, 86, 35, 23}

type AESHandler interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type AESHandlerImp struct {
	b64encoding *base64.Encoding
	cipher      cipher.Block
}

func NewAesHandler(key []byte) (AESHandler, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ans := &AESHandlerImp{
		cipher:      c,
		b64encoding: base64.URLEncoding,
	}
	return ans, nil
}

func (a *AESHandlerImp) Encrypt(text string) (string, error) {

	cfb := cipher.NewCFBEncrypter(a.cipher, iv)
	ciphertext := make([]byte, len(text))
	cfb.XORKeyStream(ciphertext, []byte(text))

	return a.b64encoding.EncodeToString(ciphertext), nil
}

func (a *AESHandlerImp) Decrypt(text string) (string, error) {

	ciphertext, err := a.b64encoding.DecodeString(text)
	if err != nil {
		return "", err
	}

	cfb := cipher.NewCFBDecrypter(a.cipher, iv)
	plaintext := make([]byte, len(ciphertext))
	cfb.XORKeyStream(plaintext, ciphertext)
	return string(plaintext), nil
}
