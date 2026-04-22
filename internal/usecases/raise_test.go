package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	errs       []error
	raiseCalls int
	ctxs       []context.Context
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

type notifierStub struct {
	events []domain.RaiseEvent
	err    error
}

func (s *notifierStub) Notify(_ context.Context, event domain.RaiseEvent) error {
	s.events = append(s.events, event)
	return s.err
}

func TestRaiseResume(t *testing.T) {
	t.Run("reauth only on auth error", func(t *testing.T) {
		creds := domain.NewCredentials("phone", "password")
		session := domain.NewSession("xsrf", "token")
		resume := domain.NewResume("resume-id", "resume")

		auth := &authGatewayStub{session: session}
		raise := &raiseGatewayStub{errs: []error{domain.ErrRaiseAuthRequired, nil}}
		notify := &notifierStub{}

		uc := NewRaise(auth, raise, notify, creds, session, time.Second, true)
		err := uc.RaiseResume(context.Background(), resume, 0)

		require.NoError(t, err)
		assert.Equal(t, 1, auth.loginCalls)
		assert.Equal(t, 2, raise.raiseCalls)
		require.Len(t, notify.events, 1)
		assert.True(t, notify.events[0].Success)
		assert.Equal(t, "resume", notify.events[0].ResumeTitle)
	})

	t.Run("does not reauth on unexpected error", func(t *testing.T) {
		creds := domain.NewCredentials("phone", "password")
		session := domain.NewSession("xsrf", "token")
		resume := domain.NewResume("resume-id", "resume")

		auth := &authGatewayStub{session: session}
		raise := &raiseGatewayStub{errs: []error{domain.ErrRaiseUnexpectedResponse}}
		notify := &notifierStub{}

		uc := NewRaise(auth, raise, notify, creds, session, time.Second, true)
		err := uc.RaiseResume(context.Background(), resume, 0)

		require.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrRaiseUnexpectedResponse))
		assert.Equal(t, 0, auth.loginCalls)
		assert.Equal(t, 1, raise.raiseCalls)
		require.Len(t, notify.events, 1)
		assert.False(t, notify.events[0].Success)
	})

	t.Run("sets status code on notification event for unexpected status", func(t *testing.T) {
		creds := domain.NewCredentials("phone", "password")
		session := domain.NewSession("xsrf", "token")
		resume := domain.NewResume("resume-id", "resume")

		auth := &authGatewayStub{session: session}
		raise := &raiseGatewayStub{errs: []error{&domain.ErrUnexpectedStatus{Code: 500}}}
		notify := &notifierStub{}

		uc := NewRaise(auth, raise, notify, creds, session, time.Second, true)
		err := uc.RaiseResume(context.Background(), resume, 0)

		require.Error(t, err)
		require.Len(t, notify.events, 1)
		assert.Equal(t, 500, notify.events[0].StatusCode)
		assert.False(t, notify.events[0].Success)
	})

	t.Run("uses timeout contexts for auth and raise", func(t *testing.T) {
		timeout := 150 * time.Millisecond
		creds := domain.NewCredentials("phone", "password")
		resume := domain.NewResume("resume-id", "resume")

		auth := &authGatewayStub{session: domain.NewSession("xsrf", "token")}
		raise := &raiseGatewayStub{}
		notify := &notifierStub{}

		uc := NewRaise(auth, raise, notify, creds, nil, timeout, true)
		err := uc.RaiseResume(context.Background(), resume, 0)
		require.NoError(t, err)

		require.Equal(t, 1, auth.loginCalls)
		require.Len(t, auth.ctxs, 1)
		require.Equal(t, 1, raise.raiseCalls)
		require.Len(t, raise.ctxs, 1)

		authDeadline, ok := auth.ctxs[0].Deadline()
		require.True(t, ok)
		raiseDeadline, ok := raise.ctxs[0].Deadline()
		require.True(t, ok)

		authRemaining := time.Until(authDeadline)
		raiseRemaining := time.Until(raiseDeadline)
		assert.Greater(t, authRemaining, time.Duration(0))
		assert.LessOrEqual(t, authRemaining, timeout)
		assert.Greater(t, raiseRemaining, time.Duration(0))
		assert.LessOrEqual(t, raiseRemaining, timeout)
	})

	t.Run("returns context cancellation while waiting delay", func(t *testing.T) {
		creds := domain.NewCredentials("phone", "password")
		resume := domain.NewResume("resume-id", "resume")
		auth := &authGatewayStub{session: domain.NewSession("xsrf", "token")}
		raise := &raiseGatewayStub{}
		notify := &notifierStub{}
		uc := NewRaise(auth, raise, notify, creds, nil, time.Second, true)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := uc.RaiseResume(ctx, resume, time.Second)
		require.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
		assert.Equal(t, 0, auth.loginCalls)
		assert.Equal(t, 0, raise.raiseCalls)
		assert.Empty(t, notify.events)
	})

	t.Run("does not notify on success when notifyOnSuccess is false", func(t *testing.T) {
		creds := domain.NewCredentials("phone", "password")
		resume := domain.NewResume("resume-id", "resume")
		auth := &authGatewayStub{}
		raise := &raiseGatewayStub{}
		notify := &notifierStub{}
		session := domain.NewSession("xsrf", "token")
		uc := NewRaise(auth, raise, notify, creds, session, time.Second, false)

		err := uc.RaiseResume(context.Background(), resume, 0)
		require.NoError(t, err)
		assert.Empty(t, notify.events)
	})

	t.Run("notifies on failure regardless of notifyOnSuccess flag", func(t *testing.T) {
		creds := domain.NewCredentials("phone", "password")
		resume := domain.NewResume("resume-id", "resume")
		auth := &authGatewayStub{}
		raise := &raiseGatewayStub{errs: []error{&domain.ErrUnexpectedStatus{Code: 500}}}
		notify := &notifierStub{}
		session := domain.NewSession("xsrf", "token")
		uc := NewRaise(auth, raise, notify, creds, session, time.Second, false)

		err := uc.RaiseResume(context.Background(), resume, 0)
		require.Error(t, err)
		assert.Len(t, notify.events, 1)
		assert.False(t, notify.events[0].Success)
	})
}
