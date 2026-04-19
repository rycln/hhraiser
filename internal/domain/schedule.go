package domain

import (
	"cmp"
	"fmt"
	"math/rand/v2"
	"slices"
	"time"
)

type TimeOfDay struct {
	Hour   int
	Minute int
}

type Schedule struct {
	times  []TimeOfDay
	jitter time.Duration
	loc    *time.Location
}

func NewSchedule(times []TimeOfDay, jitter time.Duration, loc *time.Location) (*Schedule, error) {
	if len(times) == 0 {
		return nil, fmt.Errorf("schedule must have at least one time")
	}
	if loc == nil {
		return nil, fmt.Errorf("location must not be nil")
	}

	sorted := make([]TimeOfDay, len(times))
	copy(sorted, times)
	slices.SortFunc(sorted, func(a, b TimeOfDay) int {
		if a.Hour != b.Hour {
			return cmp.Compare(a.Hour, b.Hour)
		}
		return cmp.Compare(a.Minute, b.Minute)
	})

	return &Schedule{
		times:  sorted,
		jitter: jitter,
		loc:    loc,
	}, nil
}

func (s *Schedule) NextTrigger(now time.Time) time.Time {
	now = now.In(s.loc)

	for _, t := range s.times {
		candidate := time.Date(
			now.Year(), now.Month(), now.Day(),
			t.Hour, t.Minute, 0, 0,
			s.loc,
		)
		if candidate.After(now) {
			return candidate
		}
	}

	tomorrow := now.AddDate(0, 0, 1)
	first := s.times[0]
	return time.Date(
		tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
		first.Hour, first.Minute, 0, 0,
		s.loc,
	)
}

func (s *Schedule) JitteredDelay() time.Duration {
	if s.jitter <= 0 {
		return 0
	}
	return time.Duration(rand.Int64N(int64(s.jitter)))
}
