package notifier

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

	t.Run("sends success payload with correct fields", func(t *testing.T) {
		var gotPayload webhookPayload

		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
			defer r.Body.Close()
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
			w.WriteHeader(http.StatusOK)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		err := wh.NotifyRaise(context.Background(), domain.RaiseEvent{
			ResumeTitle: "Go Resume",
			Success:     true,
		})

		require.NoError(t, err)
		assert.Equal(t, "Резюме поднято", gotPayload.Title)
		assert.Equal(t, "Go Resume", gotPayload.Body)
		assert.Equal(t, "success", gotPayload.Type)
	})

	t.Run("sends failure payload without status code", func(t *testing.T) {
		var gotPayload webhookPayload

		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
			w.WriteHeader(http.StatusOK)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		err := wh.NotifyRaise(context.Background(), domain.RaiseEvent{
			ResumeTitle: "Go Resume",
			Success:     false,
		})

		require.NoError(t, err)
		assert.Equal(t, "Ошибка подъёма резюме", gotPayload.Title)
		assert.Equal(t, "Go Resume", gotPayload.Body)
		assert.Equal(t, "failure", gotPayload.Type)
	})

	t.Run("sends failure payload with status code in body", func(t *testing.T) {
		var gotPayload webhookPayload

		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
			w.WriteHeader(http.StatusOK)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		err := wh.NotifyRaise(context.Background(), domain.RaiseEvent{
			ResumeTitle: "Go Resume",
			Success:     false,
			StatusCode:  403,
		})

		require.NoError(t, err)
		assert.Equal(t, "failure", gotPayload.Type)
		assert.Contains(t, gotPayload.Body, "403")
	})

	t.Run("sends auth header when secret is set", func(t *testing.T) {
		var gotAuth string

		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
		})

		wh := NewWebhook(server.Client(), server.URL, "secret")
		err := wh.NotifyRaise(context.Background(), domain.RaiseEvent{Success: true})
		require.NoError(t, err)
		assert.Equal(t, "Bearer secret", gotAuth)
	})

	t.Run("omits auth header when secret is empty", func(t *testing.T) {
		var gotAuth string

		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		err := wh.NotifyRaise(context.Background(), domain.RaiseEvent{Success: true})
		require.NoError(t, err)
		assert.Empty(t, gotAuth)
	})

	t.Run("returns error on non-2xx response", func(t *testing.T) {
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		err := wh.NotifyRaise(context.Background(), domain.RaiseEvent{Success: true})
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
		var gotPayload webhookPayload

		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
			w.WriteHeader(http.StatusOK)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		err := wh.NotifyApp(context.Background(), domain.AppEvent{
			Event: domain.AppEventStarted,
		})

		require.NoError(t, err)
		assert.Equal(t, "hhraiser запущен", gotPayload.Title)
		assert.Equal(t, "info", gotPayload.Type)
	})

	t.Run("sends app_stopped payload", func(t *testing.T) {
		var gotPayload webhookPayload

		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
			w.WriteHeader(http.StatusOK)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		err := wh.NotifyApp(context.Background(), domain.AppEvent{
			Event: domain.AppEventStopped,
		})

		require.NoError(t, err)
		assert.Equal(t, "hhraiser остановлен", gotPayload.Title)
		assert.Equal(t, "info", gotPayload.Type)
	})

	t.Run("returns error on non-2xx response", func(t *testing.T) {
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		wh := NewWebhook(server.Client(), server.URL, "")
		err := wh.NotifyApp(context.Background(), domain.AppEvent{
			Event: domain.AppEventStarted,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "non-2xx")
	})
}
