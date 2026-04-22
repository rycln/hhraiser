package gateways

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rycln/hhraiser/internal/domain"
	"github.com/rycln/hhraiser/internal/infrastructure/httpclient"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	t.Run("succeeds on redirect with cookies", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodHead && r.URL.Path == "/":
				http.SetCookie(w, &http.Cookie{Name: "_xsrf", Value: "bootstrap-xsrf", Path: "/"})
				w.WriteHeader(http.StatusOK)
			case r.Method == http.MethodPost && r.URL.Path == loginEndpoint:
				http.SetCookie(w, &http.Cookie{Name: "_xsrf", Value: "final-xsrf", Path: "/"})
				http.SetCookie(w, &http.Cookie{Name: "hhtoken", Value: "final-token", Path: "/"})
				w.Header().Set("Location", "/")
				w.WriteHeader(http.StatusFound)
			case r.Method == http.MethodGet && r.URL.Path == "/":
				w.WriteHeader(http.StatusOK)
			default:
				t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
		}))
		defer server.Close()

		client, err := httpclient.New()
		require.NoError(t, err)

		gw := NewGatewayWithURL(client, server.URL)
		creds := domain.NewCredentials("phone", "password")
		session, err := gw.Login(context.Background(), creds)
		require.NoError(t, err)
		require.True(t, session.IsAuthenticated())
	})

	t.Run("fails when bootstrap does not set xsrf", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodHead && r.URL.Path == "/" {
				w.WriteHeader(http.StatusOK)
				return
			}
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}))
		defer server.Close()

		client, err := httpclient.New()
		require.NoError(t, err)

		gw := NewGatewayWithURL(client, server.URL)
		creds := domain.NewCredentials("phone", "password")
		_, err = gw.Login(context.Background(), creds)
		require.Error(t, err)
	})

	t.Run("fails when bootstrap status is error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodHead && r.URL.Path == "/" {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}))
		defer server.Close()

		client, err := httpclient.New()
		require.NoError(t, err)

		gw := NewGatewayWithURL(client, server.URL)
		creds := domain.NewCredentials("phone", "password")
		_, err = gw.Login(context.Background(), creds)
		require.Error(t, err)
	})
}
