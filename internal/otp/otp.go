package otp

import (
	"fmt"
	"github.com/ocontest/backend/pkg"
	"math/rand"
)

type OTPStorage interface {
	GenRegisterOTP(userID string) (string, error)
	GenLoginOTP(userID string) (string, error)
	CheckRegisterOTP(userID, otp string) error
	CheckLoginOTP(userID, otp string) error
}

// TODO: use redis or memcache instead of map
// TODO: rule of DRY has been violated here, fix it
type OTPStorageMapImp struct {
	m map[string]string
}

func NewOTPStorage() OTPStorage {
	return &OTPStorageMapImp{
		m: make(map[string]string),
	}
}

func (o *OTPStorageMapImp) set(key, value string) error {
	o.m[key] = value
	return nil
}

func (o *OTPStorageMapImp) get(key string) (string, error) {
	ans, ok := o.m[key]
	if !ok {
		return "", pkg.ErrNotFound
	}
	return ans, nil
}

func (o *OTPStorageMapImp) getRegisterOTPKey(userID string) string {
	return fmt.Sprintf("register/%v", userID)
}

func (o *OTPStorageMapImp) getLoginOTPKey(userID string) string {
	return fmt.Sprintf("login/%v", userID)
}

func (o *OTPStorageMapImp) GenRegisterOTP(userID string) (string, error) {
	k := o.getRegisterOTPKey(userID)
	v := fmt.Sprintf("%06d", rand.Intn(1000000))
	return v, o.set(k, v)
}

func (o *OTPStorageMapImp) GenLoginOTP(userID string) (string, error) {
	k := o.getLoginOTPKey(userID)
	v := fmt.Sprintf("%06d", rand.Intn(1000000))
	return v, o.set(k, v)
}

func (o *OTPStorageMapImp) CheckRegisterOTP(userID, otp string) error {
	k := o.getRegisterOTPKey(userID)
	ans, err := o.get(k)
	if err != nil {
		return err
	}
	if ans != otp {
		return pkg.ErrForbidden
	}
	return nil
}

func (o *OTPStorageMapImp) CheckLoginOTP(userID, otp string) error {
	k := o.getLoginOTPKey(userID)
	ans, err := o.get(k)
	if err != nil {
		return err
	}
	if ans != otp {
		return pkg.ErrForbidden
	}
	return nil
}
