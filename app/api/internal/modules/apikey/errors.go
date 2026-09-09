package apikey

import "errors"

var (
	ErrValidation      = errors.New("validation error")
	ErrKeyNotFound     = errors.New("api key not found")
	ErrProjectNotFound = errors.New("project not found")
)
