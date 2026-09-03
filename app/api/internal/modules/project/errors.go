package project

import "errors"

var (
	ErrValidation        = errors.New("validation error")
	ErrProjectNotFound   = errors.New("project not found")
	ErrSlugAlreadyExists = errors.New("slug already exists")
)
