package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSchedule(t *testing.T) {
	loc := time.UTC

	t.Run("fails when empty", func(t *testing.T) {
		_, err := NewSchedule(nil, 0, loc)
		require.Error(t, err)
	})

	t.Run("fails when location is nil", func(t *testing.T) {
		_, err := NewSchedule([]TimeOfDay{{Hour: 9, Minute: 0}}, 0, nil)
		require.Error(t, err)
	})

	t.Run("sorts times ascending", func(t *testing.T) {
		s, err := NewSchedule(
			[]TimeOfDay{
				{Hour: 18, Minute: 0},
				{Hour: 9, Minute: 30},
				{Hour: 9, Minute: 0},
			},
			0,
			loc,
		)
		require.NoError(t, err)

		now := time.Date(2026, 4, 21, 8, 0, 0, 0, loc)
		next := s.NextTrigger(now)
		assert.Equal(t, 9, next.Hour())
		assert.Equal(t, 0, next.Minute())
	})
}

func TestScheduleNextTrigger(t *testing.T) {
	loc := time.UTC
	s, err := NewSchedule(
		[]TimeOfDay{
			{Hour: 9, Minute: 0},
			{Hour: 18, Minute: 30},
		},
		0,
		loc,
	)
	require.NoError(t, err)

	t.Run("returns next time in same day", func(t *testing.T) {
		now := time.Date(2026, 4, 21, 10, 0, 0, 0, loc)
		next := s.NextTrigger(now)
		assert.Equal(t, time.Date(2026, 4, 21, 18, 30, 0, 0, loc), next)
	})

	t.Run("rolls to first time tomorrow", func(t *testing.T) {
		now := time.Date(2026, 4, 21, 23, 0, 0, 0, loc)
		next := s.NextTrigger(now)
		assert.Equal(t, time.Date(2026, 4, 22, 9, 0, 0, 0, loc), next)
	})
}

func TestScheduleJitteredDelay(t *testing.T) {
	loc := time.UTC
	t.Run("returns zero when jitter disabled", func(t *testing.T) {
		s, err := NewSchedule([]TimeOfDay{{Hour: 9, Minute: 0}}, 0, loc)
		require.NoError(t, err)
		assert.Zero(t, s.JitteredDelay())
	})

	t.Run("returns delay within jitter bound", func(t *testing.T) {
		s, err := NewSchedule([]TimeOfDay{{Hour: 9, Minute: 0}}, time.Second, loc)
		require.NoError(t, err)

		delay := s.JitteredDelay()
		assert.GreaterOrEqual(t, delay, time.Duration(0))
		assert.Less(t, delay, time.Second)
	})
}
