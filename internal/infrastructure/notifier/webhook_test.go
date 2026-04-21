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

func TestWebhookNotify(t *testing.T) {
	t.Run("skips when url is empty", func(t *testing.T) {
		w := NewWebhook(http.DefaultClient, "", "secret")
		err := w.Notify(context.Background(), domain.RaiseEvent{})
		require.NoError(t, err)
	})

	t.Run("sends payload and auth header", func(t *testing.T) {
		var gotAuth string
		var gotPayload webhookPayload

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
			defer r.Body.Close()

			err := json.NewDecoder(r.Body).Decode(&gotPayload)
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		w := NewWebhook(server.Client(), server.URL, "secret")
		event := domain.RaiseEvent{
			ResumeTitle: "Go Resume",
			Success:     true,
			Timestamp:   time.Now().UTC(),
		}

		err := w.Notify(context.Background(), event)
		require.NoError(t, err)
		assert.Equal(t, "Bearer secret", gotAuth)
		assert.Equal(t, "raise_success", gotPayload.Event)
		assert.Equal(t, "Go Resume", gotPayload.ResumeTitle)
	})

	t.Run("returns error on non 2xx", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		w := NewWebhook(server.Client(), server.URL, "")
		err := w.Notify(context.Background(), domain.RaiseEvent{Success: false, Timestamp: time.Now()})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "non-2xx")
	})
}

func TestResolveEvent(t *testing.T) {
	t.Run("success resolves to raise_success", func(t *testing.T) {
		assert.Equal(t, "raise_success", resolveEvent(domain.RaiseEvent{Success: true}))
	})

	t.Run("failure resolves to raise_failure", func(t *testing.T) {
		assert.Equal(t, "raise_failure", resolveEvent(domain.RaiseEvent{Success: false}))
	})
}
