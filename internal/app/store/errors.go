package store

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrRecordNotFound = errors.New("record not found")
)
