package gateways

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rycln/hhraiser/internal/domain"
	"github.com/rycln/hhraiser/internal/infrastructure/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRaise_StatusCodeMapping(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		wantError error
	}{
		{name: "ok", status: http.StatusOK, wantError: nil},
		{name: "too early", status: http.StatusConflict, wantError: domain.ErrRaiseTooEarly},
		{name: "unauthorized", status: http.StatusUnauthorized, wantError: domain.ErrRaiseAuthRequired},
		{name: "forbidden", status: http.StatusForbidden, wantError: domain.ErrRaiseAuthRequired},
		{name: "unexpected", status: http.StatusInternalServerError, wantError: domain.ErrRaiseUnexpectedResponse},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != raiseEndpoint {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			client, err := httpclient.New()
			if err != nil {
				t.Fatalf("build client: %v", err)
			}

			gw := NewGatewayWithURL(client, server.URL)
			resume := domain.NewResume("resume-id", "resume")
			session := domain.NewSession("xsrf", "token")

			err = gw.Raise(context.Background(), resume, session)
			if tt.wantError == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantError))
			if tt.status == http.StatusInternalServerError {
				var statusErr *domain.ErrUnexpectedStatus
				require.True(t, errors.As(err, &statusErr))
				assert.Equal(t, http.StatusInternalServerError, statusErr.Code)
			}
		})
	}
}
