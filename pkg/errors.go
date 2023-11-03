package pkg

import "errors"

var (
	ErrBadRequest          = errors.New("bad request")
	ErrInternalServerError = errors.New("something is wrong with server")
)
