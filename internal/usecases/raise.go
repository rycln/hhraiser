package usecases

import (
	"context"
	"time"

	"github.com/rycln/hhraiser/internal/domain"
)

type raiseGateway interface {
	Raise(context.Context, *domain.Resume, *domain.Session) error
}

type Raise struct {
	gateway raiseGateway
	timeout time.Duration
}

func NewRaise(gateway raiseGateway, timeout time.Duration) *Raise {
	return &Raise{
		gateway: gateway,
		timeout: timeout,
	}
}

func (uc *Raise) RaiseResume(ctx context.Context, resume *domain.Resume, session *domain.Session) error {
	ctxRaise, cancelRaise := context.WithTimeout(ctx, uc.timeout)
	defer cancelRaise()

	err := uc.gateway.Raise(ctxRaise, resume, session)
	if err != nil {
		return err
	}

	return nil
}
