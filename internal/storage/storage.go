package storage

import "errors"

var (
	ErrUserExists      = errors.New("user already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrAdminExists     = errors.New("user already admin")
	ErrAdminNotFound   = errors.New("admin not found")
	ErrCreatorNotFound = errors.New("admin not found")
	ErrAppNotFound     = errors.New("app not found")
)
