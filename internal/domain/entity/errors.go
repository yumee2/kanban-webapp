package entity

import "errors"

var (
	ErrEmailExist           = errors.New("email already exists")
	ErrInvalidEmail         = errors.New("invalid email")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrInvalidName          = errors.New("invalid name")
	ErrEmailNotFound        = errors.New("email not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
)

// Board related errors
var (
	ErrBoardNotFound   = errors.New("board not found")
	ErrOwnerNotFound   = errors.New("owner not found")
	ErrInvalidBoardID  = errors.New("invalid board ID")
	ErrInvalidOwnerID  = errors.New("invalid owner ID")
	ErrBoardTitleEmpty = errors.New("board title cannot be empty")
)

// List related errors
var (
	ErrListNotFound = errors.New("list not found")
)
