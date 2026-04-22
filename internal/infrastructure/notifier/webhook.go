package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rycln/hhraiser/internal/domain"
)

type raisePayload struct {
	Event       string    `json:"event"`
	ResumeTitle string    `json:"resume_title,omitempty"`
	StatusCode  int       `json:"status_code,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

type appPayload struct {
	Event     string    `json:"event"`
	Timestamp time.Time `json:"timestamp"`
}

type Webhook struct {
	client *http.Client
	url    string
	secret string
}

func NewWebhook(client *http.Client, url, secret string) *Webhook {
	return &Webhook{client: client, url: url, secret: secret}
}

func (w *Webhook) NotifyRaise(ctx context.Context, event domain.RaiseEvent) error {
	payload := raisePayload{
		Event:       resolveRaiseEvent(event),
		ResumeTitle: event.ResumeTitle,
		StatusCode:  event.StatusCode,
		Timestamp:   event.Timestamp,
	}
	return w.send(ctx, payload)
}

func (w *Webhook) NotifyApp(ctx context.Context, event domain.AppEvent) error {
	payload := appPayload{
		Event:     event.Event,
		Timestamp: event.Timestamp,
	}
	return w.send(ctx, payload)
}

func (w *Webhook) send(ctx context.Context, payload any) error {
	if w.url == "" {
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if w.secret != "" {
		req.Header.Set("Authorization", "Bearer "+w.secret)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}

func resolveRaiseEvent(e domain.RaiseEvent) string {
	if e.Success {
		return "raise_success"
	}
	return "raise_failure"
}
