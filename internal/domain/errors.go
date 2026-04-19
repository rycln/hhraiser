package domain

import "errors"

var (
	ErrRaiseUnexpectedResponse = errors.New("raise failed: unexpected response")
	ErrRaiseAuthRequired       = errors.New("raise failed: authentication required")
	ErrRaiseTooEarly           = errors.New("raise too early: interval not elapsed")
	ErrEmptySchedule           = errors.New("schedule is empty")
)
