package user

import "errors"

var (
	ErrValidation             = errors.New("validation error")
	ErrUserEmailAlreadyExists = errors.New("user email already exists")
	ErrUserNotFound           = errors.New("user not found")
)
