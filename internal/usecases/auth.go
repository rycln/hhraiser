package usecases

import (
	"context"
	"time"

	"github.com/rycln/hhraiser/internal/domain"
)

type authGateway interface {
	Login(context.Context, *domain.Credentials) (*domain.Session, error)
}

type Auth struct {
	gateway authGateway
	timeout time.Duration
}

func NewAuth(gateway authGateway, timeout time.Duration) *Auth {
	return &Auth{
		gateway: gateway,
		timeout: timeout,
	}
}

func (uc *Auth) Authenticate(ctx context.Context, creds *domain.Credentials) (*domain.Session, error) {
	ctxLogin, cancelLogin := context.WithTimeout(ctx, uc.timeout)
	defer cancelLogin()

	session, err := uc.gateway.Login(ctxLogin, creds)
	if err != nil {
		return nil, err
	}

	return session, nil
}
