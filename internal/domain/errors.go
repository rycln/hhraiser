package domain

import (
	"errors"
	"fmt"
)

var (
	ErrRaiseUnexpectedResponse = errors.New("raise failed: unexpected response")
	ErrRaiseAuthRequired       = errors.New("raise failed: authentication required")
	ErrRaiseTooEarly           = errors.New("raise too early: interval not elapsed")
	ErrEmptySchedule           = errors.New("schedule is empty")
)

type ErrUnexpectedStatus struct {
	Code int
}

func (e *ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("unexpected status code: %d", e.Code)
}

func (e *ErrUnexpectedStatus) Unwrap() error {
	return ErrRaiseUnexpectedResponse
}
