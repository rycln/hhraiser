package usecases

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/rycln/hhraiser/internal/domain"
)

type authGateway interface {
	Login(context.Context, *domain.Credentials) (*domain.Session, error)
}

type raiseGateway interface {
	Raise(context.Context, *domain.Resume, *domain.Session) error
}

type notifier interface {
	Notify(ctx context.Context, event domain.RaiseEvent) error
}

type RaiseUsecase struct {
	auth     authGateway
	raise    raiseGateway
	notifier notifier
	creds    *domain.Credentials
	session  *domain.Session
	timeout  time.Duration
}

func NewRaise(auth authGateway, raise raiseGateway, notifier notifier, creds *domain.Credentials, session *domain.Session, timeout time.Duration) *RaiseUsecase {
	return &RaiseUsecase{
		auth:     auth,
		raise:    raise,
		notifier: notifier,
		creds:    creds,
		session:  session,
		timeout:  timeout,
	}
}

func (uc *RaiseUsecase) RaiseResume(ctx context.Context, resume *domain.Resume, delay time.Duration) error {
	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	err := uc.doRaise(ctx, resume)

	event := domain.RaiseEvent{
		ResumeTitle: resume.GetTitle(),
		Success:     err == nil,
		Timestamp:   time.Now(),
	}

	var statusErr *domain.ErrUnexpectedStatus
	if errors.As(err, &statusErr) {
		event.StatusCode = statusErr.Code
	}

	if notifyErr := uc.notifier.Notify(ctx, event); notifyErr != nil {
		slog.WarnContext(ctx, "failed to send notification", "error", notifyErr)
	}

	return err
}

func (uc *RaiseUsecase) doRaise(ctx context.Context, resume *domain.Resume) error {
	if !uc.session.IsAuthenticated() {
		if err := uc.authenticate(ctx); err != nil {
			return err
		}
	}

	err := uc.raiseResume(ctx, resume)
	if errors.Is(err, domain.ErrRaiseAuthRequired) {
		slog.InfoContext(ctx, "re-auth attempt")
		uc.session = nil
		if err := uc.authenticate(ctx); err != nil {
			return err
		}

		return uc.raiseResume(ctx, resume)
	}

	return err
}

func (uc *RaiseUsecase) authenticate(ctx context.Context) error {
	ctxLogin, cancelLogin := context.WithTimeout(ctx, uc.timeout)
	defer cancelLogin()

	session, err := uc.auth.Login(ctxLogin, uc.creds)
	if err != nil {
		return err
	}

	uc.session = session

	return nil
}

func (uc *RaiseUsecase) raiseResume(ctx context.Context, resume *domain.Resume) error {
	ctxRaise, cancelRaise := context.WithTimeout(ctx, uc.timeout)
	defer cancelRaise()

	return uc.raise.Raise(ctxRaise, resume, uc.session)
}
