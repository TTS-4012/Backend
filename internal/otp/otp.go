package otp

import (
	"fmt"
	"math/rand"
	"ocontest/pkg"
)

type OTPStorage interface {
	GenRegisterOTP(userID string) (string, error)
	GenForgotPasswordOTP(userID string) (string, error)
	CheckRegisterOTP(userID, otp string) error
	CheckForgotPasswordOTP(userID, otp string) error
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

func (o *OTPStorageMapImp) getForgetPasswordOTPKey(userID string) string {
	return fmt.Sprintf("forget/%v", userID)
}

func (o *OTPStorageMapImp) GenRegisterOTP(userID string) (string, error) {
	k := o.getRegisterOTPKey(userID)
	v := fmt.Sprintf("%06d", rand.Intn(1000000))
	return v, o.set(k, v)
}

func (o *OTPStorageMapImp) GenForgotPasswordOTP(userID string) (string, error) {
	k := o.getForgetPasswordOTPKey(userID)
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

func (o *OTPStorageMapImp) CheckForgotPasswordOTP(userID, otp string) error {
	k := o.getForgetPasswordOTPKey(userID)
	ans, err := o.get(k)
	if err != nil {
		return err
	}
	if ans != otp {
		return pkg.ErrForbidden
	}
	return nil
}
