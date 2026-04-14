package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Phone       string `env:"HH_PHONE"`
	Password    string `env:"HH_PASSWORD"`
	ResumeID    string `env:"HH_RESUME_ID"`
	ResumeTitle string `env:"HH_RESUME_TITLE"`
}

func Load() (*Config, error) {
	envPath := filepath.Join(getConfigDir(), ".env")

	err := godotenv.Load(envPath)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
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

func GetTokensPath() string {
	return filepath.Join(getConfigDir(), "tokens.json")
}
