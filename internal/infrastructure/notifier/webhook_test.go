package notifier

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rycln/hhraiser/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func TestWebhookNotifyRaise(t *testing.T) {
	t.Run("skips when url is empty", func(t *testing.T) {
		w := NewWebhook(http.DefaultClient, "", "")
		err := w.NotifyRaise(context.Background(), domain.RaiseEvent{})
		require.NoError(t, err)
	})

	t.Run("sends raise_success payload with auth header", func(t *testing.T) {
		var gotAuth string
		var gotPayload raisePayload

		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
			defer r.Body.Close()
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
			w.WriteHeader(http.StatusOK)
		})

		wh := NewWebhook(server.Client(), server.URL, "secret")
		event := domain.RaiseEvent{
			ResumeTitle: "Go Resume",
			Success:     true,
			Timestamp:   time.Now().UTC(),
		}

		err := wh.NotifyRaise(context.Background(), event)
		require.NoError(t, err)
		assert.Equal(t, "Bearer secret", gotAuth)
		assert.Equal(t, "raise_success", gotPayload.Event)
		assert.Equal(t, "Go Resume", gotPayload.ResumeTitle)
	})

	t.Run("sends raise_failure payload with status code", func(t *testing.T) {
		var gotPayload raisePayload

		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
			w.WriteHeader(http.StatusOK)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		event := domain.RaiseEvent{
			ResumeTitle: "Go Resume",
			Success:     false,
			StatusCode:  500,
			Timestamp:   time.Now().UTC(),
		}

		err := wh.NotifyRaise(context.Background(), event)
		require.NoError(t, err)
		assert.Equal(t, "raise_failure", gotPayload.Event)
		assert.Equal(t, 500, gotPayload.StatusCode)
	})

	t.Run("omits auth header when secret is empty", func(t *testing.T) {
		var gotAuth string

		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		err := wh.NotifyRaise(context.Background(), domain.RaiseEvent{Timestamp: time.Now()})
		require.NoError(t, err)
		assert.Empty(t, gotAuth)
	})

	t.Run("returns error on non-2xx response", func(t *testing.T) {
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		err := wh.NotifyRaise(context.Background(), domain.RaiseEvent{Timestamp: time.Now()})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "non-2xx")
	})
}

func TestWebhookNotifyApp(t *testing.T) {
	t.Run("skips when url is empty", func(t *testing.T) {
		w := NewWebhook(http.DefaultClient, "", "")
		err := w.NotifyApp(context.Background(), domain.AppEvent{})
		require.NoError(t, err)
	})

	t.Run("sends app_started payload", func(t *testing.T) {
		var gotPayload appPayload

		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
			w.WriteHeader(http.StatusOK)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		event := domain.AppEvent{
			Event:     domain.AppEventStarted,
			Timestamp: time.Now().UTC(),
		}

		err := wh.NotifyApp(context.Background(), event)
		require.NoError(t, err)
		assert.Equal(t, domain.AppEventStarted, gotPayload.Event)
	})

	t.Run("sends app_stopped payload", func(t *testing.T) {
		var gotPayload appPayload

		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
			w.WriteHeader(http.StatusOK)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		event := domain.AppEvent{
			Event:     domain.AppEventStopped,
			Timestamp: time.Now().UTC(),
		}

		err := wh.NotifyApp(context.Background(), event)
		require.NoError(t, err)
		assert.Equal(t, domain.AppEventStopped, gotPayload.Event)
	})

	t.Run("returns error on non-2xx response", func(t *testing.T) {
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		err := wh.NotifyApp(context.Background(), domain.AppEvent{
			Event:     domain.AppEventStarted,
			Timestamp: time.Now(),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "non-2xx")
	})
}

func TestResolveRaiseEvent(t *testing.T) {
	t.Run("success resolves to raise_success", func(t *testing.T) {
		assert.Equal(t, "raise_success", resolveRaiseEvent(domain.RaiseEvent{Success: true}))
	})

	t.Run("failure resolves to raise_failure", func(t *testing.T) {
		assert.Equal(t, "raise_failure", resolveRaiseEvent(domain.RaiseEvent{Success: false}))
	})
}
