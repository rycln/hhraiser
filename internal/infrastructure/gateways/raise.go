package gateways

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/rycln/hhraiser/internal/domain"
)

const raiseEndpoint = "/applicant/resumes/touch"

func (g *Gateway) Raise(ctx context.Context, resume *domain.Resume, session *domain.Session) error {
	slog.DebugContext(ctx, "attempting raise", "url", g.baseURL)

	fullURL, err := url.JoinPath(g.baseURL, raiseEndpoint)
	if err != nil {
		return err
	}

	body, contentType, err := g.createRaiseBody(resume)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-xsrftoken", session.GetXSRF())

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		slog.InfoContext(ctx, "resume", resume.GetTitle(), "raised successfully")
		return nil
	case http.StatusConflict:
		return domain.ErrRaiseTooEarly
	default:
		slog.ErrorContext(ctx, "request failed with status",
			"status_code", resp.StatusCode,
			"status_text", http.StatusText(resp.StatusCode))
		return domain.ErrInvalidSession
	}
}

func (g *Gateway) createRaiseBody(resume *domain.Resume) (io.Reader, string, error) {
	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)

	if err := writer.WriteField("resume", resume.GetID()); err != nil {
		return nil, "", err
	}
	if err := writer.WriteField("undirectable", "true"); err != nil {
		return nil, "", err
	}

	err := writer.Close()
	if err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}
