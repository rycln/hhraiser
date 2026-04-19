package gateways

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rycln/hhraiser/internal/domain"
	"github.com/rycln/hhraiser/internal/infrastructure/httpclient"
)

func TestLogin_SucceedsOnRedirectWithCookies(t *testing.T) {
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
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	gw := NewGatewayWithURL(client, server.URL)
	creds := domain.NewCredentials("phone", "password")
	session, err := gw.Login(context.Background(), creds)
	if err != nil {
		t.Fatalf("expected login success, got %v", err)
	}
	if !session.IsAuthenticated() {
		t.Fatal("expected authenticated session")
	}
}

func TestLogin_FailsWhenBootstrapDoesNotSetXSRF(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead && r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	client, err := httpclient.New()
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	gw := NewGatewayWithURL(client, server.URL)
	creds := domain.NewCredentials("phone", "password")
	if _, err = gw.Login(context.Background(), creds); err == nil {
		t.Fatal("expected login failure when bootstrap xsrf is missing")
	}
}

func TestLogin_FailsWhenBootstrapStatusIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead && r.URL.Path == "/" {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	client, err := httpclient.New()
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	gw := NewGatewayWithURL(client, server.URL)
	creds := domain.NewCredentials("phone", "password")
	if _, err = gw.Login(context.Background(), creds); err == nil {
		t.Fatal("expected login failure on bootstrap error status")
	}
}
