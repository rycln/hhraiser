package domain

import "errors"

var (
	ErrRaiseUnexpectedResponse = errors.New("raise failed: unexpected response")
	ErrRaiseTooEarly           = errors.New("raise too early: interval not elapsed")
	ErrEmptySchedule           = errors.New("schedule is empty")
)
