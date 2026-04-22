package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	base := &Config{
		TZ: "UTC",
		Schedule: ScheduleConfig{
			Times:  []string{"09:00"},
			Jitter: time.Minute,
		},
		HTTP: HTTPConfig{
			Timeout: time.Second,
		},
	}

	t.Run("http timeout must be positive", func(t *testing.T) {
		cfg := *base
		cfg.HTTP.Timeout = 0
		err := validate(&cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP_TIMEOUT")
	})

	t.Run("invalid timezone fails", func(t *testing.T) {
		cfg := *base
		cfg.TZ = "not/a-timezone"
		err := validate(&cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "TZ=")
	})

	t.Run("invalid schedule time fails", func(t *testing.T) {
		cfg := *base
		cfg.Schedule.Times = []string{"25:00"}
		err := validate(&cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SCHEDULE_TIMES")
	})

	t.Run("negative jitter fails", func(t *testing.T) {
		cfg := *base
		cfg.Schedule.Jitter = -time.Second
		err := validate(&cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SCHEDULE_JITTER")
	})

	t.Run("valid config passes", func(t *testing.T) {
		cfg := *base
		require.NoError(t, validate(&cfg))
	})
}

func TestParseTime(t *testing.T) {
	t.Run("valid value parses", func(t *testing.T) {
		got, err := ParseTime(" 09:05 ")
		require.NoError(t, err)
		assert.Equal(t, 9, got.Hour)
		assert.Equal(t, 5, got.Minute)
	})

	t.Run("missing separator fails", func(t *testing.T) {
		_, err := ParseTime("0905")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected HH:MM")
	})

	t.Run("invalid hour fails", func(t *testing.T) {
		_, err := ParseTime("24:00")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid hour")
	})

	t.Run("invalid minute fails", func(t *testing.T) {
		_, err := ParseTime("23:60")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid minute")
	})
}

func TestParseTimes(t *testing.T) {
	t.Run("all valid values parse", func(t *testing.T) {
		got, err := ParseTimes([]string{"09:00", "18:30"})
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, 9, got[0].Hour)
		assert.Equal(t, 30, got[1].Minute)
	})

	t.Run("any invalid value fails", func(t *testing.T) {
		_, err := ParseTimes([]string{"09:00", "invalid"})
		require.Error(t, err)
	})
}
