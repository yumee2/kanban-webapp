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

// Moodle integration related errors
var (
	ErrMoodleConnectionNotFound = errors.New("moodle connection not found")
	ErrMoodleAuthFailed         = errors.New("moodle authentication failed")
	ErrMoodleCourseNotFound     = errors.New("moodle course not found")
	ErrMoodleTokenKeyMissing    = errors.New("moodle token encryption key is missing")
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
	ErrListNotFound     = errors.New("list not found")
	ErrInvalidListTitle = errors.New("list title cannot be empty")
	ErrInvalidPosition  = errors.New("invalid position value")
)

// Cards related errors
var (
	ErrCardNotFound        = errors.New("card not found")
	ErrInvalidCardTitle    = errors.New("card title cannot be empty")
	ErrInvalidListID       = errors.New("invalid list ID")
	ErrInvalidCardPriority = errors.New("invalid card priority")
)

// Tags related errors
var (
	ErrTagNotFound        = errors.New("tag not found")
	ErrInvalidTagName     = errors.New("tag name cannot be empty")
	ErrInvalidTagColor    = errors.New("tag color cannot be empty")
	ErrTagAlreadyAttached = errors.New("tag already attached to card")
)
