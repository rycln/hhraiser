package domain

import "time"

type RaiseEvent struct {
	ResumeTitle string
	Success     bool
	StatusCode  int
	Timestamp   time.Time
}
