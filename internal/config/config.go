package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/rycln/hhraiser/internal/domain"
)

type Config struct {
	HH       HHConfig       `envPrefix:"HH_"`
	Schedule ScheduleConfig `envPrefix:"SCHEDULE_"`
	HTTP     HTTPConfig     `envPrefix:"HTTP_"`
	TZ       string         `env:"TZ"        envDefault:"UTC"`
	LogLevel string         `env:"LOG_LEVEL" envDefault:"info"`
}

type HHConfig struct {
	Phone       string `env:"PHONE,required"`
	Password    string `env:"PASSWORD,required"`
	ResumeID    string `env:"RESUME_ID,required"`
	ResumeTitle string `env:"RESUME_TITLE"`
}

type ScheduleConfig struct {
	Times  []string      `env:"TIMES,required" envSeparator:","`
	Jitter time.Duration `env:"JITTER"         envDefault:"5m"`
}

type HTTPConfig struct {
	Timeout time.Duration `env:"TIMEOUT" envDefault:"30s"`
}

func Load() (*Config, error) {
	envPath := filepath.Join(getConfigDir(), ".env")

	if err := godotenv.Load(envPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func validate(cfg *Config) error {
	if _, err := time.LoadLocation(cfg.TZ); err != nil {
		return fmt.Errorf("TZ=%q is not a valid timezone: %w", cfg.TZ, err)
	}

	for _, raw := range cfg.Schedule.Times {
		if _, err := ParseTime(raw); err != nil {
			return fmt.Errorf("SCHEDULE_TIMES: %w", err)
		}
	}

	if cfg.Schedule.Jitter < 0 {
		return fmt.Errorf("SCHEDULE_JITTER must be non-negative")
	}

	return nil
}

func ParseTime(s string) (domain.TimeOfDay, error) {
	s = strings.TrimSpace(s)
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return domain.TimeOfDay{}, fmt.Errorf("expected HH:MM, got %q", s)
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return domain.TimeOfDay{}, fmt.Errorf("invalid hour in %q", s)
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return domain.TimeOfDay{}, fmt.Errorf("invalid minute in %q", s)
	}

	return domain.TimeOfDay{Hour: hour, Minute: minute}, nil
}

func ParseTimes(raw []string) ([]domain.TimeOfDay, error) {
	times := make([]domain.TimeOfDay, 0, len(raw))
	for _, s := range raw {
		t, err := ParseTime(s)
		if err != nil {
			return nil, err
		}
		times = append(times, t)
	}
	return times, nil
}

func getConfigDir() string {
	if dir := os.Getenv("HH_CONFIG_DIR"); dir != "" {
		return dir
	}

	if _, err := os.Stat("/config"); err == nil {
		if isRunningInContainer() {
			return "/config"
		}
	}

	return "./config"
}

func isRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		return strings.Contains(string(data), "docker")
	}

	return false
}
