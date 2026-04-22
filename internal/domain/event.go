package domain

import "time"

const (
	AppEventStarted = "app_started"
	AppEventStopped = "app_stopped"
)

type RaiseEvent struct {
	ResumeTitle string
	Success     bool
	StatusCode  int
	Timestamp   time.Time
}

type AppEvent struct {
	Event     string
	Timestamp time.Time
}
