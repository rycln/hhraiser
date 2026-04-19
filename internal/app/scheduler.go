package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/rycln/hhraiser/internal/domain"
)

type raiseUsecase interface {
	RaiseResume(ctx context.Context, resume *domain.Resume, delay time.Duration) error
}

type Scheduler struct {
	uc       raiseUsecase
	resume   *domain.Resume
	schedule *domain.Schedule
}

func NewScheduler(uc raiseUsecase, resume *domain.Resume, schedule *domain.Schedule) *Scheduler {
	return &Scheduler{
		uc:       uc,
		resume:   resume,
		schedule: schedule,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	for {
		next := s.schedule.NextTrigger(time.Now())
		slog.Info("next raise scheduled", "at", next.Format(time.DateTime))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Until(next)):
		}

		delay := s.schedule.JitteredDelay()
		err := s.uc.RaiseResume(ctx, s.resume, delay)
		if err != nil {
			slog.Error("raise failed",
				"error", err,
			)
		}
	}
}
