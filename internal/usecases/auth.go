package usecases

import (
	"context"
	"time"

	"github.com/rycln/hhraiser/internal/domain"
)

type authGateway interface {
	GetAnonymousTokens(context.Context) (*domain.Session, error)
	Login(context.Context, *domain.Credentials, *domain.Session) (*domain.Session, error)
}

type authSessionRepository interface {
	Save(*domain.Session) error
}

type Auth struct {
	sessionRepo authSessionRepository
	gateway     authGateway
	timeout     time.Duration
}

func NewAuth(sessionRepo authSessionRepository, gateway authGateway, timeout time.Duration) *Auth {
	return &Auth{
		sessionRepo: sessionRepo,
		gateway:     gateway,
		timeout:     timeout,
	}
}

func (uc *Auth) Authenticate(ctx context.Context, creds *domain.Credentials) error {
	ctxTokens, cancelTokens := context.WithTimeout(ctx, uc.timeout)
	defer cancelTokens()

	anonSession, err := uc.gateway.GetAnonymousTokens(ctxTokens)
	if err != nil {
		return err
	}

	ctxLogin, cancelLogin := context.WithTimeout(ctx, uc.timeout)
	defer cancelLogin()

	session, err := uc.gateway.Login(ctxLogin, creds, anonSession)
	if err != nil {
		return err
	}

	return uc.sessionRepo.Save(session)
}
