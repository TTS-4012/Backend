package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

type AESHandler interface {
	Encrypt(plaintext []byte) []byte
	Decrypt(ciphertext []byte) []byte
}

type AESHandlerImp struct {
	cipher cipher.Block
}

func NewAesHandler(key []byte) (AESHandler, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ans := &AESHandlerImp{cipher: c}
	return ans, nil
}

func (a *AESHandlerImp) Encrypt(plaintext []byte) []byte {

	// allocate space for ciphered data
	out := make([]byte, len(plaintext))

	// encrypt
	a.cipher.Encrypt(out, plaintext)

	return out
}

func (a *AESHandlerImp) Decrypt(encryptedtext []byte) []byte {

	// allocate space for ciphered data
	out := make([]byte, len(encryptedtext))

	// encrypt
	a.cipher.Decrypt(out, encryptedtext)

	return out
}

func main() {
	sec := "aaaaaaaaaaaaaaaa"
	a, err := NewAesHandler([]byte(sec))
	if err != nil {
		panic(err)
	}
	pt := "salam khoobie?jkjkjkjkjkjk"
	enc := a.Encrypt([]byte(pt))
	dec := string(a.Decrypt(enc))
	fmt.Println(enc, dec)

}
