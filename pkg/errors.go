package pkg

import "errors"

var (
	ErrBadRequest          = errors.New("bad request")
	ErrExpired             = errors.New("expired")
	ErrNotFound            = errors.New("not found")
	ErrInternalServerError = errors.New("something is wrong with server")
	ErrForbidden           = errors.New("forbidden")
)
