package entity

import "errors"

var (
	ErrEmailExist      = errors.New("email already exists")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidName     = errors.New("invalid name")
	ErrEmailNotFound   = errors.New("email not found")
)
