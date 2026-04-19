// internal/app/app.go
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/rycln/hhraiser/internal/config"
	"github.com/rycln/hhraiser/internal/domain"
	"github.com/rycln/hhraiser/internal/infrastructure/gateways"
	"github.com/rycln/hhraiser/internal/infrastructure/httpclient"
	"github.com/rycln/hhraiser/internal/usecases"
)

const (
	baseURL    = "https://hh.ru"
	reqTimeout = 10 * time.Second
)

type App struct {
	scheduler *Scheduler
}

func New(cfg *config.Config) (*App, error) {
	loc, err := time.LoadLocation(cfg.TZ)
	if err != nil {
		return nil, fmt.Errorf("load timezone: %w", err)
	}

	times, err := config.ParseTimes(cfg.Schedule.Times)
	if err != nil {
		return nil, fmt.Errorf("parse schedule times: %w", err)
	}

	schedule, err := domain.NewSchedule(times, cfg.Schedule.Jitter, loc)
	if err != nil {
		return nil, fmt.Errorf("build schedule: %w", err)
	}

	client, err := httpclient.New(cfg.HTTP.Timeout)
	if err != nil {
		return nil, fmt.Errorf("build http client: %w", err)
	}

	hhgateway := gateways.NewGateway(client, baseURL)

	creds := domain.NewCredentials(cfg.HH.Phone, cfg.HH.Password)
	var session *domain.Session

	uc := usecases.NewRaise(hhgateway, hhgateway, creds, session, reqTimeout)

	resume := domain.NewResume(cfg.HH.ResumeID, cfg.HH.ResumeTitle)

	scheduler := NewScheduler(uc, resume, schedule)

	return &App{scheduler: scheduler}, nil
}

func (a *App) Run(ctx context.Context) error {
	return a.scheduler.Run(ctx)
}
