package config

import (
	"testing"
	"time"
)

func TestValidate_HTTPTimeoutMustBePositive(t *testing.T) {
	cfg := &Config{
		TZ: "UTC",
		Schedule: ScheduleConfig{
			Times:  []string{"09:00"},
			Jitter: time.Minute,
		},
		HTTP: HTTPConfig{
			Timeout: 0,
		},
	}

	if err := validate(cfg); err == nil {
		t.Fatal("expected validation error for non-positive HTTP timeout")
	}
}

func TestValidate_HTTPTimeoutPositivePasses(t *testing.T) {
	cfg := &Config{
		TZ: "UTC",
		Schedule: ScheduleConfig{
			Times:  []string{"09:00"},
			Jitter: time.Minute,
		},
		HTTP: HTTPConfig{
			Timeout: time.Second,
		},
	}

	if err := validate(cfg); err != nil {
		t.Fatalf("expected validation success, got %v", err)
	}
}
