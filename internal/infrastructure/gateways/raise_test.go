package gateways

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rycln/hhraiser/internal/domain"
	"github.com/rycln/hhraiser/internal/infrastructure/httpclient"
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
			if tt.wantError == nil && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if tt.wantError != nil && err != tt.wantError {
				t.Fatalf("expected error %v, got %v", tt.wantError, err)
			}
		})
	}
}
