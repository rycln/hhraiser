package network

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/rycln/hhraiser/internal/domain"
	"github.com/rycln/hhraiser/internal/infrastructure/httpclient"
)

const (
	loginEndpoint = "/account/login"
)

type Gateway struct {
	client  *httpclient.Client
	baseURL string
}

func NewGateway(client *httpclient.Client, baseURL string) *Gateway {
	return &Gateway{
		client:  client,
		baseURL: baseURL,
	}
}

func (g *Gateway) Login(ctx context.Context, creds *domain.Credentials) (*domain.Session, error) {
	err := g.getAnonymousCookies(ctx)
	if err != nil {
		return nil, err
	}

	xsrf, ok := g.client.GetCookieValue(g.baseURL, "_xsrf")
	if !ok {
		return nil, fmt.Errorf("failed to retrieve XSRF token")
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
	req.Header.Set("X-XSRF-Token", xsrf)

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
		return nil, fmt.Errorf("failed to retrieve XSRF token")
	}
	hhtoken, ok := g.client.GetCookieValue(g.baseURL, "hhtoken")
	if !ok {
		return nil, fmt.Errorf("failed to retrieve hhtoken")
	}

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

	_ = writer.WriteField("_xsrf", xsrf)
	_ = writer.WriteField("backUrl", g.baseURL)
	_ = writer.WriteField("failUrl", loginEndpoint)
	_ = writer.WriteField("remember", "yes")
	_ = writer.WriteField("username", creds.GetPhone())
	_ = writer.WriteField("password", creds.GetPassword())
	_ = writer.WriteField("isBot", "false")

	err := writer.Close()
	if err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}
