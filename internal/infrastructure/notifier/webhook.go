package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rycln/hhraiser/internal/domain"
)

type webhookPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Type  string `json:"type"`
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
	var payload webhookPayload

	if event.Success {
		payload = webhookPayload{
			Title: "Резюме поднято",
			Body:  event.ResumeTitle,
			Type:  "success",
		}
	} else {
		body := event.ResumeTitle
		if event.StatusCode != 0 {
			body = fmt.Sprintf("%s — код ошибки: %d", event.ResumeTitle, event.StatusCode)
		}
		payload = webhookPayload{
			Title: "Ошибка подъёма резюме",
			Body:  body,
			Type:  "failure",
		}
	}

	return w.send(ctx, payload)
}

func (w *Webhook) NotifyApp(ctx context.Context, event domain.AppEvent) error {
	var payload webhookPayload

	switch event.Event {
	case domain.AppEventStarted:
		payload = webhookPayload{
			Title: "hhraiser запущен",
			Body:  "Приложение успешно стартовало",
			Type:  "info",
		}
	case domain.AppEventStopped:
		payload = webhookPayload{
			Title: "hhraiser остановлен",
			Body:  "Приложение завершило работу",
			Type:  "info",
		}
	default:
		payload = webhookPayload{
			Title: "hhraiser",
			Body:  event.Event,
			Type:  "info",
		}
	}

	return w.send(ctx, payload)
}

func (w *Webhook) send(ctx context.Context, payload webhookPayload) error {
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
