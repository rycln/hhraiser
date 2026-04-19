package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rycln/hhraiser/internal/domain"
)

type authGatewayStub struct {
	loginCalls int
	session    *domain.Session
	err        error
	ctxs       []context.Context
}

func (s *authGatewayStub) Login(ctx context.Context, _ *domain.Credentials) (*domain.Session, error) {
	s.loginCalls++
	s.ctxs = append(s.ctxs, ctx)
	if s.err != nil {
		return nil, s.err
	}
	return s.session, nil
}

type raiseGatewayStub struct {
	errs      []error
	raiseCalls int
	ctxs      []context.Context
}

func (s *raiseGatewayStub) Raise(ctx context.Context, _ *domain.Resume, _ *domain.Session) error {
	s.raiseCalls++
	s.ctxs = append(s.ctxs, ctx)
	idx := s.raiseCalls - 1
	if idx < len(s.errs) {
		return s.errs[idx]
	}
	return nil
}

func TestRaiseResume_ReauthOnlyOnAuthError(t *testing.T) {
	creds := domain.NewCredentials("phone", "password")
	session := domain.NewSession("xsrf", "token")
	resume := domain.NewResume("resume-id", "resume")

	auth := &authGatewayStub{session: session}
	raise := &raiseGatewayStub{errs: []error{domain.ErrRaiseAuthRequired, nil}}

	uc := NewRaise(auth, raise, creds, session, time.Second)
	err := uc.RaiseResume(context.Background(), resume, 0)
	if err != nil {
		t.Fatalf("expected success after re-auth, got error: %v", err)
	}
	if auth.loginCalls != 1 {
		t.Fatalf("expected one re-auth call, got %d", auth.loginCalls)
	}
	if raise.raiseCalls != 2 {
		t.Fatalf("expected two raise attempts, got %d", raise.raiseCalls)
	}
}

func TestRaiseResume_DoesNotReauthUnexpectedError(t *testing.T) {
	creds := domain.NewCredentials("phone", "password")
	session := domain.NewSession("xsrf", "token")
	resume := domain.NewResume("resume-id", "resume")

	auth := &authGatewayStub{session: session}
	raise := &raiseGatewayStub{errs: []error{domain.ErrRaiseUnexpectedResponse}}

	uc := NewRaise(auth, raise, creds, session, time.Second)
	err := uc.RaiseResume(context.Background(), resume, 0)
	if !errors.Is(err, domain.ErrRaiseUnexpectedResponse) {
		t.Fatalf("expected unexpected response error, got %v", err)
	}
	if auth.loginCalls != 0 {
		t.Fatalf("expected no re-auth for unexpected response, got %d", auth.loginCalls)
	}
	if raise.raiseCalls != 1 {
		t.Fatalf("expected one raise attempt, got %d", raise.raiseCalls)
	}
}

func TestRaiseResume_TimeoutContextsUsedForAuthAndRaise(t *testing.T) {
	timeout := 150 * time.Millisecond
	creds := domain.NewCredentials("phone", "password")
	resume := domain.NewResume("resume-id", "resume")

	auth := &authGatewayStub{session: domain.NewSession("xsrf", "token")}
	raise := &raiseGatewayStub{}

	uc := NewRaise(auth, raise, creds, nil, timeout)
	err := uc.RaiseResume(context.Background(), resume, 0)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if auth.loginCalls != 1 || len(auth.ctxs) != 1 {
		t.Fatalf("expected one auth call with context, got calls=%d ctxs=%d", auth.loginCalls, len(auth.ctxs))
	}
	if raise.raiseCalls != 1 || len(raise.ctxs) != 1 {
		t.Fatalf("expected one raise call with context, got calls=%d ctxs=%d", raise.raiseCalls, len(raise.ctxs))
	}

	authDeadline, ok := auth.ctxs[0].Deadline()
	if !ok {
		t.Fatal("expected auth context with deadline")
	}
	raiseDeadline, ok := raise.ctxs[0].Deadline()
	if !ok {
		t.Fatal("expected raise context with deadline")
	}

	authRemaining := time.Until(authDeadline)
	raiseRemaining := time.Until(raiseDeadline)
	if authRemaining <= 0 || authRemaining > timeout {
		t.Fatalf("auth deadline should be within timeout window, remaining=%v timeout=%v", authRemaining, timeout)
	}
	if raiseRemaining <= 0 || raiseRemaining > timeout {
		t.Fatalf("raise deadline should be within timeout window, remaining=%v timeout=%v", raiseRemaining, timeout)
	}
}
