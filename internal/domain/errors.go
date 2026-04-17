package domain

import "errors"

var (
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenMissing  = errors.New("token missing: authorization required")
	ErrRaiseTooEarly = errors.New("raise too early: interval not elapsed")
	ErrNotFound      = errors.New("not found")
)
