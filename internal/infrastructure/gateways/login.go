package gateways

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/rycln/hhraiser/internal/domain"
)

const loginEndpoint = "/account/login"

var (
	errXSRFTokenRetrieval = errors.New("failed to retrieve XSRF token")
	errHHTokenRetrieval   = errors.New("failed to retrieve HHToken")
)

func (g *Gateway) Login(ctx context.Context, creds *domain.Credentials) (*domain.Session, error) {
	slog.DebugContext(ctx, "attempting login")

	err := g.getAnonymousCookies(ctx)
	if err != nil {
		return nil, err
	}

	xsrf, ok := g.client.GetCookieValue(g.baseURL, "_xsrf")
	if !ok {
		return nil, errXSRFTokenRetrieval
	}

	fullURL, err := url.JoinPath(g.baseURL, loginEndpoint)
	if err != nil {
		return nil, err
	}

	body, contentType, err := g.createLoginBody(creds, xsrf)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-xsrftoken", xsrf)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status: %s", resp.Status)
	}

	xsrf, ok = g.client.GetCookieValue(g.baseURL, "_xsrf")
	if !ok {
		return nil, errXSRFTokenRetrieval
	}
	hhtoken, ok := g.client.GetCookieValue(g.baseURL, "hhtoken")
	if !ok {
		return nil, errHHTokenRetrieval
	}

	slog.InfoContext(ctx, "successfully logged in")
	return domain.NewSession(xsrf, hhtoken), nil
}

func (g *Gateway) getAnonymousCookies(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, g.baseURL, nil)
	if err != nil {
		return err
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (g *Gateway) createLoginBody(creds *domain.Credentials, xsrf string) (io.Reader, string, error) {
	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)

	if err := writer.WriteField("_xsrf", xsrf); err != nil {
		return nil, "", err
	}
	if err := writer.WriteField("backUrl", g.baseURL); err != nil {
		return nil, "", err
	}
	if err := writer.WriteField("failUrl", loginEndpoint); err != nil {
		return nil, "", err
	}
	if err := writer.WriteField("remember", "yes"); err != nil {
		return nil, "", err
	}
	if err := writer.WriteField("username", creds.GetPhone()); err != nil {
		return nil, "", err
	}
	if err := writer.WriteField("password", creds.GetPassword()); err != nil {
		return nil, "", err
	}
	if err := writer.WriteField("isBot", "false"); err != nil {
		return nil, "", err
	}

	err := writer.Close()
	if err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}
