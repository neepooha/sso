package storage

import "errors"

var (
	ErrAppExists  = errors.New("app already exists")
	ErrUserExists = errors.New("user already exists")

	ErrAppNotFound     = errors.New("app not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrAdminNotFound   = errors.New("admin not found")
	ErrCreatorNotFound = errors.New("creator not found")

	ErrAdminExists = errors.New("user already admin")
)
