package usecases

import (
	"context"
	"time"

	"github.com/rycln/hhraiser/internal/domain"
)

type Tokens struct {
	XSRF  string
	Token string
}

type LoginRequest struct {
	Phone    string
	Password string
	XSRF     string
	Token    string
}

type authGateway interface {
	GetAnonymousTokens(context.Context) (Tokens, error)
	Login(context.Context, LoginRequest) (*domain.Session, error)
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

	tokens, err := uc.gateway.GetAnonymousTokens(ctxTokens)
	if err != nil {
		return err
	}

	req := LoginRequest{
		Phone:    creds.GetPhone(),
		Password: creds.GetPassword(),
		XSRF:     tokens.XSRF,
		Token:    tokens.Token,
	}

	ctxLogin, cancelLogin := context.WithTimeout(ctx, uc.timeout)
	defer cancelLogin()

	session, err := uc.gateway.Login(ctxLogin, req)
	if err != nil {
		return err
	}

	return uc.sessionRepo.Save(session)
}
