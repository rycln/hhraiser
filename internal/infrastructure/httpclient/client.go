package httpclient

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/corpix/uarand"
)

type Client struct {
	httpClient *http.Client
	jar        http.CookieJar
}

func New(timeout time.Duration) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
			Jar:     jar,
			Transport: &headerTransport{
				base:      http.DefaultTransport,
				userAgent: uarand.GetRandom(),
			},
		},
		jar: jar,
	}, err
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c *Client) GetCookieValue(targetURL string, cookieName string) (string, bool) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return "", false
	}

	cookies := c.jar.Cookies(u)

	for _, c := range cookies {
		if c.Name == cookieName {
			return c.Value, true
		}
	}

	return "", false
}
