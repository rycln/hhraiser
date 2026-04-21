package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientGetCookieValue(t *testing.T) {
	t.Run("returns cookie value when present", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc", Path: "/"})
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, err := New()
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		value, ok := client.GetCookieValue(server.URL, "session")
		require.True(t, ok)
		assert.Equal(t, "abc", value)
	})

	t.Run("returns false on invalid url", func(t *testing.T) {
		client, err := New()
		require.NoError(t, err)

		value, ok := client.GetCookieValue("://invalid", "session")
		assert.False(t, ok)
		assert.Empty(t, value)
	})
}

func TestHeaderTransportRoundTrip(t *testing.T) {
	t.Run("sets user-agent header", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "agent-test", r.Header.Get("User-Agent"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		transport := &headerTransport{
			base:      http.DefaultTransport,
			userAgent: "agent-test",
		}

		client := &http.Client{Transport: transport}
		resp, err := client.Get(server.URL)
		require.NoError(t, err)
		resp.Body.Close()
	})
}
