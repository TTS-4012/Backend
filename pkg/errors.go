package pkg

import "errors"

var (
	ErrBadRequest          = errors.New("bad request")
	ErrExpired             = errors.New("expired")
	ErrInternalServerError = errors.New("something is wrong with server")
	ErrForbidden           = errors.New("forbidden")
)
