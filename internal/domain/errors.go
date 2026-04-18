package domain

import "errors"

var (
	ErrInvalidSession = errors.New("invalid session: authentication needed")
	ErrRaiseTooEarly  = errors.New("raise too early: interval not elapsed")
)
